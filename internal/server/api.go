package server

import (
	"net/http"

	"github.com/LjungErik/zetra-kv-store/internal/server/models"
)

func applyApiHandlers(mux *http.ServeMux, hs *HTTPServer) {
	apiMux := http.NewServeMux()
	apiMux.Handle("GET /store/{key}", HandleNoBody(hs.getValue))

	apiMux.Handle("POST /store", hs.leaderProxy(Handle(hs.insertValue)))
	apiMux.Handle("DELETE /store", hs.leaderProxy(Handle(hs.deleteValue)))

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

	resp := models.KeyStoreGetResponse{
		Key:   key,
		Value: value,
	}

	return WriteJSON(rw, http.StatusOK, resp)
}

func (hs *HTTPServer) insertValue(rw http.ResponseWriter, r *http.Request, req models.KeyStoreInsertRequest) error {
	if err := hs.store.Set(req.Key, req.Value); err != nil {
		return NewHTTPError(http.StatusBadRequest, "failed to set the given value")
	}

	rw.WriteHeader(http.StatusNoContent)

	return nil
}

func (hs *HTTPServer) deleteValue(rw http.ResponseWriter, request *http.Request, req models.KeyStoreDeleteRequest) error {
	if err := hs.store.Delete(req.Key); err != nil {
		return NewHTTPError(http.StatusBadRequest, "failed to delete key")
	}

	rw.WriteHeader(http.StatusNoContent)

	return nil
}
