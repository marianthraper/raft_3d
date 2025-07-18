package models

type Printer struct {
	ID      string `json:"id"`
	Company string `json:"company"`
	Model   string `json:"model"`
}

type Filament struct {
	ID                   string `json:"id"`
	Type                 string `json:"type"` // PLA, PETG, etc.
	Color                string `json:"color"`
	TotalWeightGrams     int    `json:"total_weight_grams"`
	RemainingWeightGrams int    `json:"remaining_weight_grams"`
}

type PrintJob struct {
	ID               string `json:"id"`
	PrinterID        string `json:"printer_id"`
	FilamentID       string `json:"filament_id"`
	FilePath         string `json:"filepath"`
	PrintWeightGrams int    `json:"print_weight_grams"`
	Status           string `json:"status"` // Queued, Running, Done, Canceled
}
