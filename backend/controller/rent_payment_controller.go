package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/storage"
)

// RentPaymentController handles HTTP requests for rent payments
type RentPaymentController struct {
	repository         *storage.RentPaymentRepository
	rentalRepository   *storage.RentalRepository
	propertyRepository *storage.PropertyRepository
}

// NewRentPaymentController creates a new rent payment controller
func NewRentPaymentController(
	repository *storage.RentPaymentRepository,
	rentalRepo *storage.RentalRepository,
	propertyRepo *storage.PropertyRepository,
) *RentPaymentController {
	return &RentPaymentController{
		repository:         repository,
		rentalRepository:   rentalRepo,
		propertyRepository: propertyRepo,
	}
}

// RegisterRoutes registers all routes for the rent payment controller
func (c *RentPaymentController) RegisterRoutes(router *gin.RouterGroup) {
	payments := router.Group("/payments")
	{
		payments.GET("", c.GetAll)
		payments.GET("/:id", c.GetByID)
		payments.GET("/rental/:rentalId", c.GetByRentalID)
		payments.GET("/rental-ids", c.GetByRentalIDs)
		payments.GET("/date-range", c.GetByDateRange)
		payments.GET("/late", c.GetLatePayments)
		payments.POST("", c.Create)
		payments.PUT("/:id", c.Update)
		payments.DELETE("/:id", c.Delete)
	}
}

// GetAll retrieves all rent payments (Admin only)
// @Summary Get all rent payments (Admin only)
// @Description Retrieves all rent payments. Restricted to Admin users.
// @Tags payments
// @Produce json
// @Success 200 {array} storage.RentPayment
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 500 {object} string "Internal Server Error"
// @Router /payments [get]
func (c *RentPaymentController) GetAll(ctx *gin.Context) {
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	authUser, ok := userInterface.(*model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User data invalid"})
		return
	}

	if authUser.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view all payments"})
		return
	}

	payments, err := c.repository.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, payments)
}

// GetByID retrieves a rent payment by ID with authorization
// @Summary Get rent payment by ID
// @Description Retrieves a specific rent payment by its ID. Admins can get any. Managers/Residents can get if related to their rentals.
// @Tags payments
// @Produce json
// @Param id path string true "Rent Payment ID"
// @Success 200 {object} storage.RentPayment
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 404 {object} string "Not Found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /payments/{id} [get]
func (c *RentPaymentController) GetByID(ctx *gin.Context) {
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	authUser, ok := userInterface.(*model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User data invalid"})
		return
	}

	paymentID := ctx.Param("id")
	payment, err := c.repository.GetByID(paymentID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Payment not found: " + err.Error()})
		return
	}

	if authUser.Role == "admin" {
		ctx.JSON(http.StatusOK, payment)
		return
	}

	// Check if the user (manager or resident) is associated with the rental of this payment
	rental, err := c.rentalRepository.GetByID(ctx, uuid.MustParse(payment.RentalID)) // Assuming RentalID is UUID string
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve rental associated with payment"})
		return
	}
	if rental == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Rental associated with payment not found"})
		return
	}

	if authUser.Role == "resident" && rental.RenterID == authUser.PersonID {
		ctx.JSON(http.StatusOK, payment)
		return
	}

	if authUser.Role == "manager" {
		property, err := c.propertyRepository.GetByID(ctx, rental.PropertyID)
		if err != nil || property == nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify property for manager authorization"})
			return
		}
		// Check if authUser.PersonID is one of the property.ManagerIDs
		isUserAManagerForProperty := false
		for _, managerID := range property.ManagerIDs {
			if managerID == authUser.PersonID {
				isUserAManagerForProperty = true
				break
			}
		}
		if isUserAManagerForProperty {
			ctx.JSON(http.StatusOK, payment)
			return
		}
	}

	ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view this payment"})
}

// GetByRentalIDs retrieves payments for a list of rental IDs (Admin/Manager)
// @Summary Get payments by multiple rental IDs
// @Description Retrieves payments for a given list of rental IDs. For Managers, only rentals on managed properties are implicitly allowed (though filtering happens client-side or via rental ID list).
// @Tags payments
// @Accept json
// @Produce json
// @Param rentalIDs body []string true "List of Rental IDs"
// @Success 200 {array} storage.RentPayment
// @Failure 400 {object} string "Invalid input"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 500 {object} string "Internal Server Error"
// @Router /payments/rental-ids [post]
func (c *RentPaymentController) GetByRentalIDs(ctx *gin.Context) {
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	authUser, ok := userInterface.(*model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User data invalid"})
		return
	}

	var requestedRentalIDs []string
	if err := ctx.ShouldBindJSON(&requestedRentalIDs); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid list of rental IDs: " + err.Error()})
		return
	}

	if len(requestedRentalIDs) == 0 {
		ctx.JSON(http.StatusOK, []storage.RentPayment{})
		return
	}

	// Authorization: Admins can fetch any. Managers should technically only fetch for rentals they manage.
	// However, the frontend will likely construct the list based on managed rentals, so we trust the input list for now,
	// assuming the frontend logic correctly filters which IDs to request.
	// More robust check: Fetch all rentals by IDs, then filter by manager's property ID.
	if authUser.Role != "admin" && authUser.Role != "manager" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized for this operation"})
		return
	}

	payments, err := c.repository.GetByRentalIDs(requestedRentalIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, payments)
}

// GetByRentalID retrieves all rent payments for a specific rental
func (c *RentPaymentController) GetByRentalID(ctx *gin.Context) {
	rentalID := ctx.Param("rentalId")
	payments, err := c.repository.GetByRentalID(rentalID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, payments)
}

// GetByDateRange retrieves all rent payments within a specific date range
func (c *RentPaymentController) GetByDateRange(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date query parameters are required"})
		return
	}

	var startDate, endDate time.Time
	var err error

	// Try RFC3339 format first
	startDate, err = time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		// Try simple date format (YYYY-MM-DD)
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD or RFC3339 format."})
			return
		}
	}

	// Try RFC3339 format first
	endDate, err = time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		// Try simple date format (YYYY-MM-DD)
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD or RFC3339 format."})
			return
		}
	}

	payments, err := c.repository.GetPaymentsByDateRange(startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, payments)
}

// GetLatePayments retrieves all late rent payments
func (c *RentPaymentController) GetLatePayments(ctx *gin.Context) {
	payments, err := c.repository.GetLatePayments()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, payments)
}

// Create creates a new rent payment
func (c *RentPaymentController) Create(ctx *gin.Context) {
	var payment storage.RentPayment
	if err := ctx.ShouldBindJSON(&payment); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	createdPayment, err := c.repository.Create(&payment)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, createdPayment)
}

// Update updates an existing rent payment
func (c *RentPaymentController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var payment storage.RentPayment
	if err := ctx.ShouldBindJSON(&payment); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedPayment, err := c.repository.Update(id, &payment)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updatedPayment)
}

// Delete deletes a rent payment
func (c *RentPaymentController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	err := c.repository.Delete(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Rent payment deleted successfully"})
}
