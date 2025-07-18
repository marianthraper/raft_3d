package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"raft3d/api"
	"raft3d/fsm"
	"raft3d/raftnode"
	"github.com/go-chi/chi/v5" // Ensure this import is present
)

func main() {
	// Setup Raft node and FSM
	nodeID := os.Getenv("NODE_ID")
	raftDir := fmt.Sprintf("raft-data-%s", nodeID)
	bindAddr := os.Getenv("BIND_ADDR")

	f := fsm.NewFSM()

	// Setup the Raft instance
	_, err := raftnode.SetupRaft(nodeID, raftDir, f, bindAddr)
	if err != nil {
		log.Fatalf("Failed to set up raft: %v", err)
	}

	// HTTP server setup
	routes := chi.NewRouter()
	api.InitAPI(routes, f)

	// Start HTTP server
	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", routes))
}
