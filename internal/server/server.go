package server

import (
	"context"
	"net/http"
	"time"

	"github.com/LjungErik/zetra-kv-store/internal/proxy"
	raft_internal "github.com/LjungErik/zetra-kv-store/internal/raft"
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
	Proxy        proxy.Proxy
	Raft         *raft_internal.Raft
	ReadTimeout  *time.Duration
	WriteTimeout *time.Duration
}

type HTTPServer struct {
	store  store.KVStore
	server *http.Server
	raft   *raft_internal.Raft
	proxy  proxy.Proxy
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
		proxy:  cfg.Proxy,
		raft:   cfg.Raft,
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
