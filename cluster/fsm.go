package cluster

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/raft"
	"io"
	"walkerDb/database"
)

type FSM struct {
	db *database.DB
}
type logEntryData struct {
	Op    string `json:"op,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// Apply applies a Raft log entry to the key-value store.
func (f *FSM) Apply(logEntry *raft.Log) interface{} {
	e := logEntryData{}
	if err := json.Unmarshal(logEntry.Data, &e); err != nil {
		panic("Failed unmarshaling Raft log entry. This is a bug.")
	}
	switch e.Op {
	case "set":
		return f.db.Put([]byte(e.Key), []byte(e.Value))
	case "delete":
		return f.db.Delete([]byte(e.Key))
	default:
		panic(fmt.Sprintf("unrecognized command op: %s", e.Op))
	}
}

// Snapshot returns a latest snapshot
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	// Clone the map.
	o := make(map[string][]byte)
	err := f.db.Fold(func(key []byte, value []byte) bool {
		o[string(key)] = value
		return true
	})
	if err != nil {
		return nil, err
	}
	return &fsmSnapshot{store: o}, nil
}

// Restore stores the key-value store to a previous state.
func (f *FSM) Restore(serialized io.ReadCloser) error {
	var newData map[string][]byte
	if err := json.NewDecoder(serialized).Decode(&newData); err != nil {
		return err
	}
	for key, value := range newData {
		err := f.db.Put([]byte(key), value)
		return err
	}
	return nil
}
