package server

import (
	"encoding/json"
	"log"
	"net/http"
)

type KeyStoreGetResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func applyApiHandlers(mux *http.ServeMux, hs *HTTPServer) {
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("GET /store/{key}", hs.getValue)

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
		log.Printf("failed to write response: %v", err)
	}
}
