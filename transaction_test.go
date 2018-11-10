package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestTransactionCommitSuccessfulStatus(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectCommit()

	handler := Transaction(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Act
	handler.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatalf("StatusOK 200 expected but was %v", w.Code)
	}
}

func TestTransactionRollbackNotOkStatus(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectRollback()

	handler := Transaction(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))

	// Act
	handler.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("StatusServiceUnavailable 503 expected but was %v", w.Code)
	}
}

func TestTransactionRollbackPanic(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	db, mock, _ := sqlmock.New()

	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectRollback()

	handler := Transaction(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(errors.New("EVERYTHING IS ON FIRE, DON'T COMMIT"))
	}))

	// Act
	handler.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("StatusInternalServerError 500 expected but was %v", w.Code)
	}
}

func TestTransactionCommitReadTransactionInHandler(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	db, mock, _ := sqlmock.New()

	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectRollback()

	handler := Transaction(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tx := GetTransaction(r.Context())
		if tx == nil {
			panic(errors.New("GetTransaction should return a *sql.Tx"))
		}

	}))

	// Act
	handler.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatalf("StatusOK 200 expected but was %v", w.Code)
	}
}

func TestTransactionCommitWriteInHandler(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	db, mock, _ := sqlmock.New()

	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectCommit()

	handler := Transaction(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("It worked"))
	}))

	// Act
	handler.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatalf("StatusOK 200 expected but was %v", w.Code)
	}

	if s := string(w.Body.Bytes()); s != "It worked" {
		t.Fatalf("\"It worked\" response body expected but was %v", s)
	}
}

func TestTransactionRollbackErrorDuringCommit(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	db, mock, _ := sqlmock.New()

	defer db.Close()
	mock.ExpectBegin()
	// mock.ExpectCommit()

	handler := Transaction(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Act
	handler.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("StatusInternalServerError 500 expected but was %v", w.Code)
	}
}

func TestTransactionErrorDuringTxBegin(t *testing.T) {

	// Arrange
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	db, _, _ := sqlmock.New()

	defer db.Close()
	// mock.ExpectBegin()

	handler := Transaction(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Act
	handler.ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("StatusInternalServerError 500 expected but was %v", w.Code)
	}
}
