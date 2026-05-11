package server

import (
	"errors"
	"net/http"
)

type KeyStoreGetResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KeyStoreInsertRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (r KeyStoreInsertRequest) Validate() error {
	if r.Key == "" {
		return errors.New("key not set")
	}

	return nil
}

func applyApiHandlers(mux *http.ServeMux, hs *HTTPServer) {
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("GET /store/{key}", HandleNoBody(hs.getValue))
	apiMux.HandleFunc("POST /store", Handle(hs.insertValue))
	apiMux.HandleFunc("DELETE /store/{key}", HandleNoBody(hs.deleteValue))

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiMux))
}

func (hs *HTTPServer) getValue(rw http.ResponseWriter, r *http.Request) error {
	key := r.PathValue("key")

	if key == "" {
		return NewHTTPError(http.StatusBadRequest, "key not set")
	}

	value, ok := hs.store.Get(key)
	if !ok {
		return NewHTTPError(http.StatusNotFound, "not found")
	}

	resp := KeyStoreGetResponse{
		Key:   key,
		Value: value,
	}

	return WriteJSON(rw, http.StatusOK, resp)
}

func (hs *HTTPServer) insertValue(rw http.ResponseWriter, r *http.Request, req KeyStoreInsertRequest) error {
	if err := hs.store.Set(req.Key, req.Value); err != nil {
		return NewHTTPError(http.StatusBadRequest, "failed to set the given value")
	}

	rw.WriteHeader(http.StatusNoContent)

	return nil
}

func (hs *HTTPServer) deleteValue(rw http.ResponseWriter, request *http.Request) error {
	key := request.PathValue("key")

	if key == "" {
		return NewHTTPError(http.StatusBadRequest, "key not set")
	}

	if err := hs.store.Delete(key); err != nil {
		return NewHTTPError(http.StatusBadRequest, "failed to delete key")
	}

	rw.WriteHeader(http.StatusNoContent)

	return nil
}
