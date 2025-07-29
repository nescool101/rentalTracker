package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/storage"
)

// PricingController handles HTTP requests for pricing entities
type PricingController struct {
	repository *storage.PricingRepository
}

// NewPricingController creates a new PricingController
func NewPricingController(repository *storage.PricingRepository) *PricingController {
	return &PricingController{
		repository: repository,
	}
}

// RegisterRoutes sets up the pricing routes for an admin-protected group.
// The incoming 'adminRouter' is expected to be a group like /api/pricing
// that already has necessary admin middleware.
func (c *PricingController) RegisterRoutes(adminRouter *gin.RouterGroup) {
	// Routes will be registered directly on adminRouter.
	// For example, adminRouter.GET("", ...) will become GET /api/pricing
	// adminRouter.GET("/:id", ...) will become GET /api/pricing/:id

	adminRouter.POST("", c.HandleCreatePricing)                        // POST /api/pricing
	adminRouter.GET("", c.HandleGetAllPricing)                         // GET /api/pricing
	adminRouter.GET("/:id", c.HandleGetPricingByID)                    // GET /api/pricing/:id
	adminRouter.GET("/rental/:rentalId", c.HandleGetPricingByRentalID) // GET /api/pricing/rental/:rentalId
	adminRouter.PUT("/:id", c.HandleUpdatePricing)                     // PUT /api/pricing/:id
	adminRouter.DELETE("/:id", c.HandleDeletePricing)                  // DELETE /api/pricing/:id

	log.Println("INFO: Registered admin pricing routes under /api/pricing")
}

// HandleCreatePricing creates new pricing information
func (c *PricingController) HandleCreatePricing(ctx *gin.Context) {
	var pricing model.Pricing
	if err := ctx.ShouldBindJSON(&pricing); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Validate required fields (e.g., RentalID, MonthlyRent, DueDay)
	if pricing.RentalID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "RentalID is required"})
		return
	}
	if pricing.MonthlyRent <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "MonthlyRent must be positive"})
		return
	}
	if pricing.DueDay < 1 || pricing.DueDay > 31 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "DueDay must be between 1 and 31"})
		return
	}

	createdPricing, err := c.repository.Create(ctx, pricing)
	if err != nil {
		log.Printf("Error creating pricing: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pricing record"})
		return
	}
	ctx.JSON(http.StatusCreated, createdPricing)
}

// HandleGetAllPricing retrieves all pricing records
func (c *PricingController) HandleGetAllPricing(ctx *gin.Context) {
	pricingList, err := c.repository.GetAll(ctx)
	if err != nil {
		log.Printf("Error getting all pricing: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pricing records"})
		return
	}
	ctx.JSON(http.StatusOK, pricingList)
}

// HandleGetPricingByID retrieves pricing information by its ID
func (c *PricingController) HandleGetPricingByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Pricing ID format"})
		return
	}

	pricing, err := c.repository.GetByID(ctx, id)
	if err != nil {
		log.Printf("Error getting pricing by ID %s: %v", idStr, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pricing record"})
		return
	}
	if pricing == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Pricing record not found"})
		return
	}
	ctx.JSON(http.StatusOK, pricing)
}

// HandleGetPricingByRentalID retrieves pricing information for a specific rental ID
func (c *PricingController) HandleGetPricingByRentalID(ctx *gin.Context) {
	rentalIDStr := ctx.Param("rentalId")
	rentalID, err := uuid.Parse(rentalIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Rental ID format"})
		return
	}

	pricing, err := c.repository.GetByRentalID(ctx, rentalID) // Assumes one pricing per rental
	if err != nil {
		log.Printf("Error getting pricing by RentalID %s: %v", rentalIDStr, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pricing for rental"})
		return
	}
	if pricing == nil {
		// This is a valid case, a rental might not have pricing yet.
		ctx.JSON(http.StatusNotFound, gin.H{"message": "No pricing information found for this rental"})
		return
	}
	ctx.JSON(http.StatusOK, pricing)
}

// HandleUpdatePricing updates an existing pricing record
func (c *PricingController) HandleUpdatePricing(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Pricing ID format"})
		return
	}

	var pricingUpdate model.Pricing
	if err := ctx.ShouldBindJSON(&pricingUpdate); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Ensure the ID in the path matches the ID in the body, or set it.
	pricingUpdate.ID = id

	// Basic Validations
	if pricingUpdate.RentalID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "RentalID is required"})
		return
	}
	if pricingUpdate.MonthlyRent <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "MonthlyRent must be positive"})
		return
	}
	if pricingUpdate.DueDay < 1 || pricingUpdate.DueDay > 31 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "DueDay must be between 1 and 31"})
		return
	}

	updatedPricing, err := c.repository.Update(ctx, pricingUpdate)
	if err != nil {
		log.Printf("Error updating pricing ID %s: %v", idStr, err)
		// Could be an actual DB error or record not found if Update doesn't distinguish
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pricing record"})
		return
	}
	if updatedPricing == nil {
		// This can happen if the Update method returns nil when record not found
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Pricing record not found or no changes made"})
		return
	}

	ctx.JSON(http.StatusOK, updatedPricing)
}

// HandleDeletePricing deletes a pricing record by its ID
func (c *PricingController) HandleDeletePricing(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Pricing ID format"})
		return
	}

	// Optional: Check if pricing record exists before attempting delete
	// existing, _ := c.repository.GetByID(ctx, id)
	// if existing == nil {
	// 	ctx.JSON(http.StatusNotFound, gin.H{"error": "Pricing record not found"})
	// 	return
	// }

	err = c.repository.Delete(ctx, id)
	if err != nil {
		log.Printf("Error deleting pricing ID %s: %v", idStr, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete pricing record"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Pricing record deleted successfully"})
}
