package raft

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"

	"github.com/LjungErik/zetra-kv-store/internal/config"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

type Raft struct {
	*raft.Raft
	peers map[string]config.Peer
}

func SetupRaft(id, bindAddr, advertiseAddr, dataDir string, peers []config.Peer, fsm raft.FSM, cfg SetupConfig) (*Raft, error) {
	peersMap := make(map[string]config.Peer, len(peers))
	servers := make([]raft.Server, len(peers))

	for i, peer := range peers {
		peersMap[peer.ID] = peer
		servers[i] = raft.Server{
			ID:      raft.ServerID(peer.ID),
			Address: raft.ServerAddress(peer.Address),
		}
	}

	raft, err := setupRaft(id, bindAddr, advertiseAddr, dataDir, servers, fsm, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to setup raft: %w", err)
	}

	return &Raft{
		Raft:  raft,
		peers: peersMap,
	}, nil
}

func (r *Raft) IsLeaderNode() bool {
	return r.State() == raft.Leader
}

func (r *Raft) GetLeadersRestAddress() string {
	_, leaderID := r.LeaderWithID()

	peer := r.peers[string(leaderID)]

	return peer.RestAddress
}

func setupRaft(id, bindAddr, advertiseAddr, dataDir string, peers []raft.Server, fsm raft.FSM, cfg SetupConfig) (*raft.Raft, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(id)
	raftConfig.Logger = hclog.New(&hclog.LoggerOptions{
		Name:       "raft",
		Level:      hclog.Info,
		Output:     os.Stdout,
		JSONFormat: true,
	})

	if cfg.LeaderLeaseTimeout > 0 {
		raftConfig.LeaderLeaseTimeout = cfg.LeaderLeaseTimeout
	}

	if cfg.HeartbeatTimeout > 0 {
		raftConfig.HeartbeatTimeout = cfg.HeartbeatTimeout
	}
	if cfg.ElectionTimeout > 0 {
		raftConfig.ElectionTimeout = cfg.ElectionTimeout
	}
	if cfg.CommitTimeout > 0 {
		raftConfig.CommitTimeout = cfg.CommitTimeout
	}
	if cfg.SnapshotInterval > 0 {
		raftConfig.SnapshotInterval = cfg.SnapshotInterval
	}
	if cfg.SnapshotThreshold > 0 {
		raftConfig.SnapshotThreshold = cfg.SnapshotThreshold
	}
	if cfg.TrailingLogs > 0 {
		raftConfig.TrailingLogs = cfg.TrailingLogs
	}

	advertiseTCPAddr, err := net.ResolveTCPAddr("tcp", advertiseAddr)
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransport(bindAddr, advertiseTCPAddr, cfg.MaxPool, cfg.Timeout, io.Discard)
	if err != nil {
		return nil, err
	}

	logStore, err := raftboltdb.NewBoltStore(path.Join(dataDir, "raft-log.db"))
	if err != nil {
		return nil, err
	}

	stableStore, err := raftboltdb.NewBoltStore(path.Join(dataDir, "raft-stable.db"))
	if err != nil {
		return nil, err
	}

	snapshots, err := raft.NewFileSnapshotStore(dataDir, cfg.SnapshotsToRetain, io.Discard)
	if err != nil {
		return nil, err
	}

	r, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, err
	}

	r.BootstrapCluster(raft.Configuration{Servers: peers})

	return r, nil
}
