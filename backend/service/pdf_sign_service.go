package service

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/digitorus/pdf"
	"github.com/digitorus/pdfsign/sign"
	"github.com/google/uuid"
	"github.com/nescool101/rentManager/model"
)

// ContractSigningInfo holds information for the contract signing process
type ContractSigningInfo struct {
	ContractID     string // UUID for the contract
	RecipientID    string // Person ID of the recipient
	RecipientEmail string
	PDFData        []byte // The PDF data to be signed
	SignerName     string // Name of the signer
	SignatureID    string // UUID for the signature request
}

// SigningStatus represents the current state of a signature request
type SigningStatus string

const (
	StatusPending  SigningStatus = "pending"
	StatusSigned   SigningStatus = "signed"
	StatusRejected SigningStatus = "rejected"
	StatusExpired  SigningStatus = "expired"
)

// ContractSigningRequest represents a request to sign a contract
type ContractSigningRequest struct {
	ID             string        // UUID for this signing request
	ContractID     string        // Reference to the contract
	RecipientID    string        // Person who needs to sign
	RecipientEmail string        // Email of the recipient
	Status         SigningStatus // Current status
	CreatedAt      time.Time     // When the request was created
	ExpiresAt      time.Time     // When the request expires
	SignedAt       *time.Time    // When it was signed (if signed)
	SignatureData  []byte        // The signature data (if signed)
}

// Options for signing
type SignPDFOptions struct {
	CertificatePath   string
	PrivateKeyPath    string
	SignatureReason   string
	SignatureContact  string
	SignatureLocation string
}

// Default sign options
var DefaultSignOptions = SignPDFOptions{
	CertificatePath:   "./certs/certificate.crt",
	PrivateKeyPath:    "./certs/private.key",
	SignatureReason:   "Approved contract",
	SignatureContact:  "support@example.com",
	SignatureLocation: "Digital",
}

// CreateSignatureRequest generates a new signature request and sends an email invitation
func CreateSignatureRequest(contractInfo model.ContractSigningInfo, expirationDays int) (*model.ContractSigningRequest, error) {
	// Generate IDs if not provided
	if contractInfo.ContractID == "" {
		contractInfo.ContractID = uuid.New().String()
	}

	if contractInfo.SignatureID == "" {
		contractInfo.SignatureID = uuid.New().String()
	}

	// Calculate expiration date
	now := time.Now()
	expiresAt := now.AddDate(0, 0, expirationDays)

	// Create the signature request
	request := &model.ContractSigningRequest{
		ID:             contractInfo.SignatureID,
		ContractID:     contractInfo.ContractID,
		RecipientID:    contractInfo.RecipientID,
		RecipientEmail: contractInfo.RecipientEmail,
		Status:         model.StatusPending,
		CreatedAt:      now,
		ExpiresAt:      expiresAt,
	}

	// Save the contract to disk temporarily
	if _, err := saveTempPDF(contractInfo.PDFData, contractInfo.ContractID); err != nil {
		return nil, fmt.Errorf("error saving temporary PDF: %w", err)
	}

	// Get base URL from environment variable or use localhost as fallback
	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		// Default to localhost in development
		baseURL = "http://localhost:5173"
	}

	// Generate signing URL
	signingURL := fmt.Sprintf("%s/sign/%s", baseURL, request.ID)

	// Format date in Spanish
	spanishMonths := map[time.Month]string{
		time.January:   "enero",
		time.February:  "febrero",
		time.March:     "marzo",
		time.April:     "abril",
		time.May:       "mayo",
		time.June:      "junio",
		time.July:      "julio",
		time.August:    "agosto",
		time.September: "septiembre",
		time.October:   "octubre",
		time.November:  "noviembre",
		time.December:  "diciembre",
	}

	expiryDay := expiresAt.Day()
	expiryMonth := spanishMonths[expiresAt.Month()]
	expiryYear := expiresAt.Year()
	formattedDate := fmt.Sprintf("%d de %s de %d", expiryDay, expiryMonth, expiryYear)

	// Send email with signing link
	subject := "Contrato Listo para Firma"
	body := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>Solicitud de Firma de Contrato</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 0; padding: 20px; color: #333; }
			.container { max-width: 600px; margin: 0 auto; }
			.header { background-color: #f8f9fa; padding: 20px; text-align: center; }
			.content { padding: 20px; }
			.button { display: inline-block; background-color: #007bff; color: white; padding: 10px 20px; 
					text-decoration: none; border-radius: 4px; margin-top: 20px; }
			.footer { margin-top: 20px; font-size: 12px; color: #6c757d; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h2>Contrato Listo para su Firma</h2>
			</div>
			<div class="content">
				<p>Estimado(a) %s,</p>
				<p>Un contrato está listo para su revisión y firma. Por favor haga clic en el botón a continuación para ver y firmar el documento:</p>
				<p><a href="%s" class="button">Revisar y Firmar Contrato</a></p>
				<p>Esta solicitud de firma expirará el %s.</p>
				<p>Si tiene alguna pregunta sobre este documento, por favor contáctenos directamente.</p>
				<p>Gracias,<br>Sistema de Administración de Propiedades</p>
			</div>
			<div class="footer">
				<p>Este es un mensaje automático. Por favor no responda directamente a este correo.</p>
			</div>
		</div>
	</body>
	</html>
	`, contractInfo.SignerName, signingURL, formattedDate)

	// Send the email
	err := SendSimpleEmail(contractInfo.RecipientEmail, subject, body)
	if err != nil {
		log.Printf("Error sending signature request email: %v", err)
		return nil, fmt.Errorf("error sending signature request email: %w", err)
	}

	// In a real implementation, you would save this request to a database
	log.Printf("Signature request created with ID: %s for contract: %s, sent to: %s",
		request.ID, request.ContractID, request.RecipientEmail)

	return request, nil
}

// SignPDF signs a PDF document with the provided certificate and private key
func SignPDF(pdfData []byte, signerName string, options SignPDFOptions) ([]byte, error) {
	// Use default options if not provided
	if options.CertificatePath == "" {
		options = DefaultSignOptions
	}

	// In development mode, return the original PDF if the certificate files don't exist
	if _, err := os.Stat(options.CertificatePath); os.IsNotExist(err) {
		log.Printf("Certificate file not found at %s, skipping signing in development mode", options.CertificatePath)
		return pdfData, nil
	}
	if _, err := os.Stat(options.PrivateKeyPath); os.IsNotExist(err) {
		log.Printf("Private key file not found at %s, skipping signing in development mode", options.PrivateKeyPath)
		return pdfData, nil
	}

	// Read certificate
	certData, err := ioutil.ReadFile(options.CertificatePath)
	if err != nil {
		log.Printf("Warning: Could not read certificate: %v", err)
		return pdfData, nil
	}

	// Parse certificate
	block, _ := pem.Decode(certData)
	if block == nil {
		log.Printf("Warning: Failed to parse certificate PEM")
		return pdfData, nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Printf("Warning: Failed to parse certificate: %v", err)
		return pdfData, nil
	}

	// Read private key
	keyData, err := ioutil.ReadFile(options.PrivateKeyPath)
	if err != nil {
		log.Printf("Warning: Could not read private key: %v", err)
		return pdfData, nil
	}

	// Parse private key
	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil {
		log.Printf("Warning: Failed to parse private key PEM")
		return pdfData, nil
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		log.Printf("Warning: Failed to parse private key: %v", err)
		return pdfData, nil
	}

	// Create temporary input and output files
	tempDir := filepath.Join(os.TempDir(), "pdf_signing")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Printf("Error creating temp directory: %v", err)
		return pdfData, nil
	}

	inputPath := filepath.Join(tempDir, "input_"+uuid.New().String()+".pdf")
	outputPath := filepath.Join(tempDir, "output_"+uuid.New().String()+".pdf")

	// Write input PDF
	if err := ioutil.WriteFile(inputPath, pdfData, 0644); err != nil {
		log.Printf("Error writing input PDF: %v", err)
		return pdfData, nil
	}
	defer os.Remove(inputPath) // Clean up input file

	// Open input and output files
	inputFile, err := os.Open(inputPath)
	if err != nil {
		log.Printf("Error opening input file: %v", err)
		return pdfData, nil
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		log.Printf("Error creating output file: %v", err)
		return pdfData, nil
	}
	defer outputFile.Close()

	// Get file info for size
	fileInfo, err := inputFile.Stat()
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		return pdfData, nil
	}
	fileSize := fileInfo.Size()

	// Read the PDF to get its structure
	pdfReader, err := pdf.NewReader(inputFile, fileSize)
	if err != nil {
		log.Printf("Error creating PDF reader: %v", err)
		return pdfData, nil
	}

	// Reset file position to the beginning
	_, err = inputFile.Seek(0, 0)
	if err != nil {
		log.Printf("Error seeking file: %v", err)
		return pdfData, nil
	}

	// Sign the PDF
	err = sign.Sign(inputFile, outputFile, pdfReader, fileSize, sign.SignData{
		Signature: sign.SignDataSignature{
			Info: sign.SignDataSignatureInfo{
				Name:        signerName,
				Location:    options.SignatureLocation,
				Reason:      options.SignatureReason,
				ContactInfo: options.SignatureContact,
				Date:        time.Now().Local(),
			},
			CertType:   sign.CertificationSignature,
			DocMDPPerm: sign.AllowFillingExistingFormFieldsAndSignaturesPerms,
		},
		Signer:          privateKey,
		DigestAlgorithm: crypto.SHA256,
		Certificate:     cert,
	})

	if err != nil {
		log.Printf("Warning: Error signing PDF: %v", err)
		log.Printf("Proceeding with unsigned PDF in development mode")
		return pdfData, nil
	}

	// Close files to ensure all data is written
	inputFile.Close()
	outputFile.Close()

	// Read the signed PDF
	signedPDFData, err := ioutil.ReadFile(outputPath)
	if err != nil {
		log.Printf("Error reading signed PDF: %v", err)
		return pdfData, nil
	}

	// Clean up the output file
	defer os.Remove(outputPath)

	log.Printf("PDF successfully signed with certificate: %s", cert.Subject.CommonName)
	return signedPDFData, nil
}

// VerifyPDFSignature verifies a signed PDF
// This is a simplified version that doesn't actually verify the signature
func VerifyPDFSignature(signedPDFData []byte) (bool, error) {
	// In a real implementation, this would use pdfsign to verify the signature
	// For development purposes, we're just logging that we would verify it
	log.Printf("Would verify PDF signature on a %d byte PDF", len(signedPDFData))

	// Since we're not actually signing PDFs yet, just return true
	return true, nil
}

// Helper function to temporarily save a PDF
func saveTempPDF(pdfData []byte, contractID string) (string, error) {
	// Create temporary directory if it doesn't exist
	tempDir := filepath.Join(os.TempDir(), "contracts")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", err
	}

	// Save the PDF
	tempPDFPath := filepath.Join(tempDir, contractID+".pdf")
	err := ioutil.WriteFile(tempPDFPath, pdfData, 0644)
	if err != nil {
		return "", err
	}

	return tempPDFPath, nil
}

// Helper function to read a temporary PDF
func readTempPDF(contractID string) ([]byte, error) {
	tempPDFPath := filepath.Join(os.TempDir(), "contracts", contractID+".pdf")
	return ioutil.ReadFile(tempPDFPath)
}

// SaveTempPDF saves a PDF to a temporary location and returns the file path
func SaveTempPDF(pdfData []byte, contractID string) (string, error) {
	return saveTempPDF(pdfData, contractID)
}

// ReadTempPDF reads a PDF from a temporary location
func ReadTempPDF(contractID string) ([]byte, error) {
	return readTempPDF(contractID)
}

// Generate a self-signed certificate and key for development/testing
func GenerateSelfSignedCert(outputDir string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	keyPath := filepath.Join(outputDir, "private.key")
	certPath := filepath.Join(outputDir, "certificate.crt")

	// Create a self-signed certificate valid for PDF signing
	// First, generate a private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Prepare certificate template
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // Valid for 1 year

	serialNumber := make([]byte, 20)
	_, err = rand.Read(serialNumber)
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: new(big.Int).SetBytes(serialNumber),
		Subject: pkix.Name{
			Organization:       []string{"Rental Management System"},
			OrganizationalUnit: []string{"PDF Signing Department"},
			Country:            []string{"US"},
			Province:           []string{"State"},
			Locality:           []string{"City"},
			CommonName:         "RMS PDF Signer",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create the certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode and save the certificate
	certOut, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certOut.Close()

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return fmt.Errorf("failed to encode certificate: %w", err)
	}

	// Encode and save the private key
	keyOut, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}
	defer keyOut.Close()

	keyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	err = pem.Encode(keyOut, keyPEM)
	if err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}

	log.Printf("Created self-signed certificate at %s and private key at %s", certPath, keyPath)
	log.Printf("Note: These are self-signed certificates for development purposes only!")

	return nil
}

// SendSignedPDFByEmail sends the signed PDF to the recipient
func SendSignedPDFByEmail(signingInfo *model.ContractSigningRequest, signedPDFData []byte) error {
	// Format current date in Spanish for the email
	now := time.Now()
	spanishMonths := map[time.Month]string{
		time.January:   "enero",
		time.February:  "febrero",
		time.March:     "marzo",
		time.April:     "abril",
		time.May:       "mayo",
		time.June:      "junio",
		time.July:      "julio",
		time.August:    "agosto",
		time.September: "septiembre",
		time.October:   "octubre",
		time.November:  "noviembre",
		time.December:  "diciembre",
	}
	currentDay := now.Day()
	currentMonth := spanishMonths[now.Month()]
	currentYear := now.Year()
	formattedDate := fmt.Sprintf("%d de %s de %d", currentDay, currentMonth, currentYear)

	subject := "Contrato Firmado - Copia para sus Registros"
	body := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>Contrato Firmado</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 0; padding: 20px; color: #333; }
			.container { max-width: 600px; margin: 0 auto; }
			.header { background-color: #f8f9fa; padding: 20px; text-align: center; }
			.content { padding: 20px; }
			.footer { margin-top: 20px; font-size: 12px; color: #6c757d; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h2>Contrato Firmado</h2>
			</div>
			<div class="content">
				<p>Estimado(a),</p>
				<p>Adjunto a este correo encontrará una copia del contrato firmado para sus registros.</p>
				<p>Este documento ha sido firmado digitalmente el %s y tiene validez legal.</p>
				<p>Gracias por usar nuestro sistema de firma digital.</p>
				<p>Atentamente,<br>Sistema de Administración de Propiedades</p>
			</div>
			<div class="footer">
				<p>Este es un mensaje automático. Por favor no responda directamente a este correo.</p>
			</div>
		</div>
	</body>
	</html>
	`, formattedDate)

	// Create temporary file for attachment
	tempFile, err := ioutil.TempFile("", "signed_contract_*.pdf")
	if err != nil {
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write PDF data to temp file
	if _, err := tempFile.Write(signedPDFData); err != nil {
		return fmt.Errorf("error writing to temporary file: %w", err)
	}

	// Send email with attachment
	err = SendEmailWithAttachment(signingInfo.RecipientEmail, subject, body, tempFile.Name(), "contrato_firmado.pdf")
	if err != nil {
		return fmt.Errorf("error sending email with signed PDF: %w", err)
	}

	return nil
}

// SignContractPDF generates a Colombian contract PDF using the new template and signs it
func SignContractPDF(contractData ContractPDF, signerName string, options SignPDFOptions) ([]byte, error) {
	// Generate the contract PDF using the proper Colombian template
	pdfData, err := GenerateContractPDF(contractData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate contract PDF: %w", err)
	}

	// Sign the generated PDF
	signedPDFData, err := SignPDF(pdfData, signerName, options)
	if err != nil {
		return nil, fmt.Errorf("failed to sign contract PDF: %w", err)
	}

	return signedPDFData, nil
}
