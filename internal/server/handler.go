package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type HTTPError struct {
	Status  int    `json:"-"`
	Message string `json:"error"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

func NewHTTPError(status int, msg string) *HTTPError {
	return &HTTPError{Status: status, Message: msg}
}

type Validator interface {
	Validate() error
}

type HandlerFunc[T any] func(w http.ResponseWriter, r *http.Request, body T) error

type HandlerFuncNoBody func(w http.ResponseWriter, r *http.Request) error

func Handle[T Validator](fn HandlerFunc[T]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body T

		if r.Body == nil || r.ContentLength == 0 {
			writeError(w, NewHTTPError(http.StatusBadRequest, "request body required"))
			return
		}
		defer r.Body.Close()

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, NewHTTPError(http.StatusBadRequest, "invalid JSON: "+err.Error()))
			return
		}

		if err := body.Validate(); err != nil {
			writeError(w, NewHTTPError(http.StatusBadRequest, err.Error()))
			return
		}

		if err := fn(w, r, body); err != nil {
			handleError(w, err)
		}
	})
}

func HandleNoBody(fn HandlerFuncNoBody) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			handleError(w, err)
		}
	})
}

func handleError(w http.ResponseWriter, err error) {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		writeError(w, httpErr)
	} else {
		slog.Error("unhandled error", "error", err)
		writeError(w, NewHTTPError(http.StatusInternalServerError, "internal server error"))
	}
}

func writeError(w http.ResponseWriter, err *HTTPError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(err)
}

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
