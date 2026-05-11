package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LjungErik/zetra-kv-store/internal/config"
	raft_internal "github.com/LjungErik/zetra-kv-store/internal/raft"
	"github.com/LjungErik/zetra-kv-store/internal/server"
	"github.com/LjungErik/zetra-kv-store/internal/store"
	"github.com/hashicorp/raft"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("starting node %s (raft: %s, http: %s)", cfg.Node.ID, cfg.Node.RaftAdvertiseAddr, cfg.Node.HTTPAddr)

	kvstore := store.NewKVStore()

	peers := make([]raft.Server, len(cfg.Cluster.Peers))
	for i, p := range cfg.Cluster.Peers {
		peers[i] = raft.Server{
			ID:      raft.ServerID(p.ID),
			Address: raft.ServerAddress(p.Address),
		}
	}

	raftCfg := raft_internal.SetupConfig{
		MaxPool:           cfg.Raft.MaxPool,
		Timeout:           cfg.Raft.Timeout,
		SnapshotsToRetain: cfg.Raft.SnapshotsToRetain,
		HeartbeatTimeout:  cfg.Raft.HeartbeatTimeout,
		ElectionTimeout:   cfg.Raft.ElectionTimeout,
		CommitTimeout:     cfg.Raft.CommitTimeout,
		SnapshotInterval:  cfg.Raft.SnapshotInterval,
		SnapshotThreshold: cfg.Raft.SnapshotThreshold,
		TrailingLogs:      cfg.Raft.TrailingLogs,
	}

	log.Println("setting up raft server...")

	raftInstance, err := raft_internal.SetupRaft(
		cfg.Node.ID,
		cfg.Node.RaftBindAddr,
		cfg.Node.RaftAdvertiseAddr,
		cfg.Node.DataDir,
		peers,
		kvstore,
		raftCfg,
	)
	if err != nil {
		log.Fatalf("failed to setup raft: %v", err)
	}

	kvstore.SetRaft(raftInstance)

	httpServer := server.NewServer(server.Config{
		Addr:  cfg.Node.HTTPAddr,
		Store: kvstore,
	})

	log.Println("starting http server...")

	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal("forced shutdown:", err)
	}

	log.Println("server stopped")
}
