package store

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	raft_internal "github.com/LjungErik/zetra-kv-store/internal/raft"
	"github.com/hashicorp/raft"
)

const (
	defaultApplyTimeout = 5 * time.Second
)

type KVStore interface {
	Get(key string) (string, bool)
	Set(key, value string) error
	Delete(key string) error
}

type RaftKVStore struct {
	mu   sync.RWMutex
	data map[string]string
	raft *raft_internal.Raft
}

var _ raft.FSM = (*RaftKVStore)(nil)
var _ KVStore = (*RaftKVStore)(nil)

func NewKVStore() *RaftKVStore {
	return &RaftKVStore{
		data: make(map[string]string),
	}
}

func (r *RaftKVStore) SetRaft(raft *raft_internal.Raft) {
	r.raft = raft
}

func (r *RaftKVStore) Get(key string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	val, ok := r.data[key]
	return val, ok
}

func (r *RaftKVStore) Set(key string, value string) error {
	if !r.raft.IsLeaderNode() {
		return fmt.Errorf("not leader: %s", r.raft.Leader())
	}

	cmd, err := json.Marshal(raft_internal.Command{Op: raft_internal.OpSet, Key: key, Value: value})
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	return r.raft.Apply(cmd, defaultApplyTimeout).Error()
}

func (r *RaftKVStore) Delete(key string) error {
	if !r.raft.IsLeaderNode() {
		return fmt.Errorf("not leader")
	}

	cmd, err := json.Marshal(raft_internal.Command{Op: raft_internal.OpDelete, Key: key})
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	return r.raft.Apply(cmd, defaultApplyTimeout).Error()
}

func (r *RaftKVStore) Apply(log *raft.Log) interface{} {
	var cmd raft_internal.Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return fmt.Errorf("failed to unmarshal raft log data: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	switch cmd.Op {
	case raft_internal.OpSet:
		r.data[cmd.Key] = cmd.Value
	case raft_internal.OpDelete:
		delete(r.data, cmd.Key)
	}

	return nil
}

// Restore implements [raft.FSM].
func (r *RaftKVStore) Restore(snapshot io.ReadCloser) error {
	var data map[string]string
	if err := json.NewDecoder(snapshot).Decode(&data); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.data = data

	return nil
}

// Snapshot implements [raft.FSM].
func (r *RaftKVStore) Snapshot() (raft.FSMSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	copy := make(map[string]string, len(r.data))
	for k, v := range r.data {
		copy[k] = v
	}

	return &KVSnapshot{data: copy}, nil
}
