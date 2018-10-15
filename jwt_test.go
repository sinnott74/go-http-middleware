package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// TestJWTNoHeader tests that StatusUnauthorized is returned when no Authorization header is set
func TestJWTNoHeader(t *testing.T) {

	// Arrange
	secret := []byte("SECRET_SSSHHHHHHH")
	jwtOptions := JWTOptions{Secret: secret}
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
	jwtOptions := JWTOptions{Secret: secret}
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
	jwtOptions := JWTOptions{Secret: secret}
	token := createValidJWT(t, secret, "JWT")
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
	jwtOptions := JWTOptions{Secret: secret}
	r, _ := http.NewRequest("GET", "/", nil)
	token := createJWTWithExpiration(t, secret, "JWT", time.Now().Add(-time.Minute*1))
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
	jwtOptions := JWTOptions{Secret: secret, AuthFunc: func(ctx context.Context, claims jwt.MapClaims) (context.Context, error) {
		userCtx := context.WithValue(ctx, userContextKey, "test@test.com")
		return userCtx, nil
	}}
	token := createValidJWT(t, secret, "JWT")
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
	jwtOptions := JWTOptions{Secret: secret, AuthFunc: func(ctx context.Context, claims jwt.MapClaims) (context.Context, error) {
		return ctx, errors.New("User supplied func says claims aren't good")
	}}
	token := createValidJWT(t, secret, "JWT")
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

func TestJWTValidTokenWithCustomTokenExtractor(t *testing.T) {

	// Arrange
	secret := []byte("SECRET_SSSHHHHHHH")
	jwtOptions := JWTOptions{
		Secret: secret,
		AuthFunc: func(ctx context.Context, claims jwt.MapClaims) (context.Context, error) {
			return ctx, errors.New("User supplied func says claims aren't good")
		},
		Extractor: func(authHeaderValue string) (string, error) {
			authHeaderParts := strings.Split(authHeaderValue, " ")
			if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
				return "", errors.New("Authorization header format must be bearer {token}")
			}
			return authHeaderParts[1], nil
		},
	}
	token := createValidJWT(t, secret, "bearer")
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

func TestJWTValidTokenWithCustomTokenExtractorError(t *testing.T) {

	// Arrange
	secret := []byte("SECRET_SSSHHHHHHH")
	jwtOptions := JWTOptions{
		Secret: secret,
		AuthFunc: func(ctx context.Context, claims jwt.MapClaims) (context.Context, error) {
			return ctx, errors.New("User supplied func says claims aren't good")
		},
		Extractor: func(authHeaderValue string) (string, error) {
			return "", errors.New("Some error getting token from header value")
		},
	}
	token := createValidJWT(t, secret, "bearer")
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

func createValidJWT(t *testing.T, secret []byte, scheme string) string {
	claims := jwt.MapClaims{}
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	return scheme + " " + tokenString
}

func createJWTWithExpiration(t *testing.T, secret []byte, scheme string, expiration time.Time) string {
	claims := jwt.MapClaims{}
	claims["exp"] = expiration.Unix()
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	return scheme + " " + tokenString
}
