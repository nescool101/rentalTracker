package model

import (
	"time"

	"github.com/google/uuid"
)

// ManagerRegistrationRequest represents all the data needed to register a new manager
type ManagerRegistrationRequest struct {
	// Person information
	FullName string `json:"full_name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	NIT      string `json:"nit"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`

	// Property information
	PropertyAddress   string `json:"property_address" binding:"required"`
	PropertyAptNumber string `json:"property_apt_number"`
	PropertyCity      string `json:"property_city" binding:"required"`
	PropertyState     string `json:"property_state" binding:"required"`
	PropertyZipCode   string `json:"property_zip_code"`
	PropertyType      string `json:"property_type" binding:"required"` // e.g., apartment, house, condo

	// Bank account information
	BankName      string `json:"bank_name" binding:"required"`
	AccountType   string `json:"account_type" binding:"required"`
	AccountNumber string `json:"account_number" binding:"required"`
	AccountHolder string `json:"account_holder" binding:"required"`

	// Pricing information
	MonthlyRent          float64  `json:"monthly_rent" binding:"required"`
	SecurityDeposit      float64  `json:"security_deposit" binding:"required"`
	UtilitiesIncluded    []string `json:"utilities_included"`
	TenantResponsibleFor []string `json:"tenant_responsible_for"`
	LateFee              float64  `json:"late_fee"`
	DueDay               int      `json:"due_day" binding:"required,min=1,max=31"`

	// Rental information
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	PaymentTerms string    `json:"payment_terms"`
}

// ManagerRegistrationResponse represents the response after a successful manager registration
type ManagerRegistrationResponse struct {
	Success    bool      `json:"success"`
	UserID     uuid.UUID `json:"user_id"`
	PersonID   uuid.UUID `json:"person_id"`
	PropertyID uuid.UUID `json:"property_id"`
	Message    string    `json:"message"`
}

// ManagerRegistrationTemplate represents the data sent in response to a GET request to pre-fill the form
type ManagerRegistrationTemplate struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}
