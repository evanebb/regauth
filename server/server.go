package server

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/evanebb/regauth/auth"
	"github.com/evanebb/regauth/configuration"
	"github.com/evanebb/regauth/resources/database"
	"github.com/evanebb/regauth/resources/database/migrations"
	"github.com/evanebb/regauth/store/postgres"
	"github.com/gorilla/securecookie"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/lestrrat-go/jwx/v2/jwa"
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
	if err != nil {
		return err
	}
	defer db.Close()

	start := time.Now()
	timeout := 30 * time.Second
	logger.InfoContext(ctx, fmt.Sprintf("waiting %s for database", timeout.String()))

	for {
		err := db.Ping(ctx)
		if err == nil {
			break
		}

		if time.Since(start) > timeout {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		time.Sleep(time.Second)
	}

	logger.InfoContext(ctx, "database connection successful")

	// this is a special migration to create the initial admin user from the configuration, so it isn't a plain-text
	// SQL migration and requires the configuration
	migrations.RegisterInitialAdminMigration(conf)
	goose.SetBaseFS(database.Files)
	err = goose.SetDialect("postgres")
	if err != nil {
		return err
	}

	stdlibDb := stdlib.OpenDBFromPool(db)
	defer func() {
		if err := stdlibDb.Close(); err != nil {
			logger.ErrorContext(ctx, "failed to close database connection", slog.Any("error", err))
		}
	}()
	err = goose.Up(stdlibDb, "migrations")
	if err != nil {
		return err
	}

	repoStore := postgres.NewRepositoryStore(db)
	userStore := postgres.NewUserStore(db)
	teamStore := postgres.NewTeamStore(db)
	tokenStore := postgres.NewPersonalAccessTokenStore(db)
	credentialsStore := postgres.NewUserCredentialsStore(db)

	sessionKey := []byte(conf.HTTP.SessionKey)
	if conf.HTTP.SessionKey == "" {
		// generate a random session key, note that sessions are lost on restart and this does not work for horizontal
		// scaling as each instance will need the same session key
		logger.Info("no session key given, generating random key")
		sessionKey = securecookie.GenerateRandomKey(64)
	}

	sessionStore := postgres.NewSessionStore(db, sessionKey)

	authenticator := auth.NewAuthenticator(tokenStore, userStore, conf.Pat.Prefix)
	authorizer := auth.NewAuthorizer(logger, repoStore, teamStore)

	accessTokenConfig, err := buildAccessTokenConfiguration(conf)
	if err != nil {
		return err
	}

	router := baseRouter(logger, repoStore, userStore, teamStore, tokenStore, credentialsStore, sessionStore, authenticator, authorizer, accessTokenConfig, conf.Pat.Prefix)

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

func buildAccessTokenConfiguration(conf *configuration.Configuration) (auth.AccessTokenConfiguration, error) {
	var a auth.AccessTokenConfiguration
	var err error

	a.SigningCert, err = loadCertificate(conf.Token.Certificate)
	if err != nil {
		return auth.AccessTokenConfiguration{}, err
	}

	a.SigningKey, err = loadPrivateKey(conf.Token.Key)
	if err != nil {
		return auth.AccessTokenConfiguration{}, err
	}

	a.VerificationKey = a.SigningKey.(interface {
		Public() crypto.PublicKey
	}).Public()

	a.SigningAlg = jwa.SignatureAlgorithm(conf.Token.Alg)
	a.Issuer = conf.Token.Issuer
	a.Service = conf.Token.Service

	return a, nil
}

func loadCertificate(path string) (*x509.Certificate, error) {
	certFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open certificate file %q: %w", path, err)
	}
	defer func() {
		tempErr := certFile.Close()
		if err == nil && tempErr != nil {
			err = tempErr
		}
	}()

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
	defer func() {
		tempErr := keyFile.Close()
		if err == nil && tempErr != nil {
			err = tempErr
		}
	}()

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
