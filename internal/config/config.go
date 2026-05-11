package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Peer struct {
	ID      string `mapstructure:"id"`
	Address string `mapstructure:"address"`
}

type NodeConfig struct {
	ID                string `mapstructure:"id"`
	RaftBindAddr      string `mapstructure:"raft_bind_addr"`
	RaftAdvertiseAddr string `mapstructure:"raft_advertise_addr"`
	HTTPAddr          string `mapstructure:"http_addr"`
	DataDir           string `mapstructure:"data_dir"`
}

type RaftConfig struct {
	MaxPool            int           `mapstructure:"max_pool"`
	Timeout            time.Duration `mapstructure:"timeout"`
	SnapshotsToRetain  int           `mapstructure:"snapshots_to_retain"`
	LeaderLeaseTimeout time.Duration `mapstructure:"leader_lease_timeout"`
	HeartbeatTimeout   time.Duration `mapstructure:"heartbeat_timeout"`
	ElectionTimeout    time.Duration `mapstructure:"election_timeout"`
	CommitTimeout      time.Duration `mapstructure:"commit_timeout"`
	SnapshotInterval   time.Duration `mapstructure:"snapshot_interval"`
	SnapshotThreshold  uint64        `mapstructure:"snapshot_threshold"`
	TrailingLogs       uint64        `mapstructure:"trailing_logs"`
}

type ClusterConfig struct {
	Peers []Peer `mapstructure:"peers"`
}

type Config struct {
	Node    NodeConfig    `mapstructure:"node"`
	Raft    RaftConfig    `mapstructure:"raft"`
	Cluster ClusterConfig `mapstructure:"cluster"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("/app")
	v.AddConfigPath(".")

	v.SetEnvPrefix("ZETRA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.BindEnv("node.id", "ZETRA_NODE_ID")
	v.BindEnv("node.raft_bind_addr", "ZETRA_NODE_RAFT_BIND_ADDR")
	v.BindEnv("node.raft_advertise_addr", "ZETRA_NODE_RAFT_ADVERTISE_ADDR")
	v.BindEnv("node.http_addr", "ZETRA_NODE_HTTP_ADDR")
	v.BindEnv("node.data_dir", "ZETRA_NODE_DATA_DIR")
	v.BindEnv("raft.max_pool", "ZETRA_RAFT_MAX_POOL")
	v.BindEnv("raft.timeout", "ZETRA_RAFT_TIMEOUT")
	v.BindEnv("raft.snapshots_to_retain", "ZETRA_RAFT_SNAPSHOTS_TO_RETAIN")
	v.BindEnv("raft.heartbeat_timeout", "ZETRA_RAFT_LEADER_LEASE_TIMEOUT")
	v.BindEnv("raft.heartbeat_timeout", "ZETRA_RAFT_HEARTBEAT_TIMEOUT")
	v.BindEnv("raft.election_timeout", "ZETRA_RAFT_ELECTION_TIMEOUT")
	v.BindEnv("raft.commit_timeout", "ZETRA_RAFT_COMMIT_TIMEOUT")
	v.BindEnv("raft.snapshot_interval", "ZETRA_RAFT_SNAPSHOT_INTERVAL")
	v.BindEnv("raft.snapshot_threshold", "ZETRA_RAFT_SNAPSHOT_THRESHOLD")
	v.BindEnv("raft.trailing_logs", "ZETRA_RAFT_TRAILING_LOGS")

	v.SetDefault("node.id", "node1")
	v.SetDefault("node.raft_bind_addr", "0.0.0.0:7001")
	v.SetDefault("node.raft_advertise_addr", "")
	v.SetDefault("node.http_addr", "0.0.0.0:8081")
	v.SetDefault("node.data_dir", "/tmp/node-data")
	v.SetDefault("raft.max_pool", 3)
	v.SetDefault("raft.timeout", "10s")
	v.SetDefault("raft.snapshots_to_retain", 2)
	v.SetDefault("raft.leader_lease_timeout", "100ms")
	v.SetDefault("raft.heartbeat_timeout", "200ms")
	v.SetDefault("raft.election_timeout", "200ms")
	v.SetDefault("raft.commit_timeout", "50ms")
	v.SetDefault("raft.snapshot_interval", "120s")
	v.SetDefault("raft.snapshot_threshold", 8192)
	v.SetDefault("raft.trailing_logs", 10240)
	v.SetDefault("cluster.peers", []map[string]interface{}{
		{"id": "node1", "address": "127.0.0.1:7001"},
		{"id": "node2", "address": "127.0.0.1:7002"},
		{"id": "node3", "address": "127.0.0.1:7003"},
	})

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if cfg.Node.RaftAdvertiseAddr == "" {
		cfg.Node.RaftAdvertiseAddr = cfg.Node.RaftBindAddr
	}

	return &cfg, nil
}
