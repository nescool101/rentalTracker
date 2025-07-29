package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nescool101/rentManager/model"
	supa "github.com/supabase-community/supabase-go"
)

// ContractSigningRepository handles storage operations for contract signatures
type ContractSigningRepository struct {
	client *supa.Client
}

// NewContractSigningRepository creates a new repository for contract signatures
func NewContractSigningRepository(client *supa.Client) *ContractSigningRepository {
	return &ContractSigningRepository{
		client: client,
	}
}

// ContractSigningRecord represents a contract signing record in the database
type ContractSigningRecord struct {
	ID             string     `json:"id"`
	ContractID     string     `json:"contract_id"`
	RecipientID    string     `json:"recipient_id"`
	RecipientEmail string     `json:"recipient_email"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      time.Time  `json:"expires_at"`
	SignedAt       *time.Time `json:"signed_at,omitempty"`
	RejectedAt     *time.Time `json:"rejected_at,omitempty"`
	SignatureData  []byte     `json:"signature_data,omitempty"`
	PDFPath        string     `json:"pdf_path,omitempty"`
	SignedPDFPath  string     `json:"signed_pdf_path,omitempty"`
}

// CreateSigningRequest creates a new contract signing request
func (r *ContractSigningRepository) CreateSigningRequest(ctx context.Context, request model.ContractSigningRequest) (*ContractSigningRecord, error) {
	record := ContractSigningRecord{
		ID:             request.ID,
		ContractID:     request.ContractID,
		RecipientID:    request.RecipientID,
		RecipientEmail: request.RecipientEmail,
		Status:         string(request.Status),
		CreatedAt:      request.CreatedAt,
		ExpiresAt:      request.ExpiresAt,
		SignedAt:       request.SignedAt,
	}

	data, count, err := r.client.From("contract_signatures").Insert(record, false, "exact", "", "").Execute()
	if err != nil {
		log.Printf("Error creating contract signature request: %v", err)
		return nil, err
	}

	if count == 0 {
		return nil, fmt.Errorf("no record created")
	}

	var createdRecords []ContractSigningRecord
	err = json.Unmarshal(data, &createdRecords)
	if err != nil {
		log.Printf("Error parsing created contract signature data: %v", err)
		return nil, err
	}

	if len(createdRecords) == 0 {
		return nil, fmt.Errorf("no record returned after creation")
	}

	return &createdRecords[0], nil
}

// GetByID retrieves a contract signing request by ID
func (r *ContractSigningRepository) GetByID(ctx context.Context, id string) (*ContractSigningRecord, error) {
	var records []ContractSigningRecord
	data, count, err := r.client.From("contract_signatures").Select("*", "exact", false).
		Eq("id", id).Execute()

	if err != nil {
		log.Printf("Error fetching contract signature by ID %s: %v", id, err)
		return nil, err
	}

	if count == 0 {
		return nil, nil // Not found
	}

	err = json.Unmarshal(data, &records)
	if err != nil {
		log.Printf("Error parsing contract signature data for ID %s: %v", id, err)
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil // Should not happen if count > 0, but good practice
	}

	return &records[0], nil
}

// GetByContractID retrieves contract signing requests by contract ID
func (r *ContractSigningRepository) GetByContractID(ctx context.Context, contractID string) ([]ContractSigningRecord, error) {
	var records []ContractSigningRecord
	data, count, err := r.client.From("contract_signatures").Select("*", "exact", false).
		Eq("contract_id", contractID).Execute()

	if err != nil {
		log.Printf("Error fetching contract signatures by contract_id %s: %v", contractID, err)
		return nil, err
	}

	if count == 0 {
		return []ContractSigningRecord{}, nil // Empty slice, not found
	}

	err = json.Unmarshal(data, &records)
	if err != nil {
		log.Printf("Error parsing contract signature data for contract_id %s: %v", contractID, err)
		return nil, err
	}

	return records, nil
}

// GetByRecipientID retrieves contract signing requests by recipient ID
func (r *ContractSigningRepository) GetByRecipientID(ctx context.Context, recipientID string) ([]ContractSigningRecord, error) {
	var records []ContractSigningRecord
	data, count, err := r.client.From("contract_signatures").Select("*", "exact", false).
		Eq("recipient_id", recipientID).Execute()

	if err != nil {
		log.Printf("Error fetching contract signatures by recipient_id %s: %v", recipientID, err)
		return nil, err
	}

	if count == 0 {
		return []ContractSigningRecord{}, nil // Empty slice, not found
	}

	err = json.Unmarshal(data, &records)
	if err != nil {
		log.Printf("Error parsing contract signature data for recipient_id %s: %v", recipientID, err)
		return nil, err
	}

	return records, nil
}

// GetPendingRequests retrieves contract signing requests that are pending and not expired
func (r *ContractSigningRepository) GetPendingRequests(ctx context.Context) ([]ContractSigningRecord, error) {
	var records []ContractSigningRecord
	data, count, err := r.client.From("contract_signatures").Select("*", "exact", false).
		Eq("status", "pending").
		Gt("expires_at", time.Now().Format(time.RFC3339)).
		Execute()

	if err != nil {
		log.Printf("Error fetching pending contract signatures: %v", err)
		return nil, err
	}

	if count == 0 {
		return []ContractSigningRecord{}, nil // Empty slice, not found
	}

	err = json.Unmarshal(data, &records)
	if err != nil {
		log.Printf("Error parsing pending contract signature data: %v", err)
		return nil, err
	}

	return records, nil
}

// MarkAsSigned marks a contract signing request as signed
func (r *ContractSigningRepository) MarkAsSigned(ctx context.Context, id string, signedPDFPath string) error {
	record, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if record == nil {
		return fmt.Errorf("signing request not found")
	}

	// Update fields
	now := time.Now()
	record.Status = string(model.StatusSigned)
	record.SignedAt = &now
	record.SignedPDFPath = signedPDFPath

	_, _, err = r.client.From("contract_signatures").Update(*record, "exact", "").
		Eq("id", id).Execute()

	if err != nil {
		log.Printf("Error marking contract signature as signed for ID %s: %v", id, err)
		return err
	}

	return nil
}

// MarkAsRejected marks a contract signing request as rejected
func (r *ContractSigningRepository) MarkAsRejected(ctx context.Context, id string) error {
	record, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if record == nil {
		return fmt.Errorf("signing request not found")
	}

	// Update fields
	now := time.Now()
	record.Status = string(model.StatusRejected)
	record.RejectedAt = &now

	_, _, err = r.client.From("contract_signatures").Update(*record, "exact", "").
		Eq("id", id).Execute()

	if err != nil {
		log.Printf("Error marking contract signature as rejected for ID %s: %v", id, err)
		return err
	}

	return nil
}

// UpdateExpiredStatuses updates statuses for expired signing requests
func (r *ContractSigningRepository) UpdateExpiredStatuses(ctx context.Context) (int, error) {
	// Find expired pending requests
	var records []ContractSigningRecord
	data, count, err := r.client.From("contract_signatures").Select("*", "exact", false).
		Eq("status", "pending").
		Lt("expires_at", time.Now().Format(time.RFC3339)).
		Execute()

	if err != nil {
		log.Printf("Error fetching expired contract signatures: %v", err)
		return 0, err
	}

	if count == 0 {
		return 0, nil // No expired requests
	}

	err = json.Unmarshal(data, &records)
	if err != nil {
		log.Printf("Error parsing expired contract signature data: %v", err)
		return 0, err
	}

	// Update each expired record
	updatedCount := 0
	for _, record := range records {
		record.Status = string(model.StatusExpired)
		_, _, err = r.client.From("contract_signatures").Update(record, "exact", "").
			Eq("id", record.ID).Execute()

		if err != nil {
			log.Printf("Error updating expired contract signature for ID %s: %v", record.ID, err)
			continue
		}
		updatedCount++
	}

	return updatedCount, nil
}
