package handlers

import (
	"encoding/base64"
	"errors"
	"github.com/evanebb/regauth/auth"
	"github.com/evanebb/regauth/server/response"
	"github.com/evanebb/regauth/token"
	"github.com/evanebb/regauth/user"
	"github.com/lestrrat-go/jwx/v2/cert"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

// GenerateRegistryToken generates an access token that can be used to authenticate against the registry.
func GenerateRegistryToken(
	l *slog.Logger,
	authenticator auth.Authenticator,
	authorizer auth.Authorizer,
	tokenConfig auth.AccessTokenConfiguration,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			writeRegistryErrorResponse(w, "UNSUPPORTED", "unsupported operation", http.StatusMethodNotAllowed)
			return
		}

		var p *token.PersonalAccessToken
		var u *user.User
		var subject string

		username, password, ok := r.BasicAuth()
		if ok {
			sourceIPRaw, _, err := net.SplitHostPort(r.RemoteAddr)
			sourceIP := net.ParseIP(sourceIPRaw)

			lp, lu, err := authenticator.Authenticate(r.Context(), username, password, sourceIP)
			if err != nil {
				if errors.Is(err, auth.ErrAuthenticationFailed) {
					l.Info("authentication failed", "error", err)
					writeRegistryErrorResponse(w, "UNAUTHORIZED", "authentication failed", http.StatusUnauthorized)
					return
				}
				l.Error("unknown error occurred during authentication", "error", err)
				writeRegistryErrorResponse(w, "UNKNOWN", "unknown error", http.StatusInternalServerError)
				return
			}

			p = &lp
			u = &lu
			subject = string(u.Username)
		}

		requestedAccess := parseScopes(r)
		grantedAccess, err := authorizer.AuthorizeAccess(r.Context(), u, p, requestedAccess)
		if err != nil {
			l.Error("unknown error occurred during authorization", "error", err)
			writeRegistryErrorResponse(w, "UNKNOWN", "unknown error", http.StatusInternalServerError)
			return
		}

		requestedService := r.URL.Query().Get("service")
		if requestedService != tokenConfig.Service {
			l.Info("authorization requested for unknown service")
			writeRegistryErrorResponse(w, "DENIED", "authorization requested for unknown service", http.StatusForbidden)
			return
		}

		now := time.Now()
		expiry := now.Add(30 * time.Minute)

		accessToken, err := jwt.NewBuilder().
			Issuer(tokenConfig.Issuer).
			Audience([]string{tokenConfig.Service}).
			Subject(subject).
			Expiration(expiry).
			IssuedAt(now).
			NotBefore(now).
			Claim("access", grantedAccess).
			Build()

		if err != nil {
			l.Error("unknown error occurred when building token", "error", err)
			writeRegistryErrorResponse(w, "UNKNOWN", "unknown error", http.StatusInternalServerError)
			return
		}

		var certChain cert.Chain
		_ = certChain.AddString(base64.StdEncoding.EncodeToString(tokenConfig.SigningCert.Raw))

		headers := jws.NewHeaders()
		if err = headers.Set(jws.X509CertChainKey, &certChain); err != nil {
			l.Error("unknown error occurred when setting x5c header", "error", err)
			writeRegistryErrorResponse(w, "UNKNOWN", "unknown error", http.StatusInternalServerError)
			return
		}

		signed, err := jwt.Sign(accessToken, jwt.WithKey(tokenConfig.SigningAlg, tokenConfig.SigningKey, jws.WithProtectedHeaders(headers)))
		if err != nil {
			l.Error("unknown error occurred when signing token", "error", err)
			writeRegistryErrorResponse(w, "UNKNOWN", "unknown error", http.StatusInternalServerError)
			return
		}

		signedString := string(signed)
		resp := registryTokenResponse{
			Token:       signedString,
			AccessToken: signedString,
			ExpiresIn:   int(expiry.Unix() - now.Unix()),
			IssuedAt:    now.Format(time.RFC3339),
		}

		response.WriteJSONResponse(w, http.StatusOK, resp)
	})
}

func parseScopes(r *http.Request) auth.Access {
	var requestedAccess auth.Access
	scopes := r.URL.Query()["scope"]
	for _, scope := range scopes {
		var ra auth.ResourceActions

		parts := strings.Split(scope, ":")
		// FIXME: add regex validation to parts
		switch len(parts) {
		case 3:
			ra = auth.ResourceActions{
				Type:    parts[0],
				Name:    parts[1],
				Actions: strings.Split(parts[2], ","),
			}
		case 4:
			ra = auth.ResourceActions{
				Type:    parts[0],
				Name:    parts[1] + ":" + parts[2],
				Actions: strings.Split(parts[3], ","),
			}
		default:
			// Invalid scope, just skip it
			continue
		}

		requestedAccess = append(requestedAccess, ra)
	}

	return requestedAccess
}

type registryTokenResponse struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	IssuedAt    string `json:"issued_at"`
}

type registryError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type registryErrorResponse []registryError

func writeRegistryErrorResponse(w http.ResponseWriter, code, message string, responseCode int) {
	r := registryErrorResponse{{code, message}}
	w.WriteHeader(responseCode)
	response.WriteJSONResponse(w, http.StatusOK, r)
}
