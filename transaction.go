package middleware

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
)

// Transaction middleware starts a database transaction and adds it to the request context.
// The transaction will rollback if a non successful http status code is writen to the request, if a panic occurs during the handler
func Transaction(db *sql.DB) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ctx := r.Context()
			sw := &statusWriter{rw: w, buf: bytes.NewBuffer(nil)}

			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				sw.WriteHeader(http.StatusInternalServerError)
				sw.Finish()
				return
			}

			defer func() {
				if rec := recover(); rec != nil {
					tx.Rollback()
					sw.WriteHeader(http.StatusInternalServerError)
					sw.Finish()
					return
				}

				if !isHTTPStatusOk(sw.status) {
					tx.Rollback()
					sw.Finish()
					return
				}

				err := tx.Commit()
				if err != nil {
					tx.Rollback()
					sw.WriteHeader(http.StatusInternalServerError)
					sw.Finish()
					return
				}

				sw.Finish()
			}()

			txCtx := setTransaction(ctx, tx)
			next.ServeHTTP(sw, r.WithContext(txCtx))
		})
	}
}

// tx context key
var txKey = &contextKey{"Tx"}

// setTransaction creates a child context with a transaction value
func setTransaction(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// GetTransaction gets the transation stored in the context
func GetTransaction(ctx context.Context) *sql.Tx {
	return ctx.Value(txKey).(*sql.Tx)
}

// statusWriter wraps ResponseWriter to intercept the written http status
type statusWriter struct {
	rw     http.ResponseWriter
	status int
	buf    *bytes.Buffer
}

// WriteHeader wraps setting the status
func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
}

// Write wraps ResponseWriter's Write and sets the http status if it hasn't already been set
func (sw *statusWriter) Write(b []byte) (int, error) {
	if sw.status == 0 {
		sw.status = http.StatusOK
	}
	return sw.buf.Write(b)
}

// Header wraps ResponseWriter's Header
func (sw *statusWriter) Header() http.Header {
	return sw.rw.Header()
}

func (sw *statusWriter) Finish() error {
	if sw.status != 0 {
		sw.rw.WriteHeader(sw.status)
	}
	_, err := sw.rw.Write(sw.buf.Bytes())
	return err
}
