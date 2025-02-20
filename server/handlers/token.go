package handlers

import (
	"errors"
	"github.com/evanebb/regauth/auth"
	"github.com/evanebb/regauth/pat"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

func GenerateToken(
	l *slog.Logger,
	authenticator auth.Authenticator,
	authorizer auth.Authorizer,
	issuer, service string,
	signer jose.Signer,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			writeRegistryErrorResponse(w, "UNSUPPORTED", "unsupported operation", 405)
			return
		}

		var p *pat.PersonalAccessToken
		var subject string

		username, password, ok := r.BasicAuth()
		if ok {
			sourceIPRaw, _, err := net.SplitHostPort(r.RemoteAddr)
			sourceIP := net.ParseIP(sourceIPRaw)

			lp, u, err := authenticator.Authenticate(r.Context(), username, password, sourceIP)
			if err != nil {
				if errors.Is(err, auth.ErrAuthenticationFailed) {
					l.Info("authentication failed", "error", err)
					writeRegistryErrorResponse(w, "UNAUTHORIZED", "authentication failed", 401)
					return
				}
				l.Error("unknown error occurred during authentication", "error", err)
				writeRegistryErrorResponse(w, "UNKNOWN", "unknown error", 500)
				return
			}

			p = &lp
			subject = u.Username.String()
		}

		requestedAccess := parseScopes(r)
		grantedAccess, err := authorizer.AuthorizeAccess(r.Context(), p, requestedAccess)
		if err != nil {
			l.Error("unknown error occurred during authorization", "error", err)
			writeRegistryErrorResponse(w, "UNKNOWN", "unknown error", 500)
			return
		}

		requestedService := r.URL.Query().Get("service")
		if requestedService != service {
			l.Info("authorization requested for unknown service")
			writeRegistryErrorResponse(w, "DENIED", "authorization requested for unknown service", 403)
			return
		}

		now := time.Now()
		expiry := now.Add(30 * time.Minute)

		claims := jwt.Claims{
			Issuer:    issuer,
			Audience:  jwt.Audience{requestedService},
			Subject:   subject,
			Expiry:    jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		}

		privateClaims := struct {
			Access auth.Access `json:"access"`
		}{
			Access: grantedAccess,
		}

		token, err := jwt.Signed(signer).Claims(claims).Claims(privateClaims).Serialize()
		if err != nil {
			l.Error("unknown error occurred when building token", "error", err)
			writeRegistryErrorResponse(w, "UNKNOWN", "unknown error", 500)
			return
		}

		resp := registryTokenResponse{
			Token:       token,
			AccessToken: token,
			ExpiresIn:   int(expiry.Unix() - now.Unix()),
			IssuedAt:    now.Format(time.RFC3339),
		}

		writeJSONResponse(w, resp)
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
	writeJSONResponse(w, r)
}
