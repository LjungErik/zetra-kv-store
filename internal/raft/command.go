package raft

type OperationType string

const (
	OpSet    OperationType = "set"
	OpDelete OperationType = "delete"
)

type Command struct {
	Op    OperationType `json:"op"`
	Key   string        `json:"key"`
	Value string        `json:"value"`
}
