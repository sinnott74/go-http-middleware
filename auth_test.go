package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAuthNoHeader tests that StatusUnauthorized is returned when no Authorization header is set
func TestAuthNoHeader(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	authFunc := func(ctx context.Context, authHeader string) (bool, context.Context) {
		return true, ctx
	}
	auth := Auth(authFunc, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test"))
	}))

	// Act
	auth.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusUnauthorized {
		t.Fatal("StatusUnauthorized 401 expected")
	}
}

// TestAuthFuncNotOk tests that StatusUnauthorized is returned when the supplied authFunc returns false
func TestAuthFuncNotOk(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Add("Authorization", "would_I_lie_to_you")
	w := httptest.NewRecorder()
	authFunc := func(ctx context.Context, authHeader string) (bool, context.Context) {
		return false, ctx
	}
	auth := Auth(authFunc, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test"))
	}))

	// Act
	auth.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusUnauthorized {
		t.Fatal("StatusUnauthorized 401 expected")
	}
}

// TestAuthOk tests that http handler is called when the supplied handleFunc return true
func TestAuthOk(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Add("Authorization", "magic_password")
	w := httptest.NewRecorder()
	authFunc := func(ctx context.Context, authHeader string) (bool, context.Context) {
		return true, ctx
	}
	auth := Auth(authFunc, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test"))
	}))

	// Act
	auth.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatal("StatusOK 200 expected")
	}
}

// TestAuthOk tests that when supplied authFunc returns ok, the context object returned is used with the next request
func TestAuthOkAddToContext(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Add("Authorization", "magic_password")
	w := httptest.NewRecorder()
	authFunc := func(ctx context.Context, authHeader string) (bool, context.Context) {
		userCtx := context.WithValue(ctx, "user", "test@test.com")
		return true, userCtx
	}
	auth := Auth(authFunc, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value("user") != "test@test.com" {
			t.Fatal("Expected user to be set on the request context")
		}
	}))

	// Act
	auth.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatal("StatusOK 200 expected")
	}
}
