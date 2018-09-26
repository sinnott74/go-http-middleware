package middleware

import (
	"context"
	"net/http"
)

// AuthFunc defines the user supplied function to implement Authorisation
// It is given the current request context and the Authorization header value
// and returns whether on not the request is authenticate
// and the context object to use with further chained http handlers
type AuthFunc func(context.Context, string) (bool, context.Context)

// Auth middleware is responsible handling request authentication
// The authentication is handled by the supplied AuthFunc
func Auth(authFunc AuthFunc, next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			// missing header
			w.WriteHeader(http.StatusUnauthorized)
			// w.Write(errors.New("unauthorized: no authentication provided").Error())
			return
		}
		ok, ctx := authFunc(r.Context(), auth)
		if !ok {
			// unauthorised
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
