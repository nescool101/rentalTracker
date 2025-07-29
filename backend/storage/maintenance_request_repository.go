package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/nescool101/rentManager/model"
	supa "github.com/supabase-community/supabase-go"
)

// MaintenanceRequest represents a maintenance request in storage
type MaintenanceRequest struct {
	ID          string             `json:"id"`
	PropertyID  string             `json:"property_id"`
	RenterID    string             `json:"renter_id"`
	Description string             `json:"description"`
	RequestDate model.FlexibleTime `json:"request_date"`
	Status      string             `json:"status"`
}

// MaintenanceRequestRepository interfaces with the maintenance_request table
type MaintenanceRequestRepository struct {
	client *supa.Client
}

// NewMaintenanceRequestRepository creates a new maintenance request repository
func NewMaintenanceRequestRepository(client *supa.Client) *MaintenanceRequestRepository {
	return &MaintenanceRequestRepository{
		client: client,
	}
}

// GetAll retrieves all maintenance requests
func (r *MaintenanceRequestRepository) GetAll() ([]MaintenanceRequest, error) {
	var requests []MaintenanceRequest

	data, count, err := r.client.From("maintenance_request").Select("*", "exact", false).Execute()
	if err != nil {
		log.Printf("Error fetching maintenance requests: %v", err)
		return nil, fmt.Errorf("failed to fetch maintenance requests: %w", err)
	}

	log.Printf("Retrieved %d maintenance requests", count)

	// Parse the JSON response data into our struct
	err = json.Unmarshal([]byte(data), &requests)
	if err != nil {
		log.Printf("Error parsing maintenance request data: %v", err)
		return nil, fmt.Errorf("failed to parse maintenance request data: %w", err)
	}

	return requests, nil
}

// GetByID retrieves a maintenance request by ID
func (r *MaintenanceRequestRepository) GetByID(id string) (*MaintenanceRequest, error) {
	data, count, err := r.client.From("maintenance_request").Select("*", "exact", false).
		Eq("id", id).Execute()
	if err != nil {
		log.Printf("Error fetching maintenance request by ID: %v", err)
		return nil, fmt.Errorf("failed to fetch maintenance request: %w", err)
	}

	if count == 0 {
		return nil, errors.New("maintenance request not found")
	}

	// Parse the first result
	var requests []MaintenanceRequest
	err = json.Unmarshal([]byte(data), &requests)
	if err != nil {
		log.Printf("Error parsing maintenance request data: %v", err)
		return nil, fmt.Errorf("failed to parse maintenance request data: %w", err)
	}

	if len(requests) == 0 {
		return nil, errors.New("maintenance request not found")
	}

	return &requests[0], nil
}

// GetByPropertyID retrieves all maintenance requests for a property
func (r *MaintenanceRequestRepository) GetByPropertyID(propertyID string) ([]MaintenanceRequest, error) {
	data, count, err := r.client.From("maintenance_request").Select("*", "exact", false).
		Eq("property_id", propertyID).Execute()
	if err != nil {
		log.Printf("Error fetching maintenance requests for property: %v", err)
		return nil, fmt.Errorf("failed to fetch maintenance requests for property: %w", err)
	}

	log.Printf("Retrieved %d maintenance requests for property %s", count, propertyID)

	var requests []MaintenanceRequest
	err = json.Unmarshal([]byte(data), &requests)
	if err != nil {
		log.Printf("Error parsing maintenance request data: %v", err)
		return nil, fmt.Errorf("failed to parse maintenance request data: %w", err)
	}

	return requests, nil
}

// GetByPropertyIDs retrieves all maintenance requests for a list of property IDs
func (r *MaintenanceRequestRepository) GetByPropertyIDs(propertyIDs []string) ([]MaintenanceRequest, error) {
	if len(propertyIDs) == 0 {
		return []MaintenanceRequest{}, nil
	}

	// Ensure propertyIDs are valid UUIDs before querying to prevent injection or errors
	// (This step is optional but good practice; Supabase might handle it)
	// for _, pid := range propertyIDs {
	// 	if _, err := uuid.Parse(pid); err != nil {
	// 		return nil, fmt.Errorf("invalid property ID format: %s", pid)
	// 	}
	// }

	data, count, err := r.client.From("maintenance_request").Select("*", "exact", false).
		In("property_id", propertyIDs).Execute()
	if err != nil {
		log.Printf("Error fetching maintenance requests for multiple properties: %v", err)
		return nil, fmt.Errorf("failed to fetch maintenance requests for properties: %w", err)
	}

	log.Printf("Retrieved %d maintenance requests for %d properties", count, len(propertyIDs))

	var requests []MaintenanceRequest
	err = json.Unmarshal([]byte(data), &requests)
	if err != nil {
		log.Printf("Error parsing maintenance request data: %v", err)
		return nil, fmt.Errorf("failed to parse maintenance request data: %w", err)
	}

	return requests, nil
}

// GetByRenterID retrieves all maintenance requests from a renter
func (r *MaintenanceRequestRepository) GetByRenterID(renterID string) ([]MaintenanceRequest, error) {
	data, count, err := r.client.From("maintenance_request").Select("*", "exact", false).
		Eq("renter_id", renterID).Execute()
	if err != nil {
		log.Printf("Error fetching maintenance requests for renter: %v", err)
		return nil, fmt.Errorf("failed to fetch maintenance requests for renter: %w", err)
	}

	log.Printf("Retrieved %d maintenance requests for renter %s", count, renterID)

	var requests []MaintenanceRequest
	err = json.Unmarshal([]byte(data), &requests)
	if err != nil {
		log.Printf("Error parsing maintenance request data: %v", err)
		return nil, fmt.Errorf("failed to parse maintenance request data: %w", err)
	}

	return requests, nil
}

// GetByStatus retrieves all maintenance requests with a specific status
func (r *MaintenanceRequestRepository) GetByStatus(status string) ([]MaintenanceRequest, error) {
	data, count, err := r.client.From("maintenance_request").Select("*", "exact", false).
		Eq("status", status).Execute()
	if err != nil {
		log.Printf("Error fetching maintenance requests by status: %v", err)
		return nil, fmt.Errorf("failed to fetch maintenance requests by status: %w", err)
	}

	log.Printf("Retrieved %d maintenance requests with status %s", count, status)

	var requests []MaintenanceRequest
	err = json.Unmarshal([]byte(data), &requests)
	if err != nil {
		log.Printf("Error parsing maintenance request data: %v", err)
		return nil, fmt.Errorf("failed to parse maintenance request data: %w", err)
	}

	return requests, nil
}

// Create creates a new maintenance request
func (r *MaintenanceRequestRepository) Create(request *MaintenanceRequest) (*MaintenanceRequest, error) {
	if request.ID == "" {
		request.ID = uuid.New().String()
	}

	// Store the ID before insertion
	requestID := request.ID

	data, count, err := r.client.From("maintenance_request").Insert(*request, false, "exact", "", "").Execute()
	if err != nil {
		log.Printf("Error creating maintenance request: %v", err)
		return nil, fmt.Errorf("failed to create maintenance request: %w", err)
	}

	// If count is 0 but data was returned, try to parse it
	if count == 0 && len(data) > 0 {
		log.Printf("Insert returned count 0 but has data, attempting to parse: %s", data)
	}

	var createdRequest []MaintenanceRequest
	err = json.Unmarshal([]byte(data), &createdRequest)
	if err != nil {
		log.Printf("Error parsing maintenance request data: %v", err)

		// If we can't parse, try to fetch the request by ID as a fallback
		if count == 0 {
			log.Printf("Trying to fetch newly created request by ID as fallback...")
			return r.GetByID(requestID)
		}

		return nil, fmt.Errorf("failed to parse maintenance request data: %w", err)
	}

	if len(createdRequest) == 0 {
		// If empty result but count was 0, try fetching by ID as a fallback
		if count == 0 {
			log.Printf("Insert returned empty result, trying to fetch by ID as fallback...")
			return r.GetByID(requestID)
		}

		return nil, errors.New("no maintenance request was created")
	}

	return &createdRequest[0], nil
}

// Update updates an existing maintenance request
func (r *MaintenanceRequestRepository) Update(id string, request *MaintenanceRequest) (*MaintenanceRequest, error) {
	data, count, err := r.client.From("maintenance_request").Update(*request, "exact", "").
		Eq("id", id).Execute()
	if err != nil {
		log.Printf("Error updating maintenance request: %v", err)
		return nil, fmt.Errorf("failed to update maintenance request: %w", err)
	}

	if count == 0 {
		return nil, errors.New("maintenance request not found")
	}

	var updatedRequest []MaintenanceRequest
	err = json.Unmarshal([]byte(data), &updatedRequest)
	if err != nil {
		log.Printf("Error parsing maintenance request data: %v", err)
		return nil, fmt.Errorf("failed to parse maintenance request data: %w", err)
	}

	if len(updatedRequest) == 0 {
		return nil, errors.New("maintenance request not found")
	}

	return &updatedRequest[0], nil
}

// Delete deletes a maintenance request
func (r *MaintenanceRequestRepository) Delete(id string) error {
	_, _, err := r.client.From("maintenance_request").Delete("minimal", "").
		Eq("id", id).Execute()
	if err != nil {
		log.Printf("Error deleting maintenance request: %v", err)
		return fmt.Errorf("failed to delete maintenance request: %w", err)
	}

	return nil
}
