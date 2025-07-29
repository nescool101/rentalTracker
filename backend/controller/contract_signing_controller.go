package controller

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/digitorus/pdf"
	"github.com/digitorus/pdfsign/revocation"
	"github.com/digitorus/pdfsign/sign"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/service"
	"github.com/nescool101/rentManager/storage"
)

// SignatureMetadata holds additional information to include in the signature
type SignatureMetadata struct {
	SignID     string // ID of the signature request
	SignedBy   string // Email of the person who signed
	TimeSigned string // Timestamp when signed
}

// ContractSigningController handles operations related to contract signing
type ContractSigningController struct {
	personRepo         *storage.PersonRepository
	propertyRepo       *storage.PropertyRepository
	pricingRepo        *storage.PricingRepository
	userRepo           *storage.UserRepository
	contractController *ContractController
	signingRepo        *storage.ContractSigningRepository
}

// NewContractSigningController creates a new ContractSigningController
func NewContractSigningController(
	personRepo *storage.PersonRepository,
	propertyRepo *storage.PropertyRepository,
	pricingRepo *storage.PricingRepository,
	userRepo *storage.UserRepository,
	contractController *ContractController,
	signingRepo *storage.ContractSigningRepository,
) *ContractSigningController {
	// Generate self-signed certificates for development
	certsDir := "./certs"
	if _, err := os.Stat(filepath.Join(certsDir, "certificate.crt")); os.IsNotExist(err) {
		log.Printf("Generating self-signed certificates in %s", certsDir)
		if err := service.GenerateSelfSignedCert(certsDir); err != nil {
			log.Printf("Warning: Failed to generate self-signed certificates: %v", err)
		}
	}

	return &ContractSigningController{
		personRepo:         personRepo,
		propertyRepo:       propertyRepo,
		pricingRepo:        pricingRepo,
		userRepo:           userRepo,
		contractController: contractController,
		signingRepo:        signingRepo,
	}
}

// SigningRequest represents a request to initiate a contract signing
type SigningRequest struct {
	ContractID     string `json:"contract_id" binding:"required"`
	RecipientID    string `json:"recipient_id" binding:"required"`
	ExpirationDays int    `json:"expiration_days"`
}

// SigningStatusResponse represents the current status of a signing request
type SigningStatusResponse struct {
	ID          string     `json:"id"`
	ContractID  string     `json:"contract_id"`
	RecipientID string     `json:"recipient_id"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   time.Time  `json:"expires_at"`
	SignedAt    *time.Time `json:"signed_at,omitempty"`
}

// RegisterRoutes registers the contract signing routes
func (ctrl *ContractSigningController) RegisterRoutes(router *gin.RouterGroup) {
	signingRoutes := router.Group("/contract-signing")
	{
		// Routes that require authentication
		signingRoutes.POST("/request", ctrl.CreateSigningRequest)
	}

	// Public routes that don't require authentication
	publicRoutes := router.Group("/public/contract-signing")
	{
		publicRoutes.GET("/status/:id", ctrl.GetSigningStatus)
		publicRoutes.POST("/sign/:id", ctrl.SignContract)
		publicRoutes.POST("/reject/:id", ctrl.RejectContract)
		publicRoutes.GET("/pdf/:id", ctrl.ServePDF)
	}

	// Keep original endpoints for backward compatibility but make them public too
	router.GET("/contract-signing/status/:id", ctrl.GetSigningStatus)
	router.POST("/contract-signing/sign/:id", ctrl.SignContract)
	router.POST("/contract-signing/reject/:id", ctrl.RejectContract)
	router.GET("/contract-signing/pdf/:id", ctrl.ServePDF)
}

// RegisterPublicRoutes registers only the public contract signing routes
func (ctrl *ContractSigningController) RegisterPublicRoutes(router *gin.RouterGroup) {
	// Public routes that don't require authentication
	publicRoutes := router.Group("/public/contract-signing")
	{
		publicRoutes.GET("/status/:id", ctrl.GetSigningStatus)
		publicRoutes.POST("/sign/:id", ctrl.SignContract)
		publicRoutes.POST("/reject/:id", ctrl.RejectContract)
		publicRoutes.GET("/pdf/:id", ctrl.ServePDF)
	}

	// Keep original endpoints for backward compatibility but make them public too
	router.GET("/contract-signing/status/:id", ctrl.GetSigningStatus)
	router.POST("/contract-signing/sign/:id", ctrl.SignContract)
	router.POST("/contract-signing/reject/:id", ctrl.RejectContract)
	router.GET("/contract-signing/pdf/:id", ctrl.ServePDF)
}

// RegisterAuthRoutes registers only the authenticated contract signing routes
func (ctrl *ContractSigningController) RegisterAuthRoutes(router *gin.RouterGroup) {
	signingRoutes := router.Group("/contract-signing")
	{
		// Routes that require authentication
		signingRoutes.POST("/request", ctrl.CreateSigningRequest)
	}
}

// CreateSigningRequest initiates a contract signing process
func (ctrl *ContractSigningController) CreateSigningRequest(c *gin.Context) {
	var req SigningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Set default expiration days if not provided
	if req.ExpirationDays <= 0 {
		req.ExpirationDays = 7 // Default expiration is 7 days
	}

	// Parse UUIDs
	_, err := uuid.Parse(req.ContractID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contract ID"})
		return
	}

	recipientID, err := uuid.Parse(req.RecipientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipient ID"})
		return
	}

	// Get recipient details
	recipient, err := ctrl.personRepo.GetByID(c, recipientID)
	if err != nil {
		log.Printf("Error getting recipient: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recipient details"})
		return
	}
	if recipient == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipient not found"})
		return
	}

	// Get recipient email from user record
	recipientUser, err := ctrl.userRepo.GetByPersonID(c, recipientID)
	if err != nil {
		log.Printf("Error getting recipient user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recipient user details"})
		return
	}
	if recipientUser == nil || recipientUser.Email == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipient email not found"})
		return
	}

	// Generate and retrieve the contract PDF
	// In a real implementation, you would retrieve the PDF from storage
	// For now, we'll use the existing contract controller to regenerate it
	// TODO: Get the actual contract PDF data

	// Create a new signing request
	signingID := uuid.New().String()

	// Mock PDF data for now - in a real implementation, you would get the actual PDF
	mockPDFData := []byte("Sample PDF data for contract " + req.ContractID)

	// Create signing info
	signingInfo := model.ContractSigningInfo{
		ContractID:     req.ContractID,
		RecipientID:    req.RecipientID,
		RecipientEmail: recipientUser.Email,
		PDFData:        mockPDFData,
		SignerName:     recipient.FullName,
		SignatureID:    signingID,
	}

	// Create the signature request
	signingRequest, err := service.CreateSignatureRequest(signingInfo, req.ExpirationDays)
	if err != nil {
		log.Printf("Error creating signature request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create signature request"})
		return
	}

	// Save the signing request to the database
	if ctrl.signingRepo != nil {
		_, err = ctrl.signingRepo.CreateSigningRequest(c, *signingRequest)
		if err != nil {
			log.Printf("Error saving signature request to database: %v", err)
			// Continue anyway since the email has been sent
		}
	}

	// Return the signature request details
	c.JSON(http.StatusOK, gin.H{
		"message":    "Signature request created and email sent",
		"signing_id": signingRequest.ID,
		"expires_at": signingRequest.ExpiresAt,
	})
}

// GetSigningStatus retrieves the status of a signing request
func (ctrl *ContractSigningController) GetSigningStatus(c *gin.Context) {
	signingID := c.Param("id")
	if signingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Signing ID is required"})
		return
	}

	// If repository is available, get real status
	if ctrl.signingRepo != nil {
		record, err := ctrl.signingRepo.GetByID(c, signingID)
		if err != nil {
			log.Printf("Error getting signing request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get signing request"})
			return
		}

		if record == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Signing request not found"})
			return
		}

		// Get Spanish translation of status
		spanishStatus := model.StatusTranslations[record.Status]
		if spanishStatus == "" {
			spanishStatus = record.Status // Fallback to English if no translation found
		}

		c.JSON(http.StatusOK, gin.H{
			"id":             record.ID,
			"contract_id":    record.ContractID,
			"recipient_id":   record.RecipientID,
			"status":         record.Status,
			"status_spanish": spanishStatus,
			"created_at":     record.CreatedAt,
			"expires_at":     record.ExpiresAt,
			"signed_at":      record.SignedAt,
		})
		return
	}

	// If no repository, return mock status
	c.JSON(http.StatusOK, gin.H{
		"id":             signingID,
		"status":         "pending",
		"status_spanish": "Pendiente",
		"message":        "Esta solicitud de firma está pendiente.",
	})
}

// SignContract marks a contract as signed
func (ctrl *ContractSigningController) SignContract(c *gin.Context) {
	signingId := c.Param("id")
	if signingId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Signing ID is required"})
		return
	}

	// If repository is available, update actual record
	if ctrl.signingRepo != nil {
		// Get the signing request
		record, err := ctrl.signingRepo.GetByID(c, signingId)
		if err != nil {
			log.Printf("Error getting signing request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get signing request"})
			return
		}

		if record == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Signing request not found"})
			return
		}

		// If already signed or rejected, return error
		if record.Status == string(model.StatusSigned) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Contract already signed"})
			return
		}

		if record.Status == string(model.StatusRejected) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Contract signing was rejected"})
			return
		}

		if record.Status == string(model.StatusExpired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Contract signing request has expired"})
			return
		}

		// Get signerName and email
		var signerName string
		var signerEmail string

		if recipient, err := ctrl.personRepo.GetByID(c, uuid.MustParse(record.RecipientID)); err == nil && recipient != nil {
			signerName = recipient.FullName
		} else {
			signerName = record.RecipientEmail // Fallback to email if name not available
		}

		signerEmail = record.RecipientEmail

		// Create a basic contract data structure for signing
		// In a real implementation, this data should be retrieved from the contract record
		contractData := service.ContractPDF{
			Renter:       &model.Person{FullName: signerName},
			Owner:        nil, // Will use defaults
			Property:     nil, // Will use defaults
			Pricing:      nil, // Will use defaults
			CoSigner:     nil, // Will use defaults
			Witness:      nil, // Will use defaults
			StartDate:    time.Now(),
			EndDate:      time.Now().AddDate(0, 6, 0), // 6 months default
			CreationDate: time.Now(),
		}

		// Use the simple PDF signing approach with the new template
		signedPDFData, err := service.SimpleSignPDF(
			contractData,
			signerName,
			signerEmail,
			signingId,
		)

		if err != nil {
			log.Printf("Error signing PDF with simple approach: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign PDF: " + err.Error()})
			return
		}

		// Make sure the temp directory exists
		tempDir := filepath.Join(os.TempDir(), "contracts")
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			log.Printf("Error creating temp directory: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temporary directory"})
			return
		}

		// Define output path for signed PDF
		signedPDFPath := filepath.Join(tempDir, record.ContractID+"_signed.pdf")

		// Save the signed PDF to file
		if err := os.WriteFile(signedPDFPath, signedPDFData, 0644); err != nil {
			log.Printf("Error writing signed PDF to file: %v", err)
			// Continue anyway as we still have the signed PDF data
		}

		// Mark as signed in the database
		err = ctrl.signingRepo.MarkAsSigned(c, signingId, signedPDFPath)
		if err != nil {
			log.Printf("Error marking signing request as signed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark signing request as signed"})
			return
		}

		// Create signing info for sending the signed PDF back to the signer
		signingInfo := &model.ContractSigningRequest{
			ID:             record.ID,
			ContractID:     record.ContractID,
			RecipientID:    record.RecipientID,
			RecipientEmail: record.RecipientEmail,
		}

		// Send the signed PDF to the signer via email
		err = service.SendSignedPDFByEmail(signingInfo, signedPDFData)
		if err != nil {
			log.Printf("Error sending signed PDF by email: %v", err)
			// Continue anyway as the contract is already marked as signed
		}

		currentTime := time.Now().Format(time.RFC3339)
		c.JSON(http.StatusOK, gin.H{
			"id":       signingId,
			"status":   "signed",
			"signedAt": currentTime,
			"signedBy": record.RecipientEmail,
			"message":  "Contract successfully signed",
		})
		return
	}

	// If no repository, return mock response
	c.JSON(http.StatusOK, gin.H{
		"id":      signingId,
		"status":  "signed",
		"message": "Contract successfully signed",
	})
}

// createProperPDF creates a valid PDF file with contract information
func createProperPDF(outputPath string, contractID string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create a minimal valid PDF file
	// This is a very basic PDF structure with a simple text content
	pdfContent := []byte{
		// PDF header
		'%', 'P', 'D', 'F', '-', '1', '.', '4', '\n',
		// Simple object structure
		'1', ' ', '0', ' ', 'o', 'b', 'j', '\n',
		'<', '<', '\n',
		'/', 'T', 'y', 'p', 'e', ' ', '/', 'C', 'a', 't', 'a', 'l', 'o', 'g', '\n',
		'/', 'P', 'a', 'g', 'e', 's', ' ', '2', ' ', '0', ' ', 'R', '\n',
		'>', '>', '\n',
		'e', 'n', 'd', 'o', 'b', 'j', '\n',
		// Pages object
		'2', ' ', '0', ' ', 'o', 'b', 'j', '\n',
		'<', '<', '\n',
		'/', 'T', 'y', 'p', 'e', ' ', '/', 'P', 'a', 'g', 'e', 's', '\n',
		'/', 'K', 'i', 'd', 's', ' ', '[', '3', ' ', '0', ' ', 'R', ']', '\n',
		'/', 'C', 'o', 'u', 'n', 't', ' ', '1', '\n',
		'>', '>', '\n',
		'e', 'n', 'd', 'o', 'b', 'j', '\n',
		// Page object
		'3', ' ', '0', ' ', 'o', 'b', 'j', '\n',
		'<', '<', '\n',
		'/', 'T', 'y', 'p', 'e', ' ', '/', 'P', 'a', 'g', 'e', '\n',
		'/', 'P', 'a', 'r', 'e', 'n', 't', ' ', '2', ' ', '0', ' ', 'R', '\n',
		'/', 'C', 'o', 'n', 't', 'e', 'n', 't', 's', ' ', '4', ' ', '0', ' ', 'R', '\n',
		'>', '>', '\n',
		'e', 'n', 'd', 'o', 'b', 'j', '\n',
		// Content stream
		'4', ' ', '0', ' ', 'o', 'b', 'j', '\n',
		'<', '<', ' ', '/', 'L', 'e', 'n', 'g', 't', 'h', ' ', '1', '0', '0', ' ', '>', '>', '\n',
		's', 't', 'r', 'e', 'a', 'm', '\n',
		'B', 'T', '\n',
		'/', 'F', '1', ' ', '1', '2', ' ', 'T', 'f', '\n',
		'1', '0', '0', ' ', '7', '0', '0', ' ', 'T', 'd', '\n',
		'(', 'C', 'o', 'n', 't', 'r', 'a', 'c', 't', ' ', 'I', 'D', ':', ' ',
	}

	// Add contract ID to the content
	pdfContent = append(pdfContent, []byte(contractID)...)

	// Add current date
	pdfContent = append(pdfContent, []byte{
		')', ' ', 'T', 'j', '\n',
		'0', ' ', '-', '2', '0', ' ', 'T', 'd', '\n',
		'(', 'D', 'a', 't', 'e', ':', ' ',
	}...)
	pdfContent = append(pdfContent, []byte(time.Now().Format("2006-01-02"))...)

	// Add signature placeholder
	pdfContent = append(pdfContent, []byte{
		')', ' ', 'T', 'j', '\n',
		'0', ' ', '-', '2', '0', ' ', 'T', 'd', '\n',
		'(', 'F', 'I', 'R', 'M', 'A', 'D', 'O', ':', ' ', '_', '_', '_', '_', '_', '_', '_', '_', '_', '_', '_', '_', '_', '_', '_', '_', '_', '_', '_', '_', ')', ' ', 'T', 'j', '\n',
		'E', 'T', '\n',
		'e', 'n', 'd', 's', 't', 'r', 'e', 'a', 'm', '\n',
		'e', 'n', 'd', 'o', 'b', 'j', '\n',
		// Cross-reference table
		'x', 'r', 'e', 'f', '\n',
		'0', ' ', '5', '\n',
		'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', ' ', '6', '5', '5', '3', '5', ' ', 'f', '\n',
		'0', '0', '0', '0', '0', '0', '0', '0', '1', '0', ' ', '0', '0', '0', '0', '0', ' ', 'n', '\n',
		'0', '0', '0', '0', '0', '0', '0', '1', '6', '5', ' ', '0', '0', '0', '0', '0', ' ', 'n', '\n',
		'0', '0', '0', '0', '0', '0', '0', '2', '5', '0', ' ', '0', '0', '0', '0', '0', ' ', 'n', '\n',
		'0', '0', '0', '0', '0', '0', '0', '3', '3', '5', ' ', '0', '0', '0', '0', '0', ' ', 'n', '\n',
		// Trailer
		't', 'r', 'a', 'i', 'l', 'e', 'r', '\n',
		'<', '<', '\n',
		'/', 'S', 'i', 'z', 'e', ' ', '5', '\n',
		'/', 'R', 'o', 'o', 't', ' ', '1', ' ', '0', ' ', 'R', '\n',
		'>', '>', '\n',
		's', 't', 'a', 'r', 't', 'x', 'r', 'e', 'f', '\n',
		'4', '9', '5', '\n',
		'%', '%', 'E', 'O', 'F', '\n',
	}...)

	// Write to file
	if err := os.WriteFile(outputPath, pdfContent, 0644); err != nil {
		return fmt.Errorf("failed to write PDF file: %w", err)
	}

	log.Printf("Created proper PDF at %s for contract ID %s", outputPath, contractID)
	return nil
}

// signPDFWithDigitorus signs a PDF using the Digitorus library
func signPDFWithDigitorus(input, output, signerName string, metadata *SignatureMetadata) error {
	log.Printf("Starting PDF signing process. Input: %s, Output: %s", input, output)

	// Check if input file exists
	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("input PDF file does not exist: %s", input)
	}

	input_file, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("failed to open input PDF: %w", err)
	}
	defer input_file.Close()

	// Create parent directory for output file if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	output_file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create output PDF: %w", err)
	}
	defer output_file.Close()

	finfo, err := input_file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats: %w", err)
	}
	size := finfo.Size()

	if size == 0 {
		return fmt.Errorf("input PDF file is empty")
	}

	// Validate that it's a real PDF by checking the header
	header := make([]byte, 4)
	if _, err := input_file.ReadAt(header, 0); err != nil {
		return fmt.Errorf("failed to read PDF header: %w", err)
	}

	if string(header) != "%PDF" {
		return fmt.Errorf("file is not a valid PDF - invalid header: %s", string(header))
	}

	// Reset file pointer to beginning
	if _, err := input_file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// Create a PDF reader
	rdr, err := pdf.NewReader(input_file, size)
	if err != nil {
		return fmt.Errorf("failed to create PDF reader: %w", err)
	}

	// Get certificate and private key from cert files
	certificate, privateKey, err := getSigningCertificateAndKey()
	if err != nil {
		return fmt.Errorf("failed to get certificate and key: %w", err)
	}

	// Get certificate chains
	certificate_chains, err := certificate.Verify(getX509VerifyOptions())
	if err != nil {
		log.Printf("Warning: Certificate verification error: %v", err)
		// Continue with empty chains
		certificate_chains = [][]*x509.Certificate{}
	}

	// Use the signer name in uppercase for the signature
	upperCaseSignerName := strings.ToUpper(signerName)

	// Create contact info with metadata
	contactInfo := "RentalFullNescao System"
	if metadata != nil {
		contactInfo = fmt.Sprintf("SignID: %s | SignedBy: %s | TimeSigned: %s",
			metadata.SignID, metadata.SignedBy, metadata.TimeSigned)
	}

	log.Printf("Signing PDF with signer name: %s, Contact info: %s", upperCaseSignerName, contactInfo)

	// Sign the PDF using Digitorus - this adds the signature without modifying existing content
	err = sign.Sign(input_file, output_file, rdr, size, sign.SignData{
		Signature: sign.SignDataSignature{
			Info: sign.SignDataSignatureInfo{
				Name:        upperCaseSignerName,
				Location:    "Digital Signature",
				Reason:      "Contract signing",
				ContactInfo: contactInfo,
				Date:        time.Now().Local(),
			},
			CertType:   sign.CertificationSignature,
			DocMDPPerm: sign.AllowFillingExistingFormFieldsAndSignaturesPerms,
		},
		Signer:            privateKey,
		DigestAlgorithm:   crypto.SHA256,
		Certificate:       certificate,
		CertificateChains: certificate_chains,
		// TSA settings for timestamp authority
		TSA: sign.TSA{
			URL:      "https://freetsa.org/tsr",
			Username: "",
			Password: "",
		},

		// RevocationData and RevocationFunction for certificate validation
		RevocationData:     revocation.InfoArchival{},
		RevocationFunction: sign.DefaultEmbedRevocationStatusFunction,
	})
	if err != nil {
		return fmt.Errorf("failed to sign PDF: %w", err)
	}

	log.Printf("PDF signed successfully. Signed PDF written to %s", output)
	return nil
}

// getSigningCertificateAndKey retrieves the certificate and private key for signing
func getSigningCertificateAndKey() (*x509.Certificate, crypto.Signer, error) {
	// In a real implementation, you would load your certificate and private key
	// For this example, we'll use the self-signed certificate generated by service.GenerateSelfSignedCert

	certPath := filepath.Join("./certs", "certificate.crt")
	keyPath := filepath.Join("./certs", "private.key")

	// Load certificate
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, err
	}

	certificate, err := parseCertificate(certData)
	if err != nil {
		return nil, nil, err
	}

	// Load private key
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, err
	}

	privateKey, err := parsePrivateKey(keyData)
	if err != nil {
		return nil, nil, err
	}

	return certificate, privateKey, nil
}

// parseCertificate parses a PEM encoded certificate
func parseCertificate(certPEMBlock []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(certPEMBlock)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, errors.New("failed to decode PEM block containing certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// parsePrivateKey parses a PEM encoded private key
func parsePrivateKey(keyPEMBlock []byte) (crypto.Signer, error) {
	block, _ := pem.Decode(keyPEMBlock)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return privateKey, nil

	case "PRIVATE KEY":
		privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		switch k := privateKey.(type) {
		case *rsa.PrivateKey:
			return k, nil
		default:
			return nil, errors.New("unsupported private key type")
		}

	default:
		return nil, errors.New("unsupported PEM block type: " + block.Type)
	}
}

// getX509VerifyOptions returns the options for certificate verification
func getX509VerifyOptions() x509.VerifyOptions {
	// In a real implementation, you would configure this with proper root CAs
	return x509.VerifyOptions{
		// You might want to add trusted roots here
		Roots: nil, // Use system roots
	}
}

// RejectContract marks a contract signing request as rejected
func (ctrl *ContractSigningController) RejectContract(c *gin.Context) {
	signingID := c.Param("id")
	if signingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Signing ID is required"})
		return
	}

	// If repository is available, update actual record
	if ctrl.signingRepo != nil {
		err := ctrl.signingRepo.MarkAsRejected(c, signingID)
		if err != nil {
			log.Printf("Error marking signing request as rejected: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark signing request as rejected"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":      signingID,
			"status":  "rejected",
			"message": "Contract signing rejected",
		})
		return
	}

	// If no repository, return mock response
	c.JSON(http.StatusOK, gin.H{
		"id":      signingID,
		"status":  "rejected",
		"message": "Contract signing rejected",
	})
}

// ServePDF serves the contract PDF for viewing or download
func (ctrl *ContractSigningController) ServePDF(c *gin.Context) {
	signingId := c.Param("id")
	if signingId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Signing ID is required"})
		return
	}

	// Check if the signed version is requested
	isSigned := c.Query("signed") == "true"

	// If repository is available, get the actual record
	if ctrl.signingRepo != nil {
		// Get the signing request
		record, err := ctrl.signingRepo.GetByID(c, signingId)
		if err != nil {
			log.Printf("Error getting signing request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get signing request"})
			return
		}

		if record == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Signing request not found"})
			return
		}

		tempDir := filepath.Join(os.TempDir(), "contracts")
		var pdfPath string

		// Determine which file to serve - the original or signed version
		if isSigned && record.Status == string(model.StatusSigned) {
			// Serve the signed PDF
			pdfPath = filepath.Join(tempDir, record.ContractID+"_signed.pdf")

			// Check if the file exists
			if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
				// Try to get signer information for regenerating
				var signerName string
				var signerEmail string

				if recipient, err := ctrl.personRepo.GetByID(c, uuid.MustParse(record.RecipientID)); err == nil && recipient != nil {
					signerName = recipient.FullName
				} else {
					signerName = record.RecipientEmail
				}
				signerEmail = record.RecipientEmail

				// Create a basic contract data structure for regenerating signed PDF
				contractData := service.ContractPDF{
					Renter:       &model.Person{FullName: signerName},
					Owner:        nil, // Will use defaults
					Property:     nil, // Will use defaults
					Pricing:      nil, // Will use defaults
					CoSigner:     nil, // Will use defaults
					Witness:      nil, // Will use defaults
					StartDate:    time.Now(),
					EndDate:      time.Now().AddDate(0, 6, 0), // 6 months default
					CreationDate: time.Now(),
				}

				// Regenerate the signed PDF
				signedPDFData, err := service.SimpleSignPDF(
					contractData,
					signerName,
					signerEmail,
					signingId,
				)

				if err != nil {
					log.Printf("Error regenerating signed PDF: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to regenerate signed PDF"})
					return
				}

				// Save the regenerated file
				if err := os.WriteFile(pdfPath, signedPDFData, 0644); err != nil {
					log.Printf("Error writing regenerated signed PDF: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save regenerated signed PDF"})
					return
				}
			}
		} else {
			// Serve the original (unsigned) PDF
			pdfPath = filepath.Join(tempDir, record.ContractID+".pdf")

			// Check if the file exists
			if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
				// Try to get property and recipient info for the PDF
				var propertyAddress string
				var renterName string

				// Get property information if available
				property, propertyErr := ctrl.propertyRepo.GetByID(c, uuid.MustParse(record.ContractID))
				if propertyErr == nil && property != nil {
					propertyAddress = fmt.Sprintf("%s, %s, %s, %s",
						property.Address,
						property.City,
						property.State,
						property.ZipCode)
				} else {
					propertyAddress = "Dirección no disponible"
				}

				// Get renter information if available
				if recipient, err := ctrl.personRepo.GetByID(c, uuid.MustParse(record.RecipientID)); err == nil && recipient != nil {
					renterName = recipient.FullName
				} else {
					renterName = record.RecipientEmail
				}

				// Generate a simple contract PDF
				pdfData, err := service.CreateSimpleContractPDF(record.ContractID, propertyAddress, renterName)
				if err != nil {
					log.Printf("Error creating simple contract PDF: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate contract PDF"})
					return
				}

				// Ensure the directory exists
				if err := os.MkdirAll(filepath.Dir(pdfPath), 0755); err != nil {
					log.Printf("Error creating directory for PDF: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory for PDF"})
					return
				}

				// Save the PDF
				if err := os.WriteFile(pdfPath, pdfData, 0644); err != nil {
					log.Printf("Error writing contract PDF: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save contract PDF"})
					return
				}
			}
		}

		// Set content disposition for browser to display the PDF
		c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s.pdf", record.ContractID))
		c.Header("Content-Type", "application/pdf")

		// Serve the file
		c.File(pdfPath)
		return
	}

	// If no repository, generate and serve a sample PDF on-the-fly
	ctrl.createSamplePDF(c, signingId, isSigned)
}

// createSamplePDF creates and serves a sample PDF for testing or development
func (ctrl *ContractSigningController) createSamplePDF(c *gin.Context, signingId string, isSigned bool) {
	// Create a temporary file for the PDF
	tempDir := filepath.Join(os.TempDir(), "contracts")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp directory"})
		return
	}

	pdfPath := filepath.Join(tempDir, "sample_contract.pdf")

	// Create a basic PDF content
	title := "Sample Contract"
	if isSigned {
		title += " (SIGNED)"
	}

	// Create PDF content
	pdfContent := []byte{
		// PDF header
		'%', 'P', 'D', 'F', '-', '1', '.', '4', '\n',
		// Simple object structure
		'1', ' ', '0', ' ', 'o', 'b', 'j', '\n',
		'<', '<', '\n',
		'/', 'T', 'y', 'p', 'e', ' ', '/', 'C', 'a', 't', 'a', 'l', 'o', 'g', '\n',
		'/', 'P', 'a', 'g', 'e', 's', ' ', '2', ' ', '0', ' ', 'R', '\n',
		'>', '>', '\n',
		'e', 'n', 'd', 'o', 'b', 'j', '\n',
		// Pages object
		'2', ' ', '0', ' ', 'o', 'b', 'j', '\n',
		'<', '<', '\n',
		'/', 'T', 'y', 'p', 'e', ' ', '/', 'P', 'a', 'g', 'e', 's', '\n',
		'/', 'K', 'i', 'd', 's', ' ', '[', '3', ' ', '0', ' ', 'R', ']', '\n',
		'/', 'C', 'o', 'u', 'n', 't', ' ', '1', '\n',
		'>', '>', '\n',
		'e', 'n', 'd', 'o', 'b', 'j', '\n',
		// Page object
		'3', ' ', '0', ' ', 'o', 'b', 'j', '\n',
		'<', '<', '\n',
		'/', 'T', 'y', 'p', 'e', ' ', '/', 'P', 'a', 'g', 'e', '\n',
		'/', 'P', 'a', 'r', 'e', 'n', 't', ' ', '2', ' ', '0', ' ', 'R', '\n',
		'/', 'C', 'o', 'n', 't', 'e', 'n', 't', 's', ' ', '4', ' ', '0', ' ', 'R', '\n',
		'>', '>', '\n',
		'e', 'n', 'd', 'o', 'b', 'j', '\n',
		// Content stream
		'4', ' ', '0', ' ', 'o', 'b', 'j', '\n',
		'<', '<', ' ', '/', 'L', 'e', 'n', 'g', 't', 'h', ' ', '1', '0', '0', ' ', '>', '>', '\n',
		's', 't', 'r', 'e', 'a', 'm', '\n',
		'B', 'T', '\n',
		'/', 'F', '1', ' ', '1', '4', ' ', 'T', 'f', '\n',
		'1', '0', '0', ' ', '7', '0', '0', ' ', 'T', 'd', '\n',
		'(',
	}

	// Add title
	pdfContent = append(pdfContent, []byte(title)...)

	// Add contract ID and date
	pdfContent = append(pdfContent, []byte{
		')', ' ', 'T', 'j', '\n',
		'0', ' ', '-', '2', '0', ' ', 'T', 'd', '\n',
		'(', 'C', 'o', 'n', 't', 'r', 'a', 'c', 't', ' ', 'I', 'D', ':', ' ',
	}...)
	pdfContent = append(pdfContent, []byte(signingId)...)

	pdfContent = append(pdfContent, []byte{
		')', ' ', 'T', 'j', '\n',
		'0', ' ', '-', '2', '0', ' ', 'T', 'd', '\n',
		'(', 'D', 'a', 't', 'e', ':', ' ',
	}...)
	pdfContent = append(pdfContent, []byte(time.Now().Format("2006-01-02"))...)

	// Add signature if requested
	if isSigned {
		pdfContent = append(pdfContent, []byte{
			')', ' ', 'T', 'j', '\n',
			'0', ' ', '-', '2', '0', ' ', 'T', 'd', '\n',
			'(', 'S', 'i', 'g', 'n', 'e', 'd', ' ', 'o', 'n', ':', ' ',
		}...)
		pdfContent = append(pdfContent, []byte(time.Now().Format("2006-01-02 15:04:05"))...)
		pdfContent = append(pdfContent, []byte{
			')', ' ', 'T', 'j', '\n',
			'0', ' ', '-', '2', '0', ' ', 'T', 'd', '\n',
			'(', 'S', 'i', 'g', 'n', 'a', 't', 'u', 'r', 'e', ':', ' ', 'D', 'i', 'g', 'i', 't', 'a', 'l', 'l', 'y', ' ', 's', 'i', 'g', 'n', 'e', 'd', ' ', 'b', 'y', ' ', 'E', 'C', 'D', 'S', 'A', ')', ' ', 'T', 'j', '\n',
		}...)
	} else {
		pdfContent = append(pdfContent, []byte{
			')', ' ', 'T', 'j', '\n',
			'0', ' ', '-', '2', '0', ' ', 'T', 'd', '\n',
			'(', 'S', 't', 'a', 't', 'u', 's', ':', ' ', 'P', 'e', 'n', 'd', 'i', 'n', 'g', ' ', 's', 'i', 'g', 'n', 'a', 't', 'u', 'r', 'e', ')', ' ', 'T', 'j', '\n',
		}...)
	}

	// End the PDF
	pdfContent = append(pdfContent, []byte{
		'E', 'T', '\n',
		'e', 'n', 'd', 's', 't', 'r', 'e', 'a', 'm', '\n',
		'e', 'n', 'd', 'o', 'b', 'j', '\n',
		// Cross-reference table
		'x', 'r', 'e', 'f', '\n',
		'0', ' ', '5', '\n',
		'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', ' ', '6', '5', '5', '3', '5', ' ', 'f', '\n',
		'0', '0', '0', '0', '0', '0', '0', '0', '1', '0', ' ', '0', '0', '0', '0', '0', ' ', 'n', '\n',
		'0', '0', '0', '0', '0', '0', '0', '1', '6', '5', ' ', '0', '0', '0', '0', '0', ' ', 'n', '\n',
		'0', '0', '0', '0', '0', '0', '0', '2', '5', '0', ' ', '0', '0', '0', '0', '0', ' ', 'n', '\n',
		'0', '0', '0', '0', '0', '0', '0', '3', '3', '5', ' ', '0', '0', '0', '0', '0', ' ', 'n', '\n',
		// Trailer
		't', 'r', 'a', 'i', 'l', 'e', 'r', '\n',
		'<', '<', '\n',
		'/', 'S', 'i', 'z', 'e', ' ', '5', '\n',
		'/', 'R', 'o', 'o', 't', ' ', '1', ' ', '0', ' ', 'R', '\n',
		'>', '>', '\n',
		's', 't', 'a', 'r', 't', 'x', 'r', 'e', 'f', '\n',
		'4', '9', '5', '\n',
		'%', '%', 'E', 'O', 'F', '\n',
	}...)

	// Write to file
	if err := os.WriteFile(pdfPath, pdfContent, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sample PDF"})
		return
	}

	// Set content disposition for browser to display the PDF
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=sample_contract.pdf"))
	c.Header("Content-Type", "application/pdf")

	// Serve the file
	c.File(pdfPath)
}
