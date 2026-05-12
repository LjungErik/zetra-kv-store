package server

import (
	"net/http"

	"github.com/LjungErik/zetra-kv-store/internal/proxy"
	raft_internal "github.com/LjungErik/zetra-kv-store/internal/raft"
	"github.com/LjungErik/zetra-kv-store/internal/store"
	"github.com/gin-gonic/gin"
)

type Server struct {
	store store.KVStore
	raft  *raft_internal.Raft
	proxy proxy.Proxy
}

type Option func(*Server)

func WithStore(store store.KVStore) Option {
	return func(s *Server) {
		s.store = store
	}
}

func WithRaft(raft *raft_internal.Raft) Option {
	return func(s *Server) {
		s.raft = raft
	}
}

func WithProxy(proxy proxy.Proxy) Option {
	return func(s *Server) {
		s.proxy = proxy
	}
}

func NewServer(options ...Option) *Server {
	s := &Server{}

	for _, opt := range options {
		opt(s)
	}

	return s
}

func (s *Server) Routes() http.Handler {
	router := gin.Default()

	apiReader := router.Group("/api/v1")
	apiReader.GET("/store/:key", s.getValue)

	apiWriter := router.Group("/api/v1")
	apiWriter.Use(s.leaderProxy())
	apiWriter.POST("/store", s.insertValue)
	apiWriter.DELETE("/store", s.deleteValue)

	return router
}
