package main

import (
	"log"

	// "github.com/gin-gonic/gin" // Not used directly if StartHTTPServer handles router setup
	"github.com/joho/godotenv"
	"github.com/nescool101/rentManager/config"
	"github.com/nescool101/rentManager/controller"
	"github.com/nescool101/rentManager/service"
	// Keep for service.StartScheduler if un-commented
	// storage package might not be needed directly in main anymore
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Using default configurations.")
	}

	// Initialize email configuration
	config.InitEmailConfig()

	// Initialize Supabase Storage service for file uploads
	if err := service.InitializeSupabaseStorageService(); err != nil {
		log.Printf("❌ Error inicializando Supabase Storage: %v", err)
		log.Fatal("No se puede continuar sin servicio de archivos")
	}

	// Initialize Telegram service for file backup (only if enabled)
	if service.IsTelegramEnabled() {
		if err := service.InitializeTelegramService(); err != nil {
			log.Printf("⚠️ Advertencia: Servicio de Telegram no disponible: %v", err)
			log.Printf("ℹ️ Los archivos se eliminarán sin backup en Telegram")
		}
	} else {
		log.Printf("ℹ️ Integración de Telegram deshabilitada por feature flag")
		log.Printf("💡 Para habilitar: establece TELEGRAM_ENABLED=true en .env")
	}

	// Supabase Storage handles file management automatically

	// storage.InitializePayersFile() // Removed as per request

	// service.LoadPayers() // Removed as per request

	// go service.StartScheduler() // Temporarily commented out. Uncomment and ensure logic is DB-based if used.

	// Start HTTP server - Controllers and Repositories are initialized within this function
	// This assumes StartHTTPServer initializes the Gin router and all routes.
	if err := controller.StartHTTPServer(); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
