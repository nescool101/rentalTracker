package storage

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/nescool101/rentManager/model"
	supa "github.com/supabase-community/supabase-go"
)

// PricingRepository provides methods to interact with the pricing table
type PricingRepository struct {
	client *supa.Client
}

// NewPricingRepository creates a new PricingRepository
func NewPricingRepository(client *supa.Client) *PricingRepository {
	return &PricingRepository{
		client: client,
	}
}

// GetByRentalID retrieves pricing information for a specific rental ID
func (r *PricingRepository) GetByRentalID(ctx context.Context, rentalID uuid.UUID) (*model.Pricing, error) {
	var results []model.Pricing
	data, count, err := r.client.From("pricing").Select("*", "exact", false).
		Eq("rental_id", rentalID.String()).Execute()

	if err != nil {
		log.Printf("Error fetching pricing by rental_id %s: %v", rentalID, err)
		return nil, err
	}

	if count == 0 {
		return nil, nil // Not found
	}

	err = json.Unmarshal([]byte(data), &results)
	if err != nil {
		log.Printf("Error parsing pricing data for rental_id %s: %v", rentalID, err)
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil // Should not happen if count > 0, but good practice
	}

	return &results[0], nil
}

// Create adds new pricing information to the database
func (r *PricingRepository) Create(ctx context.Context, pricing model.Pricing) (*model.Pricing, error) {
	if pricing.ID == uuid.Nil {
		pricing.ID = uuid.New()
	}

	var createdPricing []model.Pricing
	data, count, err := r.client.From("pricing").Insert(pricing, false, "exact", "", "").Execute()
	if err != nil {
		log.Printf("Error creating pricing: %v", err)
		return nil, err
	}

	if count == 0 {
		// This might indicate an issue, or perhaps the insert didn't return the row by default
		log.Printf("Warning: pricing insert returned count 0. Data: %s", string(data))
		// Attempt to parse even if count is 0, as Supabase might return data anyway
	}

	err = json.Unmarshal(data, &createdPricing)
	if err != nil {
		log.Printf("Error unmarshalling created pricing data: %v. Data: %s", err, string(data))
		return nil, err
	}

	if len(createdPricing) == 0 {
		log.Printf("Error: no pricing data returned after insert. Data: %s", string(data))
		return nil, errors.New("failed to create pricing: no data returned")
	}

	return &createdPricing[0], nil
}

// GetByID retrieves pricing information by its ID
func (r *PricingRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Pricing, error) {
	var results []model.Pricing
	data, count, err := r.client.From("pricing").Select("*", "exact", false).
		Eq("id", id.String()).Execute()

	if err != nil {
		log.Printf("Error fetching pricing by id %s: %v", id, err)
		return nil, err
	}

	if count == 0 {
		return nil, nil // Not found
	}

	err = json.Unmarshal(data, &results)
	if err != nil {
		log.Printf("Error parsing pricing data for id %s: %v", id, err)
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil // Should not happen if count > 0
	}

	return &results[0], nil
}

// GetAll retrieves all pricing records from the database
func (r *PricingRepository) GetAll(ctx context.Context) ([]model.Pricing, error) {
	var pricingList []model.Pricing
	data, count, err := r.client.From("pricing").Select("*", "exact", false).Execute()
	if err != nil {
		log.Printf("Error fetching all pricing: %v", err)
		return nil, err
	}

	log.Printf("Retrieved %d pricing records", count)

	err = json.Unmarshal(data, &pricingList)
	if err != nil {
		log.Printf("Error parsing all pricing data: %v", err)
		return nil, err
	}
	return pricingList, nil
}

// Update modifies existing pricing information in the database
func (r *PricingRepository) Update(ctx context.Context, pricing model.Pricing) (*model.Pricing, error) {
	var updatedPricing []model.Pricing
	data, count, err := r.client.From("pricing").Update(pricing, "exact", "").
		Eq("id", pricing.ID.String()).Execute()

	if err != nil {
		log.Printf("Error updating pricing ID %s: %v", pricing.ID, err)
		return nil, err
	}

	if count == 0 {
		log.Printf("Warning: pricing update for ID %s returned count 0. Data: %s", pricing.ID, string(data))
		// Not found or no changes made. Supabase might return data if no changes.
	}

	err = json.Unmarshal(data, &updatedPricing)
	if err != nil {
		log.Printf("Error unmarshalling updated pricing data for ID %s: %v. Data: %s", pricing.ID, err, string(data))
		return nil, err
	}

	if len(updatedPricing) == 0 {
		// It's possible for an update to not return data if nothing changed or if the record doesn't exist.
		// To be more robust, one might re-fetch by ID here or ensure `returning=representation` (if client supports it well for Update)
		log.Printf("No pricing data returned after update for ID %s. Data: %s", pricing.ID, string(data))
		// Check if the record exists as a separate step if needed or return nil if no data
		return r.GetByID(ctx, pricing.ID) // Attempt to return the record state
	}

	return &updatedPricing[0], nil
}

// Delete removes pricing information from the database by its ID
func (r *PricingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, _, err := r.client.From("pricing").Delete("minimal", "").
		Eq("id", id.String()).Execute()

	if err != nil {
		log.Printf("Error deleting pricing ID %s: %v", id, err)
		return err
	}
	return nil
}

// GetByPropertyID retrieves pricing information for a specific property ID
func (r *PricingRepository) GetByPropertyID(ctx context.Context, propertyID uuid.UUID) ([]model.Pricing, error) {
	var results []model.Pricing
	data, count, err := r.client.From("pricing").Select("*", "exact", false).
		Eq("property_id", propertyID.String()).Execute()

	if err != nil {
		log.Printf("Error fetching pricing by property_id %s: %v", propertyID, err)
		return nil, err
	}

	if count == 0 {
		return []model.Pricing{}, nil // Empty slice, not found
	}

	err = json.Unmarshal([]byte(data), &results)
	if err != nil {
		log.Printf("Error parsing pricing data for property_id %s: %v", propertyID, err)
		return nil, err
	}

	return results, nil
}
