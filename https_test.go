package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type httpsTestHandler struct {
}

func (httpsTestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestHTTPSRedirect(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/test", nil)
	r.Host = "example.com"
	r.Header.Add("x-forwarded-proto", "http")
	w := httptest.NewRecorder()
	https := HTTPS(&httpsTestHandler{})

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

func TestHTTPSOk(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/test", nil)
	r.Host = "example.com"
	r.Header.Add("x-forwarded-proto", "https")
	w := httptest.NewRecorder()
	https := HTTPS(&httpsTestHandler{})

	// Act
	https.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatal("StatusOK 200 expected")
	}
}
