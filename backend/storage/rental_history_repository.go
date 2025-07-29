package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nescool101/rentManager/model"
	supa "github.com/supabase-community/supabase-go"
)

// RentalHistory represents a rental history record
type RentalHistory struct {
	ID        string             `json:"id"`
	PersonID  string             `json:"person_id"`
	RentalID  string             `json:"rental_id"`
	Status    string             `json:"status"`
	EndReason string             `json:"end_reason"`
	EndDate   model.FlexibleTime `json:"end_date"`
}

// RentalHistoryRepository interfaces with the rental_history table
type RentalHistoryRepository struct {
	client *supa.Client
}

// NewRentalHistoryRepository creates a new rental history repository
func NewRentalHistoryRepository(client *supa.Client) *RentalHistoryRepository {
	return &RentalHistoryRepository{
		client: client,
	}
}

// GetAll retrieves all rental history records
func (r *RentalHistoryRepository) GetAll() ([]RentalHistory, error) {
	var histories []RentalHistory

	data, count, err := r.client.From("rental_history").Select("*", "exact", false).Execute()
	if err != nil {
		log.Printf("Error fetching rental histories: %v", err)
		return nil, fmt.Errorf("failed to fetch rental histories: %w", err)
	}

	log.Printf("Retrieved %d rental histories", count)

	// Parse the JSON response data into our struct
	err = json.Unmarshal([]byte(data), &histories)
	if err != nil {
		log.Printf("Error parsing rental history data: %v", err)
		return nil, fmt.Errorf("failed to parse rental history data: %w", err)
	}

	return histories, nil
}

// GetByID retrieves a rental history record by ID
func (r *RentalHistoryRepository) GetByID(id string) (*RentalHistory, error) {
	data, count, err := r.client.From("rental_history").Select("*", "exact", false).
		Eq("id", id).Execute()
	if err != nil {
		log.Printf("Error fetching rental history by ID: %v", err)
		return nil, fmt.Errorf("failed to fetch rental history: %w", err)
	}

	if count == 0 {
		return nil, errors.New("rental history not found")
	}

	// Parse the first result
	var histories []RentalHistory
	err = json.Unmarshal([]byte(data), &histories)
	if err != nil {
		log.Printf("Error parsing rental history data: %v", err)
		return nil, fmt.Errorf("failed to parse rental history data: %w", err)
	}

	if len(histories) == 0 {
		return nil, errors.New("rental history not found")
	}

	return &histories[0], nil
}

// GetByPersonID retrieves all rental history records for a specific person
func (r *RentalHistoryRepository) GetByPersonID(personID string) ([]RentalHistory, error) {
	data, count, err := r.client.From("rental_history").Select("*", "exact", false).
		Eq("person_id", personID).Execute()
	if err != nil {
		log.Printf("Error fetching rental histories for person: %v", err)
		return nil, fmt.Errorf("failed to fetch rental histories for person: %w", err)
	}

	log.Printf("Retrieved %d rental histories for person %s", count, personID)

	var histories []RentalHistory
	err = json.Unmarshal([]byte(data), &histories)
	if err != nil {
		log.Printf("Error parsing rental history data: %v", err)
		return nil, fmt.Errorf("failed to parse rental history data: %w", err)
	}

	return histories, nil
}

// GetByRentalID retrieves all rental history records for a specific rental
func (r *RentalHistoryRepository) GetByRentalID(rentalID string) ([]RentalHistory, error) {
	data, count, err := r.client.From("rental_history").Select("*", "exact", false).
		Eq("rental_id", rentalID).Execute()
	if err != nil {
		log.Printf("Error fetching rental histories for rental: %v", err)
		return nil, fmt.Errorf("failed to fetch rental histories for rental: %w", err)
	}

	log.Printf("Retrieved %d rental histories for rental %s", count, rentalID)

	var histories []RentalHistory
	err = json.Unmarshal([]byte(data), &histories)
	if err != nil {
		log.Printf("Error parsing rental history data: %v", err)
		return nil, fmt.Errorf("failed to parse rental history data: %w", err)
	}

	return histories, nil
}

// GetByRentalIDs retrieves all rental history records for a list of rental IDs
func (r *RentalHistoryRepository) GetByRentalIDs(rentalIDs []string) ([]RentalHistory, error) {
	if len(rentalIDs) == 0 {
		return []RentalHistory{}, nil
	}

	data, count, err := r.client.From("rental_history").Select("*", "exact", false).
		In("rental_id", rentalIDs).Execute()
	if err != nil {
		log.Printf("Error fetching rental histories for multiple rentals: %v", err)
		return nil, fmt.Errorf("failed to fetch rental histories for rentals: %w", err)
	}

	log.Printf("Retrieved %d rental histories for %d rentals", count, len(rentalIDs))

	var histories []RentalHistory
	err = json.Unmarshal([]byte(data), &histories)
	if err != nil {
		log.Printf("Error parsing rental history data: %v", err)
		return nil, fmt.Errorf("failed to parse rental history data: %w", err)
	}

	return histories, nil
}

// GetByStatus retrieves all rental history records with a specific status
func (r *RentalHistoryRepository) GetByStatus(status string) ([]RentalHistory, error) {
	data, count, err := r.client.From("rental_history").Select("*", "exact", false).
		Eq("status", status).Execute()
	if err != nil {
		log.Printf("Error fetching rental histories by status: %v", err)
		return nil, fmt.Errorf("failed to fetch rental histories by status: %w", err)
	}

	log.Printf("Retrieved %d rental histories with status %s", count, status)

	var histories []RentalHistory
	err = json.Unmarshal([]byte(data), &histories)
	if err != nil {
		log.Printf("Error parsing rental history data: %v", err)
		return nil, fmt.Errorf("failed to parse rental history data: %w", err)
	}

	return histories, nil
}

// Create creates a new rental history record
func (r *RentalHistoryRepository) Create(history *RentalHistory) (*RentalHistory, error) {
	if history.ID == "" {
		history.ID = uuid.New().String()
	}

	// Store the ID before insertion
	historyID := history.ID

	data, count, err := r.client.From("rental_history").Insert(*history, false, "exact", "", "").Execute()
	if err != nil {
		log.Printf("Error creating rental history: %v", err)
		return nil, fmt.Errorf("failed to create rental history: %w", err)
	}

	// If count is 0 but data was returned, try to parse it
	if count == 0 && len(data) > 0 {
		log.Printf("Insert returned count 0 but has data, attempting to parse: %s", data)
	}

	var createdHistory []RentalHistory
	err = json.Unmarshal([]byte(data), &createdHistory)
	if err != nil {
		log.Printf("Error parsing rental history data: %v", err)

		// If we can't parse, try to fetch the history by ID as a fallback
		if count == 0 {
			log.Printf("Trying to fetch newly created history by ID as fallback...")
			return r.GetByID(historyID)
		}

		return nil, fmt.Errorf("failed to parse rental history data: %w", err)
	}

	if len(createdHistory) == 0 {
		// If empty result but count was 0, try fetching by ID as a fallback
		if count == 0 {
			log.Printf("Insert returned empty result, trying to fetch by ID as fallback...")
			return r.GetByID(historyID)
		}

		return nil, errors.New("no rental history was created")
	}

	return &createdHistory[0], nil
}

// Update updates an existing rental history record
func (r *RentalHistoryRepository) Update(id string, history *RentalHistory) (*RentalHistory, error) {
	data, count, err := r.client.From("rental_history").Update(*history, "exact", "").
		Eq("id", id).Execute()
	if err != nil {
		log.Printf("Error updating rental history: %v", err)
		return nil, fmt.Errorf("failed to update rental history: %w", err)
	}

	if count == 0 {
		return nil, errors.New("rental history not found")
	}

	var updatedHistory []RentalHistory
	err = json.Unmarshal([]byte(data), &updatedHistory)
	if err != nil {
		log.Printf("Error parsing rental history data: %v", err)
		return nil, fmt.Errorf("failed to parse rental history data: %w", err)
	}

	if len(updatedHistory) == 0 {
		return nil, errors.New("rental history not found")
	}

	return &updatedHistory[0], nil
}

// Delete deletes a rental history record
func (r *RentalHistoryRepository) Delete(id string) error {
	_, _, err := r.client.From("rental_history").Delete("minimal", "").
		Eq("id", id).Execute()
	if err != nil {
		log.Printf("Error deleting rental history: %v", err)
		return fmt.Errorf("failed to delete rental history: %w", err)
	}

	return nil
}

// GetRentalHistoryByDateRange retrieves all rental history records with end dates in a specific range
func (r *RentalHistoryRepository) GetRentalHistoryByDateRange(startDate, endDate time.Time) ([]RentalHistory, error) {
	startDateStr := startDate.Format(time.RFC3339)
	endDateStr := endDate.Format(time.RFC3339)

	data, count, err := r.client.From("rental_history").Select("*", "exact", false).
		Gte("end_date", startDateStr).
		Lte("end_date", endDateStr).
		Execute()

	if err != nil {
		log.Printf("Error fetching rental histories by date range: %v", err)
		return nil, fmt.Errorf("failed to fetch rental histories by date range: %w", err)
	}

	log.Printf("Retrieved %d rental histories in date range", count)

	var histories []RentalHistory
	err = json.Unmarshal([]byte(data), &histories)
	if err != nil {
		log.Printf("Error parsing rental history data: %v", err)
		return nil, fmt.Errorf("failed to parse rental history data: %w", err)
	}

	return histories, nil
}
