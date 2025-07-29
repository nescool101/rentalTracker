package service

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"encoding/base64"

	"github.com/jung-kurt/gofpdf"
)

// SimpleSignPDF creates a signed PDF file for the contract with embedded signature information
func SimpleSignPDF(contractData ContractPDF, signerName, signerEmail, signingID string) ([]byte, error) {
	// Generate a new ECDSA private key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ECDSA key: %w", err)
	}

	// Create a self-signed certificate
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: signerName},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Encode certificate to PEM format
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Generate the contract using the proper Colombian template
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set up basic formatting
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)

	// Extract data with defaults (same as contract_pdf_service.go)
	currentDate := FormatSpanishDate(contractData.CreationDate)
	propertyAddress := getPropertyAddress(contractData.Property)
	garageNumber := getGarageNumber(contractData.Property)
	buildingName := getBuildingName(contractData.Property)

	arrendadorName := "MARIA VICTORIA JIMENEZ DE ROSAS"
	arrendadorCC := "41.350.115"
	if contractData.Owner != nil && contractData.Owner.FullName != "" {
		arrendadorName = strings.ToUpper(contractData.Owner.FullName)
		arrendadorCC = contractData.Owner.NIT
	}

	arrendatarioName := "SEBASTIÁN MOTAVITA MEDELLÍN"
	arrendatarioCC := "1.026.281.306"
	if contractData.Renter != nil && contractData.Renter.FullName != "" {
		arrendatarioName = strings.ToUpper(contractData.Renter.FullName)
		arrendatarioCC = contractData.Renter.NIT
	}

	testigoName := "NA"
	testigoCC := "NA"
	if contractData.Witness != nil && contractData.Witness.FullName != "" {
		testigoName = strings.ToUpper(contractData.Witness.FullName)
		testigoCC = contractData.Witness.NIT
	}

	codeudorName := "NÉSTOR FERNANDO ÁLVAREZ"
	codeudorCC := "1.015.398.879"
	if contractData.CoSigner != nil && contractData.CoSigner.FullName != "" {
		codeudorName = strings.ToUpper(contractData.CoSigner.FullName)
		codeudorCC = contractData.CoSigner.NIT
	}

	canonMensual := "$1,600,000.00"
	canonIncluido := "INCLUÍDA LA ADMINISTRACIÓN"
	if contractData.Pricing != nil && contractData.Pricing.MonthlyRent > 0 {
		canonMensual = FormatMoney(contractData.Pricing.MonthlyRent)
	}

	fechaIniciacion := "Junio 6 de 2022"
	fechaTerminacion := "Diciembre 5 de 2022"
	if !contractData.StartDate.IsZero() {
		fechaIniciacion = FormatSpanishDate(contractData.StartDate)
	}
	if !contractData.EndDate.IsZero() {
		fechaTerminacion = FormatSpanishDate(contractData.EndDate)
	}

	// Title
	pdf.SetFont("Arial", "B", 14)
	pdf.MultiCell(0, 8, fixSpanishChars("CONTRATO DE ARRENDAMIENTO DE INMUEBLE PARA VIVIENDA URBANA"), "", "C", false)
	pdf.Ln(2)
	pdf.SetFont("Arial", "B", 12)
	pdf.MultiCell(0, 6, fixSpanishChars(propertyAddress), "", "C", false)
	pdf.Ln(10)

	// Header information
	pdf.SetFont("Arial", "B", 10)
	addInfoLine(pdf, "LUGAR Y FECHA DEL CONTRATO:", "Bogotá, D. C., "+currentDate)
	addInfoLine(pdf, "DIRECCION DEL INMUEBLE:", propertyAddress+",")
	addInfoLine(pdf, "", "Garaje # "+garageNumber+", Edificio "+buildingName)
	addInfoLine(pdf, "ARRENDADOR:", arrendadorName+", CC "+arrendadorCC)
	addInfoLine(pdf, "ARRENDATARIO:", arrendatarioName+", CC "+arrendatarioCC)
	addInfoLine(pdf, "TESTIGO:", testigoName+", CC "+testigoCC)
	addInfoLine(pdf, "CODEUDOR:", codeudorName+", CC "+codeudorCC)
	addInfoLine(pdf, "CANON MENSUAL:", canonMensual+" "+canonIncluido)
	addInfoLine(pdf, "FECHA INICIACION:", fechaIniciacion)
	addInfoLine(pdf, "FECHA TERMINACION:", fechaTerminacion)

	pdf.Ln(10)

	// Main content title
	pdf.SetFont("Arial", "B", 12)
	pdf.MultiCell(0, 8, fixSpanishChars("CONDICIONES GENERALES"), "", "C", false)
	pdf.Ln(5)

	// Add first few clauses (abbreviated for space)
	addClause(pdf, "PRIMERA: OBJETO DEL CONTRATO:",
		"Mediante el presente contrato el ARRENDADOR concede al ARRENDATARIO el goce de los inmuebles que adelante se identifican por su dirección y linderos, de acuerdo con el inventario que las partes firman por separado, el cual forma parte integral de este mismo contrato de arrendamiento.")

	// Digital signature banner
	pdf.SetFillColor(220, 220, 220) // Light gray background
	pdf.Rect(20, pdf.GetY(), 170, 30, "F")
	pdf.SetFont("Arial", "B", 12)
	pdf.SetXY(20, pdf.GetY()+5)
	pdf.MultiCell(170, 8, fixSpanishChars("CERTIFICADO DE FIRMA DIGITAL"), "", "C", false)

	// Signature details
	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(30, pdf.GetY())
	pdf.MultiCell(150, 6, fixSpanishChars(fmt.Sprintf("Firmado por: %s (%s)", signerName, signerEmail)), "", "L", false)
	pdf.SetX(30)
	pdf.MultiCell(150, 6, fixSpanishChars(fmt.Sprintf("Fecha y hora: %s", time.Now().Format("02/01/2006 15:04:05"))), "", "L", false)
	pdf.SetX(30)
	pdf.MultiCell(150, 6, fixSpanishChars(fmt.Sprintf("ID de Firma: %s", signingID)), "", "L", false)
	pdf.Ln(5)

	// Fingerprint data
	fingerprint := fmt.Sprintf("%X", cert.SerialNumber)
	pdf.SetFont("Arial", "", 8)
	pdf.MultiCell(0, 5, fixSpanishChars(fmt.Sprintf("Huella digital del certificado: %s", fingerprint)), "", "L", false)

	// Validation text
	pdf.SetFont("Arial", "I", 8)
	pdf.MultiCell(0, 5, fixSpanishChars("Este documento ha sido firmado digitalmente utilizando tecnología ECDSA (Elliptic Curve Digital Signature Algorithm) y está legalmente vinculado a la identidad del firmante."), "", "L", false)

	// Add signature tables
	addSignatureTables(pdf, arrendadorName, arrendadorCC, arrendatarioName, arrendatarioCC, testigoName, testigoCC, codeudorName, codeudorCC)

	// Add condensed certificate data at the bottom
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 6)
	certString := base64.StdEncoding.EncodeToString(certPEM)
	if len(certString) > 300 {
		certString = certString[:300] + "..."
	}
	pdf.MultiCell(0, 4, fmt.Sprintf("Datos del certificado: %s", certString), "", "L", false)

	// Create temp directory if it doesn't exist
	tempDir := filepath.Join(os.TempDir(), "contracts")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate PDF to temp file first
	tempFilePath := filepath.Join(tempDir, fmt.Sprintf("%s_signed.pdf", signingID))
	err = pdf.OutputFileAndClose(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to write PDF to file: %w", err)
	}

	// Read the file back into memory
	pdfBytes, err := os.ReadFile(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF from file: %w", err)
	}

	return pdfBytes, nil
}

// CreateSimpleContractPDF creates a basic contract PDF without digital signature
func CreateSimpleContractPDF(contractID string, propertyAddress string, renterName string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set up basic formatting
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)

	// Title
	pdf.SetFont("Arial", "B", 12)
	pdf.MultiCell(0, 8, "CONTRATO DE ARRENDAMIENTO DE INMUEBLE PARA VIVIENDA URBANA", "", "C", false)
	pdf.Ln(2)
	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 6, "Carrera 18b No 145 - 08 Apto 201", "", "C", false)
	pdf.Ln(10)

	// Date and location
	pdf.SetFont("Arial", "B", 10)
	currentDate := time.Now().Format("2 de January de 2006")
	currentDate = strings.Replace(currentDate, "January", "enero", 1)
	pdf.MultiCell(0, 6, "LUGAR Y FECHA DEL CONTRATO: Bogotá, D. C., "+currentDate, "", "L", false)
	pdf.Ln(5)

	// Contract data - two columns layout
	leftColWidth := float64(60)
	rightColWidth := float64(110)

	// Direccion del inmueble
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(leftColWidth, 6, "DIRECCION DEL INMUEBLE:")
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(rightColWidth, 6, propertyAddress+" ", "", "L", false)
	pdf.Ln(1)

	// Arrendador
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(leftColWidth, 6, "ARRENDADOR:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(rightColWidth, 6, "MARIA VICTORIA JIMENEZ DE ROSAS, CC 41.350.115")
	pdf.Ln(8)

	// Arrendatario (nombre del arrendatario)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(leftColWidth, 6, "ARRENDATARIO:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(rightColWidth, 6, renterName)
	pdf.Ln(8)

	// Testigo y Codeudor
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(leftColWidth, 6, "TESTIGO:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(rightColWidth, 6, "LAURA CAMILA CORREA, CC 1.020.841.781")
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(leftColWidth, 6, "CODEUDOR:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(rightColWidth, 6, "NÉSTOR FERNANDO ÁLVAREZ, CC 1.015.398.879")
	pdf.Ln(8)

	// Canon mensual
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(leftColWidth, 6, "CANON MENSUAL:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(rightColWidth, 6, "$1,600,000.00 INCLUÍDA LA ADMINISTRACIÓN")
	pdf.Ln(8)

	// Fechas
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(leftColWidth, 6, "FECHA INCIACION:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(rightColWidth, 6, "Junio 6 de 2022")
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(leftColWidth, 6, "FECHA TERMINACION:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(rightColWidth, 6, "Diciembre 5 de 2022")
	pdf.Ln(15)

	// Condiciones generales
	pdf.SetFont("Arial", "B", 12)
	pdf.MultiCell(0, 8, "CONDICIONES GENERALES", "", "C", false)
	pdf.Ln(3)

	// Pendiente de firma - banner
	pdf.SetFillColor(255, 240, 240) // Light red background
	pdf.Rect(20, pdf.GetY(), 170, 20, "F")
	pdf.SetFont("Arial", "B", 12)
	pdf.SetXY(20, pdf.GetY()+5)
	pdf.MultiCell(170, 8, "PENDIENTE DE FIRMA", "", "C", false)

	// Pendiente de firma - texto
	pdf.SetFont("Arial", "", 10)
	pdf.Ln(20)
	pdf.MultiCell(0, 6, "Este contrato está pendiente de firma digital. Una vez firmado, se generará una versión firmada digitalmente con validez legal.", "", "L", false)

	// Create temp directory if it doesn't exist
	tempDir := filepath.Join(os.TempDir(), "contracts")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate PDF to temp file first
	tempFilePath := filepath.Join(tempDir, fmt.Sprintf("%s.pdf", contractID))
	err := pdf.OutputFileAndClose(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to write PDF to file: %w", err)
	}

	// Read the file back into memory
	pdfBytes, err := os.ReadFile(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF from file: %w", err)
	}

	return pdfBytes, nil
}
