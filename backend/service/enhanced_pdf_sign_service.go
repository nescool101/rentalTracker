package service

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
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
)

// ECDSASignatureMetadata holds additional information to include in the signature
type ECDSASignatureMetadata struct {
	SignID     string // ID of the signature request
	SignedBy   string // Email of the person who signed
	TimeSigned string // Timestamp when signed
	Location   string // Location information
	Reason     string // Reason for signing
	Contact    string // Contact information
}

// GenerateECDSACertificate generates a new ECDSA certificate for signing PDFs
func GenerateECDSACertificate(outputDir string) (*ecdsa.PrivateKey, *x509.Certificate, error) {
	// Ensure directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create directory: %w", err)
	}

	keyPath := filepath.Join(outputDir, "ecdsa_private.key")
	certPath := filepath.Join(outputDir, "ecdsa_certificate.crt")

	// Generate ECDSA private key (P-384 curve offers good security)
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate ECDSA key: %w", err)
	}

	// Prepare certificate template
	notBefore := time.Now()
	notAfter := notBefore.AddDate(1, 0, 0) // Valid for 1 year

	// Generate a random serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"Rental Management System"},
			OrganizationalUnit: []string{"PDF Signing Department"},
			Country:            []string{"US"},
			Province:           []string{"State"},
			Locality:           []string{"City"},
			CommonName:         "RMS ECDSA PDF Signer",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create the certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Save certificate
	certOut, err := os.Create(certPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certOut.Close()

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode certificate: %w", err)
	}

	// Save private key
	keyOut, err := os.Create(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create private key file: %w", err)
	}
	defer keyOut.Close()

	// Marshal the private key to PKCS8
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	err = pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode private key: %w", err)
	}

	log.Printf("Created ECDSA certificate at %s and private key at %s", certPath, keyPath)
	return privateKey, cert, nil
}

// LoadOrGenerateECDSACertificate loads existing certificate or generates a new one
func LoadOrGenerateECDSACertificate(certsDir string) (*ecdsa.PrivateKey, *x509.Certificate, error) {
	keyPath := filepath.Join(certsDir, "ecdsa_private.key")
	certPath := filepath.Join(certsDir, "ecdsa_certificate.crt")

	// Check if files exist
	keyExists := true
	certExists := true
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		keyExists = false
	}
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		certExists = false
	}

	if !keyExists || !certExists {
		log.Printf("ECDSA certificate or key not found, generating new ones in: %s", certsDir)
		return GenerateECDSACertificate(certsDir)
	}

	// Files exist, load them
	log.Printf("Loading existing ECDSA certificate and key from: %s", certsDir)

	// Load certificate
	certPEM, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		return nil, nil, fmt.Errorf("failed to decode PEM block containing certificate")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Load private key
	keyPEM, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil || keyBlock.Type != "PRIVATE KEY" {
		return nil, nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	privateKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("private key is not an ECDSA key")
	}

	log.Printf("Successfully loaded ECDSA certificate and key")
	return privateKey, cert, nil
}

// SignPDFWithECDSA signs a PDF using ECDSA certificate
func SignPDFWithECDSA(inputPDFPath, outputPDFPath string, metadata *ECDSASignatureMetadata) error {
	// Check if input file exists
	if _, err := os.Stat(inputPDFPath); os.IsNotExist(err) {
		return fmt.Errorf("input PDF file does not exist: %s", inputPDFPath)
	}

	// Ensure the output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPDFPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get or generate certificates
	certsDir := "./certs"
	if err := os.MkdirAll(certsDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificates directory: %w", err)
	}

	privateKey, cert, err := LoadOrGenerateECDSACertificate(certsDir)
	if err != nil {
		return fmt.Errorf("failed to load or generate certificate: %w", err)
	}

	// Open input PDF file
	inFile, err := os.Open(inputPDFPath)
	if err != nil {
		return fmt.Errorf("failed to open input PDF: %w", err)
	}
	defer inFile.Close()

	// Create output PDF file
	outFile, err := os.Create(outputPDFPath)
	if err != nil {
		return fmt.Errorf("failed to create output PDF: %w", err)
	}
	defer outFile.Close()

	// Get file info
	fileInfo, err := inFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	size := fileInfo.Size()

	// Create PDF reader
	reader, err := pdf.NewReader(inFile, size)
	if err != nil {
		return fmt.Errorf("failed to create PDF reader: %w", err)
	}

	// If metadata is nil, create default metadata
	if metadata == nil {
		metadata = &ECDSASignatureMetadata{
			SignID:     "unknown",
			SignedBy:   "unknown",
			TimeSigned: time.Now().Format(time.RFC3339),
			Location:   "Digital Signature",
			Reason:     "Contract signing",
			Contact:    "RentalFullNescao System",
		}
	}

	// Create contact info with metadata
	contactInfo := fmt.Sprintf("SignID: %s | SignedBy: %s | TimeSigned: %s",
		metadata.SignID, metadata.SignedBy, metadata.TimeSigned)

	// Try to get certificate chains (but proceed even if we can't)
	certChains, err := cert.Verify(x509.VerifyOptions{})
	if err != nil {
		log.Printf("Warning: Certificate verification error: %v", err)
		certChains = [][]*x509.Certificate{}
	}

	// Sign the PDF
	err = sign.Sign(inFile, outFile, reader, size, sign.SignData{
		Signature: sign.SignDataSignature{
			Info: sign.SignDataSignatureInfo{
				Name:        metadata.SignedBy,
				Location:    metadata.Location,
				Reason:      metadata.Reason,
				ContactInfo: contactInfo,
				Date:        time.Now(),
			},
			CertType:   sign.CertificationSignature,
			DocMDPPerm: sign.AllowFillingExistingFormFieldsAndSignaturesPerms,
		},
		Signer:            privateKey,
		DigestAlgorithm:   crypto.SHA256,
		Certificate:       cert,
		CertificateChains: certChains,
	})
	if err != nil {
		return fmt.Errorf("failed to sign PDF: %w", err)
	}

	log.Printf("PDF signed successfully with ECDSA certificate. Output: %s", outputPDFPath)
	return nil
}

// SignExistingPDFWithECDSA reads an existing PDF and signs it with ECDSA certificate
func SignExistingPDFWithECDSA(contractData ContractPDF, signerName, signerEmail string, signingID string) ([]byte, error) {
	// Create temp directory if it doesn't exist
	tempDir := filepath.Join(os.TempDir(), "contracts")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Define file paths
	inputPDFPath := filepath.Join(tempDir, signingID+"_input.pdf")
	outputPDFPath := filepath.Join(tempDir, signingID+"_signed.pdf")

	// Generate the contract PDF using the proper Colombian template first
	pdfData, err := GenerateContractPDF(contractData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate contract PDF: %w", err)
	}

	// Write the PDF data to the temp file
	if err := ioutil.WriteFile(inputPDFPath, pdfData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write PDF file: %w", err)
	}

	// Create metadata for the signature
	metadata := &ECDSASignatureMetadata{
		SignID:     signingID,
		SignedBy:   signerEmail,
		TimeSigned: time.Now().Format(time.RFC3339),
		Location:   "Digital Signature",
		Reason:     "Contract signing",
		Contact:    "RentalFullNescao System",
	}

	// Sign the PDF
	if err := SignPDFWithECDSA(inputPDFPath, outputPDFPath, metadata); err != nil {
		return nil, fmt.Errorf("failed to sign PDF: %w", err)
	}

	// Read the signed PDF
	signedPDFData, err := ioutil.ReadFile(outputPDFPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read signed PDF: %w", err)
	}

	return signedPDFData, nil
}

// Helper function to create a basic PDF - same as the existing one in controller
func createProperPDF(outputPath string, contractID string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create a minimal valid PDF file with contract info
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
	if err := ioutil.WriteFile(outputPath, pdfContent, 0644); err != nil {
		return fmt.Errorf("failed to write PDF file: %w", err)
	}

	log.Printf("Created proper PDF at %s for contract ID %s", outputPath, contractID)
	return nil
}
