package middleware

import (
	"net/http"
)

// HTTPS is middleware which redirects the user to https if the x-forward-proto header is set to http
func HTTPS(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		proto := r.Header.Get("x-forwarded-proto")
		if proto == "http" {
			http.Redirect(w, r, "https://"+r.Host+r.URL.Path, http.StatusPermanentRedirect)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
