package storage

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	supa "github.com/supabase-community/supabase-go"

	"github.com/nescool101/rentManager/model"
)

// RentalRepository provides methods to interact with the Rental table in Supabase
type RentalRepository struct {
	client *supa.Client
}

// NewRentalRepository creates a new RentalRepository
func NewRentalRepository(client *supa.Client) *RentalRepository {
	return &RentalRepository{
		client: client,
	}
}

// GetAll retrieves all rentals from the database
func (r *RentalRepository) GetAll(ctx context.Context) ([]model.Rental, error) {
	var rentals []model.Rental

	data, count, err := r.client.From("rental").Select("*", "exact", false).Execute()
	if err != nil {
		log.Printf("Error fetching rentals: %v", err)
		return nil, err
	}

	log.Printf("Retrieved %d rentals", count)

	// Parse the JSON response data into our struct
	err = json.Unmarshal([]byte(data), &rentals)
	if err != nil {
		log.Printf("Error parsing rental data: %v", err)
		return nil, err
	}

	return rentals, nil
}

// GetByID retrieves a rental by ID
func (r *RentalRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Rental, error) {
	data, count, err := r.client.From("rental").Select("*", "exact", false).
		Eq("id", id.String()).Execute()
	if err != nil {
		log.Printf("Error fetching rental by ID: %v", err)
		return nil, err
	}

	if count == 0 {
		return nil, nil // Not found
	}

	// Parse the first result
	var rentals []model.Rental
	err = json.Unmarshal([]byte(data), &rentals)
	if err != nil {
		log.Printf("Error parsing rental data: %v", err)
		return nil, err
	}

	if len(rentals) == 0 {
		return nil, nil // Not found
	}

	return &rentals[0], nil
}

// GetByPropertyID retrieves rentals by property ID
func (r *RentalRepository) GetByPropertyID(ctx context.Context, propertyID uuid.UUID) ([]model.Rental, error) {
	data, count, err := r.client.From("rental").Select("*", "exact", false).
		Eq("property_id", propertyID.String()).Execute()
	if err != nil {
		log.Printf("Error fetching rentals by property ID: %v", err)
		return nil, err
	}

	if count == 0 {
		return []model.Rental{}, nil // No rentals found
	}

	var rentals []model.Rental
	err = json.Unmarshal([]byte(data), &rentals)
	if err != nil {
		log.Printf("Error parsing rental data: %v", err)
		return nil, err
	}

	return rentals, nil
}

// GetByRenterID retrieves rentals by renter ID
func (r *RentalRepository) GetByRenterID(ctx context.Context, renterID uuid.UUID) ([]model.Rental, error) {
	data, count, err := r.client.From("rental").Select("*", "exact", false).
		Eq("renter_id", renterID.String()).Execute()
	if err != nil {
		log.Printf("Error fetching rentals by renter ID: %v", err)
		return nil, err
	}

	if count == 0 {
		return []model.Rental{}, nil // No rentals found
	}

	var rentals []model.Rental
	err = json.Unmarshal([]byte(data), &rentals)
	if err != nil {
		log.Printf("Error parsing rental data: %v", err)
		return nil, err
	}

	return rentals, nil
}

// GetActiveRentals retrieves all active rentals (where end_date is in the future)
func (r *RentalRepository) GetActiveRentals(ctx context.Context) ([]model.Rental, error) {
	now := time.Now().Format(time.RFC3339)
	data, count, err := r.client.From("rental").Select("*", "exact", false).
		Gte("end_date", now).Execute()
	if err != nil {
		log.Printf("Error fetching active rentals: %v", err)
		return nil, err
	}

	if count == 0 {
		return []model.Rental{}, nil // No rentals found
	}

	var rentals []model.Rental
	err = json.Unmarshal([]byte(data), &rentals)
	if err != nil {
		log.Printf("Error parsing rental data: %v", err)
		return nil, err
	}

	return rentals, nil
}

// Create adds a new rental to the database
func (r *RentalRepository) Create(ctx context.Context, rental model.Rental) (*model.Rental, error) {
	// Convert FlexibleTime to time.Time format for database
	rentalData := map[string]interface{}{
		"id":              rental.ID.String(),
		"property_id":     rental.PropertyID.String(),
		"renter_id":       rental.RenterID.String(),
		"bank_account_id": rental.BankAccountID.String(),
		"start_date":      time.Time(rental.StartDate),
		"end_date":        time.Time(rental.EndDate),
		"payment_terms":   rental.PaymentTerms,
		"unpaid_months":   rental.UnpaidMonths,
	}

	data, count, err := r.client.From("rental").Insert(rentalData, false, "exact", "", "").Execute()
	if err != nil {
		log.Printf("Error creating rental: %v", err)
		return nil, err
	}

	if count == 0 {
		return nil, nil // No result returned
	}

	var rentals []model.Rental
	err = json.Unmarshal([]byte(data), &rentals)
	if err != nil {
		log.Printf("Error parsing rental data: %v", err)
		return nil, err
	}

	if len(rentals) == 0 {
		return nil, nil // No result returned
	}

	return &rentals[0], nil
}

// Update updates an existing rental
func (r *RentalRepository) Update(ctx context.Context, rental model.Rental) (*model.Rental, error) {
	// Convert FlexibleTime to time.Time format for database
	rentalData := map[string]interface{}{
		"id":              rental.ID.String(),
		"property_id":     rental.PropertyID.String(),
		"renter_id":       rental.RenterID.String(),
		"bank_account_id": rental.BankAccountID.String(),
		"start_date":      time.Time(rental.StartDate),
		"end_date":        time.Time(rental.EndDate),
		"payment_terms":   rental.PaymentTerms,
		"unpaid_months":   rental.UnpaidMonths,
	}

	data, count, err := r.client.From("rental").Update(rentalData, "exact", "").
		Eq("id", rental.ID.String()).Execute()
	if err != nil {
		log.Printf("Error updating rental: %v", err)
		return nil, err
	}

	if count == 0 {
		return nil, nil // No result returned
	}

	var rentals []model.Rental
	err = json.Unmarshal([]byte(data), &rentals)
	if err != nil {
		log.Printf("Error parsing rental data: %v", err)
		return nil, err
	}

	if len(rentals) == 0 {
		return nil, nil // No result returned
	}

	return &rentals[0], nil
}

// Delete removes a rental from the database
func (r *RentalRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, _, err := r.client.From("rental").Delete("minimal", "").
		Eq("id", id.String()).Execute()
	if err != nil {
		log.Printf("Error deleting rental: %v", err)
		return err
	}

	return nil
}
