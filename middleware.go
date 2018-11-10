package middleware

import "net/http"

// Middleware is defined as a function which takes a http handler
// and returns a new http handler which wraps the input with extra functionality
type Middleware func(next http.Handler) http.Handler

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation. This technique
// for defining context keys was copied from Go 1.7's new use of context in net/http.
type contextKey struct {
	name string
}

func (c *contextKey) String() string {
	return "middleware context key " + c.name
}

// isHTTPStatusOk checks if the given http status is in the 2xx range
func isHTTPStatusOk(status int) bool {
	return status >= 200 && status < 300
}
