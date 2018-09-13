package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHTTPSRedirect tests that when the x-forwarded-proto header is set to http
// the request is redirected to the HTTPS version of the url
func TestHTTPSRedirect(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/test", nil)
	r.Host = "example.com"
	r.Header.Add("x-forwarded-proto", "http")
	w := httptest.NewRecorder()
	https := HTTPS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Act
	https.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusPermanentRedirect {
		t.Fatalf("StatusPermanentRedirect 308 expected - %d", w.Code)
	}
	if w.Header().Get("Location") != "https://example.com/test" {
		t.Fatalf("Expect Location header to point at https url - %s", w.Header().Get("Location"))
	}
}

// TestHTTPSRedirect tests that when the x-forwarded-proto header is set to https
// the request continues to the next chained http handler
func TestHTTPSOk(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/test", nil)
	r.Host = "example.com"
	r.Header.Add("x-forwarded-proto", "https")
	w := httptest.NewRecorder()
	https := HTTPS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Act
	https.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatal("StatusOK 200 expected")
	}
}
