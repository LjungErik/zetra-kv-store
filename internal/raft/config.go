package raft

import "time"

type SetupConfig struct {
	MaxPool            int
	Timeout            time.Duration
	SnapshotsToRetain  int
	LeaderLeaseTimeout time.Duration
	HeartbeatTimeout   time.Duration
	ElectionTimeout    time.Duration
	CommitTimeout      time.Duration
	SnapshotInterval   time.Duration
	SnapshotThreshold  uint64
	TrailingLogs       uint64
}
