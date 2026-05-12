package raft

import (
	"context"
	"log/slog"

	"github.com/hashicorp/raft"
)

func WatchEvents(ctx context.Context, r *Raft) {
	ch := make(chan raft.Observation, 32)
	obs := raft.NewObserver(ch, false, nil)
	r.RegisterObserver(obs)
	defer r.DeregisterObserver(obs)

	for {
		select {
		case o := <-ch:
			logObservation(o)
		case <-ctx.Done():
			return
		}
	}
}

func logObservation(o raft.Observation) {
	switch v := o.Data.(type) {
	case raft.RaftState:
		slog.Info("raft state changed", "state", v.String())

	case raft.LeaderObservation:
		if v.LeaderAddr == "" {
			slog.Warn("raft leader lost")
		} else {
			slog.Info("raft leader elected",
				"leader_id", v.LeaderID,
				"leader_addr", v.LeaderAddr,
			)
		}

	case raft.PeerObservation:
		if v.Removed {
			slog.Info("raft peer removed",
				"peer_id", v.Peer.ID,
				"peer_addr", v.Peer.Address,
			)
		} else {
			slog.Info("raft peer added",
				"peer_id", v.Peer.ID,
				"peer_addr", v.Peer.Address,
			)
		}

	case raft.FailedHeartbeatObservation:
		slog.Warn("raft heartbeat failed",
			"peer_id", v.PeerID,
			"last_contact", v.LastContact,
		)

	case raft.ResumedHeartbeatObservation:
		slog.Info("raft heartbeat resumed", "peer_id", v.PeerID)
	}
}
