package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	raft_internal "github.com/LjungErik/zetra-kv-store/internal/raft"
	"github.com/LjungErik/zetra-kv-store/internal/server"
	"github.com/LjungErik/zetra-kv-store/internal/store"
	"github.com/hashicorp/raft"
)

func getEnvOrDefault(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	return val
}

func main() {
	var kvstore = store.NewKVStore()
	var peers = []raft.Server{
		{ID: "node1", Address: "127.0.0.1:7001"},
		{ID: "node2", Address: "127.0.0.1:7002"},
		{ID: "node3", Address: "127.0.0.1:7003"},
	}

	var id = getEnvOrDefault("NODE_ID", "node1")
	var ip = getEnvOrDefault("NODE_IP", "127.0.0.1")
	var port = getEnvOrDefault("NODE_PORT", "7001")
	var http_ip = getEnvOrDefault("NODE_HTTP_IP", "127.0.0.1")
	var http_port = getEnvOrDefault("NODE_HTTP_PORT", "8081")
	var dataDir = getEnvOrDefault("NODE_DATA_DIR", "/tmp/node1-data")

	var addr = fmt.Sprintf("%s:%s", ip, port)
	var http_addr = fmt.Sprintf("%s:%s", http_ip, http_port)

	log.Println("setting up raft server...")

	raftInstance, err := raft_internal.SetupRaft(id, addr, dataDir, peers, kvstore)
	if err != nil {
		panic(err)
	}

	kvstore.SetRaft(raftInstance)

	cfg := server.Config{
		Addr:  http_addr,
		Store: kvstore,
	}

	var httpServer = server.NewServer(cfg)

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
