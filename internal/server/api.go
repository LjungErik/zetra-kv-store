package server

import (
	"log/slog"
	"net/http"

	"github.com/LjungErik/zetra-kv-store/internal/server/models"
	"github.com/gin-gonic/gin"
)

func (s *Server) getValue(ctx *gin.Context) {
	key := ctx.Param("key")

	if key == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "key not set"})
		return
	}

	value, ok := s.store.Get(key)
	if !ok {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	resp := models.KeyStoreGetResponse{
		Key:   key,
		Value: value,
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *Server) insertValue(ctx *gin.Context) {
	var req models.KeyStoreInsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to parse json", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := req.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.store.Set(req.Key, req.Value); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unable to store value for key"})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (s *Server) deleteValue(ctx *gin.Context) {
	var req models.KeyStoreDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to parse json", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := req.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.store.Delete(req.Key); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unable to delete key"})
		return
	}

	ctx.Status(http.StatusNoContent)
}
