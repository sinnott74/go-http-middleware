package middleware

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"hash"
	"net/http"
	"net/http/httptest"
	"testing"
)

var responseText = "Test"

type etagTestHandler struct {
}

func (etagTestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(responseText))
}

type etagTestHandlerNotOK struct {
}

func (etagTestHandlerNotOK) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func TestDefaultETag(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	etag := DefaultEtag(&etagTestHandler{})

	expectedHash := calculateHash(md5.New(), responseText)

	// Act
	etag.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatalf("StatusOK 200 expected - %d", w.Code)
	}
	if w.Header().Get("ETag") != expectedHash {
		t.Fatalf("%s expected - %s", expectedHash, w.Header().Get("ETag"))
	}
}

func TestDefaultETagMatch(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/test", nil)
	r.Header.Add("If-None-Match", "W/\"4-DLxmEfVUC9CAmjiNyVphWw==\"")
	w := httptest.NewRecorder()
	etag := DefaultEtag(&etagTestHandler{})

	// Act
	etag.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusNotModified {
		t.Fatalf("StatusNotModified 304 expected - %d", w.Code)
	}
}

func TestDefaultETagErrorResponse(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	etag := DefaultEtag(&etagTestHandlerNotOK{})

	// Act
	etag.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("StatusInternalServerError 500 expected - %d", w.Code)
	}

	if w.Header().Get("ETag") != "" {
		t.Fatalf("expected no Etag header but got - %s", w.Header().Get("ETag"))
	}
}

func TestRequestTwice(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	etag := DefaultEtag(&etagTestHandler{})

	// Act
	etag.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatalf("StatusOK 200 expected - %d", w.Code)
	}

	etagHeader := w.Header().Get("ETag")
	r, _ = http.NewRequest("GET", "/test", nil)
	r.Header.Add("If-None-Match", etagHeader)
	w = httptest.NewRecorder()

	etag.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusNotModified {
		t.Fatalf("StatusNotModified 304 expected - %d", w.Code)
	}
}

func TestEtag(t *testing.T) {
	r, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	etag := Etag(sha1.New, &etagTestHandler{})
	expectedHash := calculateHash(sha1.New(), responseText)

	// Act
	etag.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatalf("StatusOK 200 expected - %d", w.Code)
	}
	if w.Header().Get("ETag") != expectedHash {
		t.Fatalf("%s expected - %s", expectedHash, w.Header().Get("ETag"))
	}
}

func calculateHash(hash hash.Hash, text string) string {
	hash.Write([]byte(text))
	base64Hash := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	len := len(text)
	return fmt.Sprintf("W/\"%v-%v\"", len, base64Hash)
}
