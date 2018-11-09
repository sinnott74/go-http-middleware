package middleware

import "net/http"

// Middleware is defined as a function which takes a http handler
// and returns a new http handler which wraps the input with extra functionality
type Middleware func(next http.Handler) http.Handler
