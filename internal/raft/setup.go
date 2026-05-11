package raft

import (
	"io"
	"net"
	"os"
	"path"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

func SetupRaft(id, bindAddr, advertiseAddr, dataDir string, peers []raft.Server, fsm raft.FSM, cfg SetupConfig) (*raft.Raft, error) {
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
