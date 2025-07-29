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

// RentPayment represents a payment made for a rental
type RentPayment struct {
	ID          string             `json:"id"`
	RentalID    string             `json:"rental_id"`
	PaymentDate model.FlexibleTime `json:"payment_date"`
	AmountPaid  float64            `json:"amount_paid"`
	PaidOnTime  bool               `json:"paid_on_time"`
}

// RentPaymentRepository interfaces with the rent_payment table
type RentPaymentRepository struct {
	client *supa.Client
}

// NewRentPaymentRepository creates a new rent payment repository
func NewRentPaymentRepository(client *supa.Client) *RentPaymentRepository {
	return &RentPaymentRepository{
		client: client,
	}
}

// GetAll retrieves all rent payments
func (r *RentPaymentRepository) GetAll() ([]RentPayment, error) {
	var payments []RentPayment

	data, count, err := r.client.From("rent_payment").Select("*", "exact", false).Execute()
	if err != nil {
		log.Printf("Error fetching rent payments: %v", err)
		return nil, fmt.Errorf("failed to fetch rent payments: %w", err)
	}

	log.Printf("Retrieved %d rent payments", count)

	// Parse the JSON response data into our struct
	err = json.Unmarshal([]byte(data), &payments)
	if err != nil {
		log.Printf("Error parsing rent payment data: %v", err)
		return nil, fmt.Errorf("failed to parse rent payment data: %w", err)
	}

	return payments, nil
}

// GetByID retrieves a rent payment by ID
func (r *RentPaymentRepository) GetByID(id string) (*RentPayment, error) {
	data, count, err := r.client.From("rent_payment").Select("*", "exact", false).
		Eq("id", id).Execute()
	if err != nil {
		log.Printf("Error fetching rent payment by ID: %v", err)
		return nil, fmt.Errorf("failed to fetch rent payment: %w", err)
	}

	if count == 0 {
		return nil, errors.New("rent payment not found")
	}

	// Parse the first result
	var payments []RentPayment
	err = json.Unmarshal([]byte(data), &payments)
	if err != nil {
		log.Printf("Error parsing rent payment data: %v", err)
		return nil, fmt.Errorf("failed to parse rent payment data: %w", err)
	}

	if len(payments) == 0 {
		return nil, errors.New("rent payment not found")
	}

	return &payments[0], nil
}

// GetByRentalID retrieves all payments for a specific rental
func (r *RentPaymentRepository) GetByRentalID(rentalID string) ([]RentPayment, error) {
	data, count, err := r.client.From("rent_payment").Select("*", "exact", false).
		Eq("rental_id", rentalID).Execute()
	if err != nil {
		log.Printf("Error fetching rent payments for rental: %v", err)
		return nil, fmt.Errorf("failed to fetch rent payments for rental: %w", err)
	}

	log.Printf("Retrieved %d rent payments for rental %s", count, rentalID)

	var payments []RentPayment
	err = json.Unmarshal([]byte(data), &payments)
	if err != nil {
		log.Printf("Error parsing rent payment data: %v", err)
		return nil, fmt.Errorf("failed to parse rent payment data: %w", err)
	}

	return payments, nil
}

// GetByRentalIDs retrieves all payments for a list of rental IDs
func (r *RentPaymentRepository) GetByRentalIDs(rentalIDs []string) ([]RentPayment, error) {
	if len(rentalIDs) == 0 {
		return []RentPayment{}, nil
	}

	data, count, err := r.client.From("rent_payment").Select("*", "exact", false).
		In("rental_id", rentalIDs).Execute()
	if err != nil {
		log.Printf("Error fetching rent payments for multiple rentals: %v", err)
		return nil, fmt.Errorf("failed to fetch rent payments for rentals: %w", err)
	}

	log.Printf("Retrieved %d rent payments for %d rentals", count, len(rentalIDs))

	var payments []RentPayment
	err = json.Unmarshal([]byte(data), &payments)
	if err != nil {
		log.Printf("Error parsing rent payment data: %v", err)
		return nil, fmt.Errorf("failed to parse rent payment data: %w", err)
	}

	return payments, nil
}

// Create creates a new rent payment
func (r *RentPaymentRepository) Create(payment *RentPayment) (*RentPayment, error) {
	if payment.ID == "" {
		payment.ID = uuid.New().String()
	}

	// Store the ID before insertion
	paymentID := payment.ID

	data, count, err := r.client.From("rent_payment").Insert(*payment, false, "exact", "", "").Execute()
	if err != nil {
		log.Printf("Error creating rent payment: %v", err)
		return nil, fmt.Errorf("failed to create rent payment: %w", err)
	}

	// If count is 0 but data was returned, try to parse it
	if count == 0 && len(data) > 0 {
		log.Printf("Insert returned count 0 but has data, attempting to parse: %s", data)
	}

	var createdPayment []RentPayment
	err = json.Unmarshal([]byte(data), &createdPayment)
	if err != nil {
		log.Printf("Error parsing rent payment data: %v", err)

		// If we can't parse, try to fetch the payment by ID as a fallback
		if count == 0 {
			log.Printf("Trying to fetch newly created payment by ID as fallback...")
			return r.GetByID(paymentID)
		}

		return nil, fmt.Errorf("failed to parse rent payment data: %w", err)
	}

	if len(createdPayment) == 0 {
		// If empty result but count was 0, try fetching by ID as a fallback
		if count == 0 {
			log.Printf("Insert returned empty result, trying to fetch by ID as fallback...")
			return r.GetByID(paymentID)
		}

		return nil, errors.New("no rent payment was created")
	}

	return &createdPayment[0], nil
}

// Update updates an existing rent payment
func (r *RentPaymentRepository) Update(id string, payment *RentPayment) (*RentPayment, error) {
	data, count, err := r.client.From("rent_payment").Update(*payment, "exact", "").
		Eq("id", id).Execute()
	if err != nil {
		log.Printf("Error updating rent payment: %v", err)
		return nil, fmt.Errorf("failed to update rent payment: %w", err)
	}

	if count == 0 {
		return nil, errors.New("rent payment not found")
	}

	var updatedPayment []RentPayment
	err = json.Unmarshal([]byte(data), &updatedPayment)
	if err != nil {
		log.Printf("Error parsing rent payment data: %v", err)
		return nil, fmt.Errorf("failed to parse rent payment data: %w", err)
	}

	if len(updatedPayment) == 0 {
		return nil, errors.New("rent payment not found")
	}

	return &updatedPayment[0], nil
}

// Delete deletes a rent payment
func (r *RentPaymentRepository) Delete(id string) error {
	_, _, err := r.client.From("rent_payment").Delete("minimal", "").
		Eq("id", id).Execute()
	if err != nil {
		log.Printf("Error deleting rent payment: %v", err)
		return fmt.Errorf("failed to delete rent payment: %w", err)
	}

	return nil
}

// GetPaymentsByDateRange retrieves all payments within a specific date range
func (r *RentPaymentRepository) GetPaymentsByDateRange(startDate, endDate time.Time) ([]RentPayment, error) {
	startDateStr := startDate.Format(time.RFC3339)
	endDateStr := endDate.Format(time.RFC3339)

	data, count, err := r.client.From("rent_payment").Select("*", "exact", false).
		Gte("payment_date", startDateStr).
		Lte("payment_date", endDateStr).
		Execute()
	if err != nil {
		log.Printf("Error fetching rent payments by date range: %v", err)
		return nil, fmt.Errorf("failed to fetch rent payments by date range: %w", err)
	}

	log.Printf("Retrieved %d rent payments in date range", count)

	var payments []RentPayment
	err = json.Unmarshal([]byte(data), &payments)
	if err != nil {
		log.Printf("Error parsing rent payment data: %v", err)
		return nil, fmt.Errorf("failed to parse rent payment data: %w", err)
	}

	return payments, nil
}

// GetLatePayments retrieves all payments that were not paid on time
func (r *RentPaymentRepository) GetLatePayments() ([]RentPayment, error) {
	data, count, err := r.client.From("rent_payment").Select("*", "exact", false).
		Eq("paid_on_time", "false").Execute()
	if err != nil {
		log.Printf("Error fetching late rent payments: %v", err)
		return nil, fmt.Errorf("failed to fetch late rent payments: %w", err)
	}

	log.Printf("Retrieved %d late rent payments", count)

	var payments []RentPayment
	err = json.Unmarshal([]byte(data), &payments)
	if err != nil {
		log.Printf("Error parsing rent payment data: %v", err)
		return nil, fmt.Errorf("failed to parse rent payment data: %w", err)
	}

	return payments, nil
}
