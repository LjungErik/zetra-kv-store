package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/LjungErik/zetra-kv-store/internal/config"
	raft_internal "github.com/LjungErik/zetra-kv-store/internal/raft"
	"github.com/LjungErik/zetra-kv-store/internal/server"
	"github.com/LjungErik/zetra-kv-store/internal/store"
	"github.com/hashicorp/raft"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.Info("starting node",
		"id", cfg.Node.ID,
		"raft_addr", cfg.Node.RaftAdvertiseAddr,
		"http_addr", cfg.Node.HTTPAddr,
	)

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

	slog.Info("setting up raft")

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
		slog.Error("failed to setup raft", "error", err)
		os.Exit(1)
	}

	kvstore.SetRaft(raftInstance)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		raft_internal.WatchEvents(ctx, raftInstance)
	}()

	httpServer := server.NewServer(server.Config{
		Addr:  cfg.Node.HTTPAddr,
		Store: kvstore,
	})

	slog.Info("starting http server", "addr", cfg.Node.HTTPAddr)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("http server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("http server forced shutdown", "error", err)
	}

	if err := raftInstance.Shutdown().Error(); err != nil {
		slog.Error("raft shutdown error", "error", err)
	}

	wg.Wait()
	slog.Info("server stopped")
}
