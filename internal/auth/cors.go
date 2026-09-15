package auth

import (
	"net/http"
	"net/url"
)

// CORS allows credentialed browser API requests only from frontendURL's origin.
// OAuth itself is a browser navigation; this middleware is needed for the
// frontend's subsequent GraphQL requests that send the session cookie.
func (h *OAuthHandler) CORS(next http.Handler) http.Handler {
	allowedOrigin := originFromURL(h.frontendURL)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Add("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// FrontendOrigin returns the scheme://host this handler treats as the
// single allowed origin - the same value CORS() checks incoming HTTP
// requests against. Used by cmd/server/main.go to configure the same
// allow-list for the WebSocket (subscriptions) transport's origin check,
// which is a separate mechanism from CORS (browsers don't send CORS
// preflight requests for WebSocket connections).
func (h *OAuthHandler) FrontendOrigin() string {
	return originFromURL(h.frontendURL)
}

func originFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}

	return parsed.Scheme + "://" + parsed.Host
}
