package model

import "time"

// SigningStatus represents the current state of a signature request
type SigningStatus string

const (
	StatusPending  SigningStatus = "pending"
	StatusSigned   SigningStatus = "signed"
	StatusRejected SigningStatus = "rejected"
	StatusExpired  SigningStatus = "expired"
)

// ContractSigningInfo holds information for contract signing
type ContractSigningInfo struct {
	ContractID     string // UUID for the contract
	RecipientID    string // Person ID of the recipient
	RecipientEmail string // Email of the recipient
	PDFData        []byte // PDF data
	SignerName     string // Name of the signer
	SignatureID    string // UUID for the signature
}

// ContractSigningRequest represents a request to sign a contract
type ContractSigningRequest struct {
	ID             string        // UUID for this signing request
	ContractID     string        // Reference to contract
	RecipientID    string        // Person who needs to sign
	RecipientEmail string        // Email of recipient
	Status         SigningStatus // Current status
	CreatedAt      time.Time     // When created
	ExpiresAt      time.Time     // When expires
	SignedAt       *time.Time    // When signed (if signed)
	SignatureData  []byte        // The signature data (if signed)
}

// Spanish status translations for display purposes
var StatusTranslations = map[string]string{
	string(StatusPending):  "Pendiente",
	string(StatusSigned):   "Firmado",
	string(StatusRejected): "Rechazado",
	string(StatusExpired):  "Expirado",
}
