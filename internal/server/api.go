package server

import (
	"encoding/json"
	"log/slog"
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

func applyApiHandlers(mux *http.ServeMux, hs *HTTPServer) {
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("GET /store/{key}", hs.getValue)
	apiMux.HandleFunc("POST /store", hs.insertValue)
	apiMux.HandleFunc("DELETE /store/{key}", hs.deleteValue)

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiMux))
}

func (hs *HTTPServer) getValue(rw http.ResponseWriter, request *http.Request) {
	key := request.PathValue("key")

	if key == "" {
		http.Error(rw, "key not set", http.StatusBadRequest)
		return
	}

	value, ok := hs.store.Get(key)
	if !ok {
		http.Error(rw, "not found", http.StatusNotFound)
		return
	}

	resp := KeyStoreGetResponse{
		Key:   key,
		Value: value,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(rw, "internal error", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if _, err := rw.Write(data); err != nil {
		slog.Error("failed to write response", "error", err)
	}
}

func (hs *HTTPServer) insertValue(rw http.ResponseWriter, request *http.Request) {
	var req KeyStoreInsertRequest

	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		http.Error(rw, "body incorrectly formatted", http.StatusBadRequest)
		return
	}

	if err := hs.store.Set(req.Key, req.Value); err != nil {
		http.Error(rw, "failed to set the value for the specified key", http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

func (hs *HTTPServer) deleteValue(rw http.ResponseWriter, request *http.Request) {
	key := request.PathValue("key")

	if key == "" {
		http.Error(rw, "key not set", http.StatusBadRequest)
		return
	}

	if err := hs.store.Delete(key); err != nil {
		http.Error(rw, "failed to delete key", http.StatusBadRequest)
	}

	rw.WriteHeader(http.StatusNoContent)
}
