package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"raft3d/models"
	"raft3d/fsm"
	"github.com/hashicorp/raft"
	"github.com/go-chi/chi/v5"
)

var fsmInstance *fsm.FSM

func InitAPI(r *chi.Mux, raftFSM *fsm.FSM) {
	fsmInstance = raftFSM

	// Define routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/printers", createPrinter)
		r.Get("/printers", listPrinters)
		r.Post("/filaments", createFilament)
		r.Get("/filaments", listFilaments)
		r.Post("/print_jobs", createPrintJob)
		r.Get("/print_jobs", listPrintJobs)
		r.Post("/print_jobs/{job_id}/status", updatePrintJobStatus)
	})
}

// Handler to create a printer
func createPrinter(w http.ResponseWriter, r *http.Request) {
	var printer models.Printer
	if err := json.NewDecoder(r.Body).Decode(&printer); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Send Raft command to add printer
	command := fsm.Command{
		Type: "add_printer",
		Data: mustMarshal(printer),
	}
	applyRaftCommand(command)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(printer)
}

// Handler to list all printers
func listPrinters(w http.ResponseWriter, r *http.Request) {
	printers := fsmInstance.Printers
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(printers)
}

// Handler to create a filament
func createFilament(w http.ResponseWriter, r *http.Request) {
	var filament models.Filament
	if err := json.NewDecoder(r.Body).Decode(&filament); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Send Raft command to add filament
	command := fsm.Command{
		Type: "add_filament",
		Data: mustMarshal(filament),
	}
	applyRaftCommand(command)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(filament)
}

// Handler to list all filaments
func listFilaments(w http.ResponseWriter, r *http.Request) {
	filaments := fsmInstance.Filaments
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filaments)
}

// Handler to create a print job
func createPrintJob(w http.ResponseWriter, r *http.Request) {
	var job models.PrintJob
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Validate print weight and filament status
	validatePrintJob(job)

	// Send Raft command to add print job
	command := fsm.Command{
		Type: "add_printjob",
		Data: mustMarshal(job),
	}
	applyRaftCommand(command)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

// Handler to list all print jobs
func listPrintJobs(w http.ResponseWriter, r *http.Request) {
	printJobs := fsmInstance.PrintJobs
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(printJobs)
}

// Handler to update print job status
func updatePrintJobStatus(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "job_id")
	status := r.URL.Query().Get("status")

	// Validate status transition
	if err := validateStatusTransition(jobID, status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the print job
	job, exists := fsmInstance.PrintJobs[jobID]
	if !exists {
		http.Error(w, "Print job not found", http.StatusNotFound)
		return
	}

	// Update status
	job.Status = status

	// Send Raft command to update status
	command := fsm.Command{
		Type: "add_printjob",
		Data: mustMarshal(job),
	}
	applyRaftCommand(command)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// Helper to marshal data into JSON
func mustMarshal(data interface{}) json.RawMessage {
	marshaled, _ := json.Marshal(data)
	return marshaled
}

// Helper to apply a Raft command
func applyRaftCommand(command fsm.Command) {
	log := &raft.Log{
		Type:  raft.LogCommand,
		Data:  command.Data,
	}
	fsmInstance.Apply(log)
}

// Validate that print job's weight is within limits
func validatePrintJob(job models.PrintJob) {
	// Check if filament exists
	if _, exists := fsmInstance.Filaments[job.FilamentID]; !exists {
		panic("Filament not found")
	}

	// Check if print weight exceeds filament's remaining weight
	filament := fsmInstance.Filaments[job.FilamentID]
	if job.PrintWeightGrams > filament.RemainingWeightGrams {
		panic("Print weight exceeds available filament weight")
	}
}

// Validate the status transition
func validateStatusTransition(jobID, status string) error {
	job, exists := fsmInstance.PrintJobs[jobID]
	if !exists {
		return fmt.Errorf("Print job not found")
	}

	switch status {
	case "running":
		if job.Status != "queued" {
			return fmt.Errorf("Job must be queued before starting")
		}
	case "done":
		if job.Status != "running" {
			return fmt.Errorf("Job must be running before completing")
		}
	case "canceled":
		if job.Status != "queued" && job.Status != "running" {
			return fmt.Errorf("Job can only be canceled from queued or running")
		}
	default:
		return fmt.Errorf("Invalid status")
	}

	return nil
}
