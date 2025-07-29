package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	supa "github.com/supabase-community/supabase-go"

	"github.com/nescool101/rentManager/model"
)

// BankAccountRepository provides methods to interact with the bank_account table in Supabase
type BankAccountRepository struct {
	client *supa.Client
}

// NewBankAccountRepository creates a new BankAccountRepository
func NewBankAccountRepository(client *supa.Client) *BankAccountRepository {
	return &BankAccountRepository{
		client: client,
	}
}

// GetAll retrieves all bank accounts
func (r *BankAccountRepository) GetAll(ctx context.Context) ([]model.BankAccount, error) {
	var accounts []model.BankAccount

	data, count, err := r.client.From("bank_account").Select("*", "exact", false).Execute()
	if err != nil {
		log.Printf("Error fetching bank accounts: %v", err)
		return nil, err
	}

	log.Printf("Retrieved %d bank accounts", count)

	err = json.Unmarshal([]byte(data), &accounts)
	if err != nil {
		log.Printf("Error parsing bank account data: %v", err)
		return nil, err
	}

	return accounts, nil
}

// GetByID retrieves a bank account by ID
func (r *BankAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.BankAccount, error) {
	data, count, err := r.client.From("bank_account").Select("*", "exact", false).
		Eq("id", id.String()).Execute()
	if err != nil {
		log.Printf("Error fetching bank account by ID %s: %v", id, err)
		return nil, err
	}

	if count == 0 {
		return nil, nil // Not found
	}

	var accounts []model.BankAccount
	err = json.Unmarshal([]byte(data), &accounts)
	if err != nil {
		log.Printf("Error parsing bank account data: %v", err)
		return nil, err
	}

	if len(accounts) == 0 {
		return nil, nil // Not found
	}

	return &accounts[0], nil
}

// GetByPersonID retrieves bank accounts by person ID
func (r *BankAccountRepository) GetByPersonID(ctx context.Context, personID uuid.UUID) ([]model.BankAccount, error) {
	data, count, err := r.client.From("bank_account").Select("*", "exact", false).
		Eq("person_id", personID.String()).Execute()
	if err != nil {
		log.Printf("Error fetching bank accounts by person ID %s: %v", personID, err)
		return nil, err
	}

	if count == 0 {
		return []model.BankAccount{}, nil // No accounts found
	}

	var accounts []model.BankAccount
	err = json.Unmarshal([]byte(data), &accounts)
	if err != nil {
		log.Printf("Error parsing bank account data: %v", err)
		return nil, err
	}

	return accounts, nil
}

// Create adds a new bank account
func (r *BankAccountRepository) Create(ctx context.Context, account model.BankAccount) (*model.BankAccount, error) {
	accountData := map[string]interface{}{
		"id":             account.ID.String(),
		"person_id":      account.PersonID.String(),
		"bank_name":      account.BankName,
		"account_type":   account.AccountType,
		"account_number": account.AccountNumber,
		"account_holder": account.AccountHolder,
	}

	data, _, err := r.client.From("bank_account").Insert(accountData, false, "exact", "", "").Execute()
	if err != nil {
		log.Printf("Error creating bank account: %v", err)
		return nil, fmt.Errorf("failed to create bank account: %w", err)
	}

	var createdAccounts []model.BankAccount
	err = json.Unmarshal(data, &createdAccounts)
	if err != nil {
		log.Printf("Error parsing created bank account data: %v", err)
		return nil, err
	}

	if len(createdAccounts) == 0 {
		return nil, fmt.Errorf("failed to parse created bank account, empty result set")
	}

	return &createdAccounts[0], nil
}

// Update updates an existing bank account
func (r *BankAccountRepository) Update(ctx context.Context, account model.BankAccount) (*model.BankAccount, error) {
	accountData := map[string]interface{}{
		"person_id":      account.PersonID.String(),
		"bank_name":      account.BankName,
		"account_type":   account.AccountType,
		"account_number": account.AccountNumber,
		"account_holder": account.AccountHolder,
	}

	data, count, err := r.client.From("bank_account").Update(accountData, "exact", "").
		Eq("id", account.ID.String()).Execute()
	if err != nil {
		log.Printf("Error updating bank account %s: %v", account.ID, err)
		return nil, err
	}

	if count == 0 {
		return nil, fmt.Errorf("failed to update bank account, no rows affected or returned")
	}

	var updatedAccounts []model.BankAccount
	err = json.Unmarshal(data, &updatedAccounts)
	if err != nil {
		log.Printf("Error parsing updated bank account data: %v", err)
		return nil, err
	}

	if len(updatedAccounts) == 0 {
		return nil, fmt.Errorf("failed to parse updated bank account, empty result set")
	}

	return &updatedAccounts[0], nil
}

// Delete removes a bank account
func (r *BankAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, _, err := r.client.From("bank_account").Delete("minimal", "").
		Eq("id", id.String()).Execute()
	if err != nil {
		log.Printf("Error deleting bank account %s: %v", id, err)
		return err
	}

	return nil
}
