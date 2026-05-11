package server

import (
	"context"
	"net/http"
	"time"

	"github.com/LjungErik/zetra-kv-store/internal/store"
)

const (
	defaultReadTimeout  = 5 * time.Second
	defaultWriteTimeout = 5 * time.Second
)

type Server interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

type Config struct {
	Addr         string
	Store        store.KVStore
	ReadTimeout  *time.Duration
	WriteTimeout *time.Duration
}

type HTTPServer struct {
	store  store.KVStore
	server *http.Server
}

var _ Server = (*HTTPServer)(nil)

func NewServer(cfg Config) *HTTPServer {
	mux := http.NewServeMux()

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      mux,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
	}

	if cfg.ReadTimeout != nil {
		server.ReadTimeout = *cfg.ReadTimeout
	}

	if cfg.WriteTimeout != nil {
		server.WriteTimeout = *cfg.WriteTimeout
	}

	hs := &HTTPServer{
		store:  cfg.Store,
		server: server,
	}

	applyApiHandlers(mux, hs)

	return hs
}

func (hs *HTTPServer) ListenAndServe() error {
	return hs.server.ListenAndServe()
}

func (hs *HTTPServer) Shutdown(ctx context.Context) error {
	return hs.server.Shutdown(ctx)
}
