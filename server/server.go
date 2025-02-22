package server

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/evanebb/regauth/auth"
	"github.com/evanebb/regauth/configuration"
	"github.com/evanebb/regauth/resources/database"
	"github.com/evanebb/regauth/resources/templates"
	"github.com/evanebb/regauth/session"
	"github.com/evanebb/regauth/store/postgres"
	"github.com/evanebb/regauth/template"
	"github.com/go-chi/chi/v5"
	"github.com/go-jose/go-jose/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func Run(ctx context.Context, conf *configuration.Configuration) error {
	// Initialize all the dependencies
	logger, err := buildLogger(conf)
	if err != nil {
		return err
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", conf.Database.User, conf.Database.Password, conf.Database.Host, conf.Database.Port, conf.Database.Name)
	db, err := pgxpool.New(ctx, connString)
	defer db.Close()
	if err != nil {
		return err
	}

	goose.SetBaseFS(database.Files)
	err = goose.SetDialect("postgres")
	if err != nil {
		return err
	}

	stdlibDb := stdlib.OpenDBFromPool(db)
	defer stdlibDb.Close()
	err = goose.Up(stdlibDb, "migrations")
	if err != nil {
		return err
	}

	repoStore := postgres.NewRepositoryStore(db)
	patStore := postgres.NewPersonalAccessTokenStore(db)
	userStore := postgres.NewUserStore(db)
	authUserStore := postgres.NewAuthUserStore(db)
	sessionStore := session.NewPgxStore(db, []byte(conf.HTTP.SessionKey))

	authenticator := auth.NewAuthenticator(patStore, userStore)
	authorizer := auth.NewAuthorizer(logger, repoStore)

	templater := template.NewTemplater(logger, templates.Files, sessionStore)

	certificate, err := loadCertificate(conf.Token.Certificate)
	if err != nil {
		return err
	}

	privateKey, err := loadPrivateKey(conf.Token.Key)
	if err != nil {
		return err
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.SignatureAlgorithm(conf.Token.Alg), Key: privateKey},
		(&jose.SignerOptions{}).WithHeader("x5c", []string{base64.StdEncoding.EncodeToString(certificate.Raw)}),
	)
	if err != nil {
		return fmt.Errorf("could not create signer: %w", err)
	}

	router := chi.NewRouter()
	addRoutes(router, logger, sessionStore, templater, repoStore, userStore, patStore, authUserStore, authenticator, authorizer, conf.Token.Issuer, conf.Token.Service, signer, conf.Registry.Host)

	server := &http.Server{
		Addr:    conf.HTTP.Addr,
		Handler: router,
	}

	go func() {
		if conf.HTTP.Certificate != "" && conf.HTTP.Key != "" {
			logger.Info(fmt.Sprintf("starting https server on %s", server.Addr))
			if err := server.ListenAndServeTLS(conf.HTTP.Certificate, conf.HTTP.Key); !errors.Is(err, http.ErrServerClosed) {
				logger.Error("error listening and serving", "error", err)
			}
		} else {
			logger.Info(fmt.Sprintf("starting http server on %s", server.Addr))
			if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				logger.Error("error listening and serving", "error", err)
			}
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("shutting down http server")
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("error shutting down http server: %w", err)
	}

	return nil
}

var logLevelMap = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func buildLogger(conf *configuration.Configuration) (*slog.Logger, error) {
	level, ok := logLevelMap[conf.Log.Level]
	if !ok {
		return nil, fmt.Errorf("invalid log level %q given", conf.Log.Level)
	}

	logHandlerOptions := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	switch conf.Log.Formatter {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, logHandlerOptions)
	case "text":
		handler = slog.NewTextHandler(os.Stderr, logHandlerOptions)
	}

	return slog.New(handler), nil
}

func loadCertificate(path string) (*x509.Certificate, error) {
	certFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open certificate file %q: %w", path, err)
	}
	defer certFile.Close()

	data, err := io.ReadAll(certFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file %q: %w", path, err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("could not decode certificate data as PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("could not load certificate from input: %w", err)
	}

	return cert, nil
}

func loadPrivateKey(path string) (crypto.PrivateKey, error) {
	keyFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open private key file %q: %w", path, err)
	}
	defer keyFile.Close()

	data, err := io.ReadAll(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file %q: %w", path, err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("could not decode private key data as PEM")
	}

	input := block.Bytes

	var priv crypto.PrivateKey
	priv, err = x509.ParsePKCS1PrivateKey(input)
	if err == nil {
		return priv, err
	}

	priv, err = x509.ParsePKCS8PrivateKey(input)
	if err == nil {
		return priv, err
	}

	priv, err = x509.ParseECPrivateKey(input)
	if err == nil {
		return priv, err
	}

	return nil, errors.New("could not load private key from input, no valid key found")
}
