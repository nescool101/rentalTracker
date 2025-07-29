package service

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"strings"
)

// ProtonMailConfig holds the configuration for ProtonMail SMTP service
type ProtonMailConfig struct {
	Username string
	Password string
	Host     string
	Port     int
	FromName string
}

// DefaultProtonMailConfig is the default ProtonMail configuration
var DefaultProtonMailConfig = ProtonMailConfig{
	// These values will be replaced with your actual ProtonMail credentials
	Username: "your.email@protonmail.com", // Replace with your ProtonMail email
	Password: "your-password-here",        // Replace with your ProtonMail password
	Host:     "smtp.protonmail.ch",
	Port:     587,
	FromName: "Rental Management System",
}

// SendProtonMailEmail sends an email using ProtonMail SMTP server
func SendProtonMailEmail(to, subject, htmlBody string) error {
	// Use the default ProtonMail configuration
	return SendProtonMailEmailWithConfig(to, subject, htmlBody, DefaultProtonMailConfig)
}

// SendProtonMailEmailWithConfig sends an email using ProtonMail SMTP with custom configuration
func SendProtonMailEmailWithConfig(to, subject, htmlBody string, config ProtonMailConfig) error {
	// Prepare email headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", config.FromName, config.Username)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// Build the message
	var message string
	for key, value := range headers {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + htmlBody

	// Set up authentication
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	// Connect to the server, set up TLS
	serverAddr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// Initialize TLS config
	tlsConfig := &tls.Config{
		ServerName: config.Host,
	}

	// Connect to the SMTP server
	client, err := smtp.Dial(serverAddr)
	if err != nil {
		log.Printf("❌ [EMAIL DIAL ERROR] %s - Error: %v", to, err)
		return err
	}
	defer client.Close()

	// Start TLS
	if err = client.StartTLS(tlsConfig); err != nil {
		log.Printf("❌ [EMAIL TLS ERROR] %s - Error: %v", to, err)
		return err
	}

	// Authenticate
	if err = client.Auth(auth); err != nil {
		log.Printf("❌ [EMAIL AUTH ERROR] %s - Error: %v", to, err)
		return err
	}

	// Set the sender and recipient
	if err = client.Mail(config.Username); err != nil {
		log.Printf("❌ [EMAIL SENDER ERROR] %s - Error: %v", to, err)
		return err
	}

	// Add recipients
	for _, recipient := range strings.Split(to, ",") {
		recipient = strings.TrimSpace(recipient)
		if err = client.Rcpt(recipient); err != nil {
			log.Printf("❌ [EMAIL RECIPIENT ERROR] %s - Error: %v", recipient, err)
			return err
		}
	}

	// Send the email body
	w, err := client.Data()
	if err != nil {
		log.Printf("❌ [EMAIL DATA ERROR] %s - Error: %v", to, err)
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		log.Printf("❌ [EMAIL WRITE ERROR] %s - Error: %v", to, err)
		return err
	}

	err = w.Close()
	if err != nil {
		log.Printf("❌ [EMAIL CLOSE ERROR] %s - Error: %v", to, err)
		return err
	}

	// Send the QUIT command and close the connection
	err = client.Quit()
	if err != nil {
		log.Printf("❌ [EMAIL QUIT ERROR] %s - Error: %v", to, err)
		return err
	}

	log.Printf("✅ [EMAIL SENT] %s", to)
	return nil
}

// SendEmailWithAttachment sends an email with a file attachment
func SendEmailWithAttachment(to, subject, htmlBody, attachmentPath, attachmentName string) error {
	// Use the default ProtonMail configuration
	return SendEmailWithAttachmentAndConfig(to, subject, htmlBody, attachmentPath, attachmentName, DefaultProtonMailConfig)
}

// SendEmailWithAttachmentAndConfig sends an email with a file attachment using custom config
func SendEmailWithAttachmentAndConfig(to, subject, htmlBody, attachmentPath, attachmentName string, config ProtonMailConfig) error {
	// Read the attachment file
	attachmentData, err := ioutil.ReadFile(attachmentPath)
	if err != nil {
		return fmt.Errorf("failed to read attachment file: %w", err)
	}

	// Create a unique boundary for MIME parts
	boundary := "==BOUNDARY_FOR_EMAIL_WITH_ATTACHMENT=="

	// Set up headers
	from := fmt.Sprintf("%s <%s>", config.FromName, config.Username)
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = fmt.Sprintf("multipart/mixed; boundary=%s", boundary)

	// Start building the message
	var message bytes.Buffer
	for key, value := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	message.WriteString("\r\n")

	// Add HTML part
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n\r\n")

	// Add attachment part
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString(fmt.Sprintf("Content-Type: application/pdf; name=\"%s\"\r\n", attachmentName))
	message.WriteString("Content-Transfer-Encoding: base64\r\n")
	message.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", attachmentName))

	// Encode attachment data in base64
	encodedData := base64.StdEncoding.EncodeToString(attachmentData)
	// Split base64 data into lines of 76 characters as per RFC
	for i := 0; i < len(encodedData); i += 76 {
		end := i + 76
		if end > len(encodedData) {
			end = len(encodedData)
		}
		message.WriteString(encodedData[i:end] + "\r\n")
	}

	// Close boundary
	message.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	// Set up authentication
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	// Connect to the server, set up TLS
	serverAddr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// Initialize TLS config
	tlsConfig := &tls.Config{
		ServerName: config.Host,
	}

	// Connect to the SMTP server
	client, err := smtp.Dial(serverAddr)
	if err != nil {
		log.Printf("❌ [EMAIL DIAL ERROR] %s - Error: %v", to, err)
		return err
	}
	defer client.Close()

	// Start TLS
	if err = client.StartTLS(tlsConfig); err != nil {
		log.Printf("❌ [EMAIL TLS ERROR] %s - Error: %v", to, err)
		return err
	}

	// Authenticate
	if err = client.Auth(auth); err != nil {
		log.Printf("❌ [EMAIL AUTH ERROR] %s - Error: %v", to, err)
		return err
	}

	// Set the sender and recipient
	if err = client.Mail(config.Username); err != nil {
		log.Printf("❌ [EMAIL SENDER ERROR] %s - Error: %v", to, err)
		return err
	}

	// Add recipients
	for _, recipient := range strings.Split(to, ",") {
		recipient = strings.TrimSpace(recipient)
		if err = client.Rcpt(recipient); err != nil {
			log.Printf("❌ [EMAIL RECIPIENT ERROR] %s - Error: %v", recipient, err)
			return err
		}
	}

	// Send the email body
	w, err := client.Data()
	if err != nil {
		log.Printf("❌ [EMAIL DATA ERROR] %s - Error: %v", to, err)
		return err
	}

	_, err = w.Write(message.Bytes())
	if err != nil {
		log.Printf("❌ [EMAIL WRITE ERROR] %s - Error: %v", to, err)
		return err
	}

	err = w.Close()
	if err != nil {
		log.Printf("❌ [EMAIL CLOSE ERROR] %s - Error: %v", to, err)
		return err
	}

	// Send the QUIT command and close the connection
	err = client.Quit()
	if err != nil {
		log.Printf("❌ [EMAIL QUIT ERROR] %s - Error: %v", to, err)
		return err
	}

	log.Printf("✅ [EMAIL WITH ATTACHMENT SENT] %s - %s", to, attachmentName)
	return nil
}

// SendGmailEmail sends an email using Gmail SMTP server
func SendGmailEmail(to, subject, htmlBody string) error {
	return SendProtonMailEmail(to, subject, htmlBody)
}

// SendSimpleEmail is a wrapper for backward compatibility
// This now uses Gmail SMTP instead of Resend
func SendSimpleEmail(to, subject, htmlBody string) error {
	return SendProtonMailEmail(to, subject, htmlBody)
}

// SendUploadLinkEmail envía un email con enlace de subida de archivos
func SendUploadLinkEmail(to, name, token string) error {
	subject := "📁 Enlace para Subir Archivos - Rental Manager"

	uploadURL := fmt.Sprintf("http://localhost:5173/file-upload?token=%s", token)

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Enlace de Subida de Archivos</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #2563eb; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f8fafc; padding: 30px; border-radius: 0 0 8px 8px; }
        .button { display: inline-block; background: #16a34a; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: bold; margin: 20px 0; }
        .info { background: #dbeafe; padding: 15px; border-radius: 6px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; color: #6b7280; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📁 Subir Archivos</h1>
            <p>Rental Manager</p>
        </div>
        
        <div class="content">
            <h2>Hola %s,</h2>
            
            <p>Se ha generado un enlace especial para que puedas subir archivos de forma segura.</p>
            
            <div class="info">
                <strong>📋 Instrucciones:</strong>
                <ul>
                    <li>Haz clic en el botón de abajo para acceder al portal de subida</li>
                    <li>Inicia sesión con tu cuenta</li>
                    <li>Selecciona y sube tus archivos</li>
                    <li>Los archivos serán revisados por un administrador</li>
                </ul>
            </div>
            
            <div style="text-align: center;">
                <a href="%s" class="button">🔗 Subir Archivos</a>
            </div>
            
            <div class="info">
                <strong>⚠️ Importante:</strong>
                <ul>
                    <li>Este enlace expira en 7 días</li>
                    <li>Solo puedes subir archivos PDF e imágenes</li>
                    <li>Tamaño máximo por archivo: 5MB</li>
                    <li>Debes estar autenticado para subir archivos</li>
                </ul>
            </div>
            
            <p>Si tienes alguna pregunta, no dudes en contactarnos.</p>
            
            <p>Saludos,<br>
            <strong>Equipo de Rental Manager</strong></p>
        </div>
        
        <div class="footer">
            <p>Este es un email automático, por favor no respondas a este mensaje.</p>
            <p>Si no solicitaste este enlace, puedes ignorar este email.</p>
        </div>
    </div>
</body>
</html>
	`, name, uploadURL)

	return SendProtonMailEmail(to, subject, htmlBody)
}
