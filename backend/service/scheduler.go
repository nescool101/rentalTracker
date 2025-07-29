package service

import (
	"log"

	"github.com/nescool101/rentManager/storage"
	"github.com/robfig/cron/v3"
)

// StartScheduler initializes and starts the cron scheduler.
func StartScheduler(personRepo *storage.PersonRepository, rentalRepo *storage.RentalRepository, propertyRepo *storage.PropertyRepository, userRepo *storage.UserRepository, pricingRepo *storage.PricingRepository) {
	c := cron.New()
	_, err := c.AddFunc("@monthly", func() { // You can change the schedule as needed, e.g., "0 0 1 * *" for 1st of every month
		log.Println("🗓️ [SCHEDULER] Running monthly notification job via cron...")
		NotifyAll(personRepo, rentalRepo, propertyRepo, userRepo, pricingRepo)
	})
	if err != nil {
		log.Fatalf("❌ [CRITICAL] Error adding cron job to scheduler: %v", err)
	}
	log.Println("ℹ️ [SCHEDULER] Cron scheduler started. Monthly notification job registered.")
	c.Start()

	// Keep the scheduler running in the background if this function is run as a goroutine.
	// If StartScheduler is the main blocking call for the service, this select{} is appropriate.
	// If it's a goroutine, this will block the goroutine indefinitely.
	// select {}
}
