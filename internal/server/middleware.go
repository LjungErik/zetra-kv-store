package server

import (
	"net/http"

	"github.com/hashicorp/raft"
)

func (hs *HTTPServer) leaderProxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hs.raft.State() == raft.Leader {
			next.ServeHTTP(w, r)
			return
		}

		leaderAddr := hs.raft.Leader()

		if err := hs.proxy.ProxyRequest(w, r, leaderAddr); err != nil {
			handleError(w, err)
		}
	})
}
