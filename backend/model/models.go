package model

import (
	"time"

	"github.com/google/uuid"
)

// Person represents a person in the system
type Person struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
	Phone    string    `json:"phone"`
	NIT      string    `json:"nit"`
}

// Role represents a role in the system
type Role struct {
	ID       uuid.UUID `json:"id"`
	RoleName string    `json:"role_name"`
}

// PersonRole represents the relationship between persons and roles
type PersonRole struct {
	ID       uuid.UUID `json:"id"`
	PersonID uuid.UUID `json:"person_id"`
	RoleID   uuid.UUID `json:"role_id"`
}

// Property represents a property in the system
type Property struct {
	ID         uuid.UUID   `json:"id"`
	Address    string      `json:"address"`
	AptNumber  string      `json:"apt_number"`
	City       string      `json:"city"`
	State      string      `json:"state"`
	ZipCode    string      `json:"zip_code"`
	Type       string      `json:"type"`
	ResidentID uuid.UUID   `json:"resident_id"`
	ManagerIDs []uuid.UUID `json:"manager_ids,omitempty"`
}

// BankAccount represents a bank account in the system
type BankAccount struct {
	ID            uuid.UUID `json:"id"`
	PersonID      uuid.UUID `json:"person_id"`
	BankName      string    `json:"bank_name"`
	AccountType   string    `json:"account_type"`
	AccountNumber string    `json:"account_number"`
	AccountHolder string    `json:"account_holder"`
}

// Rental represents a rental agreement in the system
type Rental struct {
	ID            uuid.UUID    `json:"id"`
	PropertyID    uuid.UUID    `json:"property_id"`
	RenterID      uuid.UUID    `json:"renter_id"`
	BankAccountID uuid.UUID    `json:"bank_account_id"`
	StartDate     FlexibleTime `json:"start_date"`
	EndDate       FlexibleTime `json:"end_date"`
	PaymentTerms  string       `json:"payment_terms"`
	UnpaidMonths  int          `json:"unpaid_months"`
}

// Pricing represents pricing information for a rental
type Pricing struct {
	ID                   uuid.UUID `json:"id"`
	RentalID             uuid.UUID `json:"rental_id"`
	MonthlyRent          float64   `json:"monthly_rent"`
	SecurityDeposit      float64   `json:"security_deposit"`
	UtilitiesIncluded    []string  `json:"utilities_included"`
	TenantResponsibleFor []string  `json:"tenant_responsible_for"`
	LateFee              float64   `json:"late_fee"`
	DueDay               int       `json:"due_day"`
}

// PaymentSchedule represents a payment schedule for a rental
type PaymentSchedule struct {
	ID             uuid.UUID `json:"id"`
	RentalID       uuid.UUID `json:"rental_id"`
	DueDate        time.Time `json:"due_date"`
	ExpectedAmount float64   `json:"expected_amount"`
	IsPaid         bool      `json:"is_paid"`
	PaidDate       time.Time `json:"paid_date"`
	ReminderSent   bool      `json:"reminder_sent"`
	Recurrence     string    `json:"recurrence"`
}

// RentPayment represents a rent payment
type RentPayment struct {
	ID          uuid.UUID    `json:"id"`
	RentalID    uuid.UUID    `json:"rental_id"`
	PaymentDate FlexibleTime `json:"payment_date"`
	AmountPaid  float64      `json:"amount_paid"`
	PaidOnTime  bool         `json:"paid_on_time"`
}

// Document represents a document attached to a rental
type Document struct {
	ID           uuid.UUID `json:"id"`
	RentalID     uuid.UUID `json:"rental_id"`
	FileURL      string    `json:"file_url"`
	DocumentType string    `json:"document_type"`
	UploadDate   time.Time `json:"upload_date"`
	UploaderID   uuid.UUID `json:"uploader_id"`
	Description  string    `json:"description"`
}

// RentalHistory represents a rental history record
type RentalHistory struct {
	ID       uuid.UUID `json:"id"`
	PersonID uuid.UUID `json:"person_id"`
}

// MaintenanceRequest represents a maintenance request for a property
type MaintenanceRequest struct {
	ID          uuid.UUID    `json:"id,omitempty"`
	PropertyID  uuid.UUID    `json:"property_id"`
	RenterID    uuid.UUID    `json:"renter_id,omitempty"`
	Description string       `json:"description"`
	RequestDate FlexibleTime `json:"request_date"`
	Status      string       `json:"status"`
	CreatedAt   FlexibleTime `json:"created_at,omitempty"`
	UpdatedAt   FlexibleTime `json:"updated_at,omitempty"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        uuid.UUID   `json:"id"`
	Action    string      `json:"action"`
	Entity    string      `json:"entity"`
	EntityID  uuid.UUID   `json:"entity_id"`
	ChangedBy uuid.UUID   `json:"changed_by"`
	Timestamp time.Time   `json:"timestamp"`
	Details   interface{} `json:"details"`
}

// User represents a user in the system
type User struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	PasswordBase64 string    `json:"password_base64"`
	Role           string    `json:"role"`
	PersonID       uuid.UUID `json:"person_id"`
	Status         string    `json:"status"` // values: 'pending', 'active', 'disabled'
}
