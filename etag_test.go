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

// TestDefaultETagOk tests that the expect MD5 ETag is returned
func TestDefaultETagOk(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	responseText := "Test"
	expectedHash := calculateHash(md5.New(), responseText)
	etag := DefaultEtag(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseText))
	}))

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

// TestDefaultETagMatch tests StatusNotModified is returned when the If-None-Match header matches the ETag
func TestDefaultETagMatch(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/test", nil)
	r.Header.Add("If-None-Match", "W/\"4-DLxmEfVUC9CAmjiNyVphWw==\"")
	w := httptest.NewRecorder()
	etag := DefaultEtag(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test"))
	}))

	// Act
	etag.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusNotModified {
		t.Fatalf("StatusNotModified 304 expected - %d", w.Code)
	}
}

// TestDefaultETagErrorResponse tests that no ETag is returned when the wrapped handler response isn't ok
func TestDefaultETagErrorResponse(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	etag := DefaultEtag(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Test"))
	}))

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

// TestRequestTwice tests that using the ETag returned by one request will result in a StatusNotModified when requesting again
func TestRequestTwice(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	etag := DefaultEtag(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test"))
	}))

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

// TestEtag tests that a different Hash struct can be supplied and is used
func TestEtag(t *testing.T) {
	r, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	responseText := "Test"
	etag := Etag(sha1.New(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(responseText))
	}))
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

// calculateHash calculates the expected Etag
func calculateHash(hash hash.Hash, text string) string {
	hash.Write([]byte(text))
	base64Hash := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	len := len(text)
	return fmt.Sprintf("W/\"%v-%v\"", len, base64Hash)
}
