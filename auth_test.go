package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAuthNoHeader tests that StatusUnauthorized is returned when no Authorization header is set
func TestAuthNoHeader(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	authFunc := func(ctx context.Context, authHeader string) (context.Context, error) {
		return ctx, nil
	}
	auth := Auth(authFunc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Next handler should not have been called")
	}))

	// Act
	auth.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("StatusUnauthorized 401 expected but was %v", w.Code)
	}
}

// TestAuthFuncNotOk tests that StatusUnauthorized is returned when the supplied authFunc returns false
func TestAuthFuncNotOk(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Add("Authorization", "would_I_lie_to_you")
	w := httptest.NewRecorder()
	authFunc := func(ctx context.Context, authHeader string) (context.Context, error) {
		return ctx, errors.New("Not authorised")
	}
	auth := Auth(authFunc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Next handler should not have been called")
	}))

	// Act
	auth.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("StatusUnauthorized 401 expected but was %v", w.Code)
	}
}

// TestAuthOk tests that http handler is called when the supplied handleFunc return true
func TestAuthOk(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Add("Authorization", "magic_password")
	w := httptest.NewRecorder()
	authFunc := func(ctx context.Context, authHeader string) (context.Context, error) {
		return ctx, nil
	}
	auth := Auth(authFunc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test"))
	}))

	// Act
	auth.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatalf("StatusOK 200 expected but was %v", w.Code)
	}
}

// TestAuthOk tests that when supplied authFunc returns ok, the context object returned is used with the next request
func TestAuthOkAddToContext(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Add("Authorization", "magic_password")
	w := httptest.NewRecorder()
	authFunc := func(ctx context.Context, authHeader string) (context.Context, error) {
		userCtx := context.WithValue(ctx, userContextKey, "test@test.com")
		return userCtx, nil
	}
	auth := Auth(authFunc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value(userContextKey) != "test@test.com" {
			t.Fatal("Expected user to be set on the request context")
		}
	}))

	// Act
	auth.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatalf("StatusOK 200 expected but was %v", w.Code)
	}
}

var userContextKey = &contextKey{"user"}
