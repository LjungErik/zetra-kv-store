package store

import (
	"encoding/json"

	"github.com/hashicorp/raft"
)

type KVSnapshot struct {
	data map[string]string
}

var _ raft.FSMSnapshot = (*KVSnapshot)(nil)

func (s *KVSnapshot) Persist(sink raft.SnapshotSink) error {
	err := json.NewEncoder(sink).Encode(s.data)
	if err != nil {
		sink.Cancel()
		return err
	}

	return sink.Close()
}

func (s *KVSnapshot) Release() {}
