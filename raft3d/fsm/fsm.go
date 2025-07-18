package fsm

import (
	// "bytes"
	"encoding/json"
	"io"
	"sync"

	"github.com/hashicorp/raft"
	"raft3d/models"
)

type FSM struct {
	mu        sync.Mutex
	Printers  map[string]models.Printer
	Filaments map[string]models.Filament
	PrintJobs map[string]models.PrintJob
}

// Command struct for incoming Raft logs
type Command struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func NewFSM() *FSM {
	return &FSM{
		Printers:  make(map[string]models.Printer),
		Filaments: make(map[string]models.Filament),
		PrintJobs: make(map[string]models.PrintJob),
	}
}

func (f *FSM) Apply(logEntry *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(logEntry.Data, &cmd); err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	switch cmd.Type {
	case "add_printer":
		var p models.Printer
		_ = json.Unmarshal(cmd.Data, &p)
		f.Printers[p.ID] = p
	case "add_filament":
		var fil models.Filament
		_ = json.Unmarshal(cmd.Data, &fil)
		f.Filaments[fil.ID] = fil
	case "add_printjob":
		var pj models.PrintJob
		_ = json.Unmarshal(cmd.Data, &pj)
		f.PrintJobs[pj.ID] = pj
	}
	return nil
}

// Snapshot takes a snapshot of the FSM
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	state := map[string]interface{}{
		"printers":  f.Printers,
		"filaments": f.Filaments,
		"printjobs": f.PrintJobs,
	}
	data, _ := json.Marshal(state)
	return &fsmSnapshot{snapshot: data}, nil
}

// Restore restores an FSM from a snapshot
func (f *FSM) Restore(rc io.ReadCloser) error {
	var state map[string]map[string]json.RawMessage
	data, _ := io.ReadAll(rc)
	return json.Unmarshal(data, &state)
}

// Snapshot struct
type fsmSnapshot struct {
	snapshot []byte
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := sink.Write(f.snapshot); err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (f *fsmSnapshot) Release() {}
