package raft

import (
	"net"
	"os"
	"path"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

const (
	defaultMaxPool           = 3
	defaultTimeout           = 10 * time.Second
	defaultSnapshotsToRetain = 2
)

func SetupRaft(id, addr, dataDir string, peers []raft.Server, fsm raft.FSM) (*raft.Raft, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(id)

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransport(addr, tcpAddr, defaultMaxPool, defaultTimeout, os.Stderr)
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

	snapshots, err := raft.NewFileSnapshotStore(dataDir, defaultSnapshotsToRetain, os.Stderr)
	if err != nil {
		return nil, err
	}

	r, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, err
	}

	configuration := raft.Configuration{
		Servers: peers,
	}

	r.BootstrapCluster(configuration)

	return r, nil
}
