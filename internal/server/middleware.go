package server

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) leaderProxy() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if s.raft.IsLeaderNode() {
			slog.Info("Inside the leader")
			ctx.Next()
			return
		}

		ctx.Abort()

		if err := s.proxy.ProxyRequest(ctx, s.raft.GetLeadersRestAddress()); err != nil {
			ctx.JSON(http.StatusGatewayTimeout, gin.H{"error": "failed to forward request to leader"})
		}
	}
}
