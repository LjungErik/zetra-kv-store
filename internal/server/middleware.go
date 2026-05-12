package server

import (
	"log/slog"
	"net/http"
)

func (hs *HTTPServer) leaderProxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Trying to check if it is leader", "is_leader", hs.raft.IsLeaderNode())

		if hs.raft.IsLeaderNode() {
			next.ServeHTTP(w, r)
			return
		}

		if err := hs.proxy.ProxyRequest(w, r, hs.raft.GetLeadersRestAddress()); err != nil {
			handleError(w, err)
		}
	})
}
