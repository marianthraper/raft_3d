package raftnode

import (
	// "log"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"raft3d/fsm"
)

func SetupRaft(nodeID, raftDir string, fsm *fsm.FSM, bindAddr string) (*raft.Raft, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)

	os.MkdirAll(raftDir, 0700)

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "raft-log.bolt"))
	if err != nil {
		return nil, err
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "raft-stable.bolt"))
	if err != nil {
		return nil, err
	}

	snapshots, err := raft.NewFileSnapshotStore(raftDir, 2, os.Stdout)
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransport(bindAddr, nil, 3, 10*time.Second, os.Stdout)
	if err != nil {
		return nil, err
	}

	r, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, err
	}

	return r, nil
}
