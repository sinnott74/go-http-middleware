package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// TestJWTNoHeader tests that StatusUnauthorized is returned when no Authorization header is set
func TestJWTNoHeader(t *testing.T) {

	// Arrange
	secret := []byte("SECRET_SSSHHHHHHH")
	jwtOptions := JWTOptions{secret: secret}
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	auth := JWT(jwtOptions)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Next handler should not have been called")
	}))

	// Act
	auth.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("StatusUnauthorized 401 expected but was %v", w.Code)
	}
}

// TestJWTBadToken tests that StatusUnauthorized is returned when a bad JWT Authorization header is set
func TestJWTBadToken(t *testing.T) {

	// Arrange
	secret := []byte("SECRET_SSSHHHHHHH")
	jwtOptions := JWTOptions{secret: secret}
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Add("Authorization", "would_I_lie_to_you")
	w := httptest.NewRecorder()
	auth := JWT(jwtOptions)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Next handler should not have been called")
	}))

	// Act
	auth.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("StatusUnauthorized 401 expected but was %v", w.Code)
	}
}

// TestJWTValidToken tests that StatusOK is returned when a valid JWT Authorization header is set
func TestJWTValidToken(t *testing.T) {

	// Arrange
	secret := []byte("SECRET_SSSHHHHHHH")
	jwtOptions := JWTOptions{secret: secret}
	token := createValidJWT(t, secret)
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Add("Authorization", token)
	w := httptest.NewRecorder()
	auth := JWT(jwtOptions)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Act
	auth.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatalf("StatusOK 200 expected but was %v", w.Code)
	}
}

// TestJWTExpiredToken tests that en expired token is not valid
func TestJWTExpiredToken(t *testing.T) {

	// Arrange
	secret := []byte("SECRET_SSSHHHHHHH")
	jwtOptions := JWTOptions{secret: secret}
	r, _ := http.NewRequest("GET", "/", nil)
	token := createJWTWithExpiration(t, secret, time.Now().Add(-time.Minute*1))
	r.Header.Add("Authorization", token)
	w := httptest.NewRecorder()
	auth := JWT(jwtOptions)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Next handler should not have been called as the token is invalid")
	}))

	// Act
	auth.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("StatusUnauthorized 401 expected but was %v", w.Code)
	}
}

func TestJWTValidTokenWithUserSuppliedFunc(t *testing.T) {

	// Arrange
	secret := []byte("SECRET_SSSHHHHHHH")
	jwtOptions := JWTOptions{secret: secret, authFunc: func(ctx context.Context, claims jwt.MapClaims) (context.Context, error) {
		userCtx := context.WithValue(ctx, userContextKey, "test@test.com")
		return userCtx, nil
	}}
	token := createValidJWT(t, secret)
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Add("Authorization", token)
	w := httptest.NewRecorder()
	auth := JWT(jwtOptions)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value(userContextKey) != "test@test.com" {
			t.Fatal("Expected user to be set on the request context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Act
	auth.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatalf("StatusOK 200 expected but was %v", w.Code)
	}
}

func TestJWTValidTokenWithUserSuppliedFuncThatReturnsError(t *testing.T) {

	// Arrange
	secret := []byte("SECRET_SSSHHHHHHH")
	jwtOptions := JWTOptions{secret: secret, authFunc: func(ctx context.Context, claims jwt.MapClaims) (context.Context, error) {
		return ctx, errors.New("User supplied func says claims aren't good")
	}}
	token := createValidJWT(t, secret)
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Add("Authorization", token)
	w := httptest.NewRecorder()
	auth := JWT(jwtOptions)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Next handler should not have been called are user supplied func returns an error")
	}))

	// Act
	auth.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("StatusUnauthorized 401 expected but was %v", w.Code)
	}
}

func createValidJWT(t *testing.T, secret []byte) string {
	claims := jwt.MapClaims{}
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	return tokenString
}

func createJWTWithExpiration(t *testing.T, secret []byte, expiration time.Time) string {
	claims := jwt.MapClaims{}
	claims["exp"] = expiration.Unix()
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	return tokenString
}