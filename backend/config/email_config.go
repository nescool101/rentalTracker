package config

import (
	"log"
	"os"
	"strconv"

	"github.com/nescool101/rentManager/service"
)

// InitEmailConfig initializes the email configuration
func InitEmailConfig() {
	// Get email configuration from environment variables - REQUIRED
	username := os.Getenv("EMAIL_USER")
	password := os.Getenv("EMAIL_PASS")
	host := os.Getenv("EMAIL_HOST")
	portStr := os.Getenv("EMAIL_PORT")
	fromName := getEnvOr("EMAIL_FROM_NAME", "Sistema de Gestión de Propiedades")

	// Validate required environment variables
	if username == "" {
		log.Fatal("❌ ERROR: EMAIL_USER environment variable is required")
	}
	if password == "" {
		log.Fatal("❌ ERROR: EMAIL_PASS environment variable is required")
	}
	if host == "" {
		log.Fatal("❌ ERROR: EMAIL_HOST environment variable is required")
	}
	if portStr == "" {
		log.Fatal("❌ ERROR: EMAIL_PORT environment variable is required")
	}

	// Parse port number
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("❌ ERROR: EMAIL_PORT must be a valid number, got: %s", portStr)
	}

	// Update Gmail configuration (reusing ProtonMail config structure)
	service.DefaultProtonMailConfig = service.ProtonMailConfig{
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
		FromName: fromName,
	}

	log.Printf("✅ Email configuration loaded: %s@%s:%d", username, host, port)
}

// getEnvOr returns environment variable value or default if not set
func getEnvOr(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
