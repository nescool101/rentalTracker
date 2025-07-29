package storage

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/nescool101/rentManager/model"
)

// Get the file path from environment variable or use default
func getFilePath() string {
	path := os.Getenv("PAYERS_FILE_PATH")
	if path == "" {
		return "payers.json" // This default might be irrelevant now
	}
	return path
}

// InitializePayersFile was removed as payers are now in the database.

func GetPayers() ([]model.Payer, error) {
	// This function is likely obsolete if payers are in the database.
	// Keeping it for now, but it should not be called by new DB-centric logic.
	filePath := getFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var payers []model.Payer
	err = json.Unmarshal(data, &payers)
	return payers, err
}

func parseDate(dateStr string) time.Time {
	// This function might be used by GetPayers or other legacy file-based logic.
	// If GetPayers is removed, and this isn't used elsewhere, it can also be removed.
	layout := time.RFC3339
	parsedDate, err := time.Parse(layout, dateStr)
	if err != nil {
		log.Fatalf("Invalid date format for %s", dateStr)
	}
	return parsedDate
}
