package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/storage"
)

// RentalHistoryController handles HTTP requests for rental history
type RentalHistoryController struct {
	repository   *storage.RentalHistoryRepository
	rentalRepo   *storage.RentalRepository
	propertyRepo *storage.PropertyRepository
	personRepo   *storage.PersonRepository // Added for fetching person details if needed
}

// NewRentalHistoryController creates a new rental history controller
func NewRentalHistoryController(
	repository *storage.RentalHistoryRepository,
	rentalRepo *storage.RentalRepository,
	propertyRepo *storage.PropertyRepository,
	personRepo *storage.PersonRepository,
) *RentalHistoryController {
	return &RentalHistoryController{
		repository:   repository,
		rentalRepo:   rentalRepo,
		propertyRepo: propertyRepo,
		personRepo:   personRepo,
	}
}

func getAuthenticatedUser(ctx *gin.Context) (*model.User, bool) {
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return nil, false
	}
	authUser, ok := userInterface.(*model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User data invalid"})
		return nil, false
	}
	return authUser, true
}

// RegisterRoutes registers all routes for the rental history controller
func (c *RentalHistoryController) RegisterRoutes(router *gin.RouterGroup) {
	history := router.Group("/rental-history")
	{
		history.GET("", c.GetAll)
		history.GET("/:id", c.GetByID)
		history.GET("/person/:personId", c.GetByPersonID)
		history.GET("/rental/:rentalId", c.GetByRentalID)
		history.GET("/status/:status", c.GetByStatus)          // Needs auth
		history.GET("/date-range", c.GetByDateRange)           // Needs auth
		history.POST("/for-rentals", c.GetMultipleByRentalIDs) // New route

		// CUD operations are typically registered in http_controller.go under admin middleware
		// If they are also registered here for some reason, they'd need admin checks too.
		// For now, assuming CUD are handled by admin middleware routes.
		// history.POST("", c.Create)
		// history.PUT("/:id", c.Update)
		// history.DELETE("/:id", c.Delete)
	}
}

// GetAll retrieves rental history records based on user role and optional admin filters.
// Admins can filter by status or date_range query parameters.
// Managers get history for rentals on their managed properties.
// Residents get history for their own rentals.
func (c *RentalHistoryController) GetAll(ctx *gin.Context) {
	authUser, ok := getAuthenticatedUser(ctx)
	if !ok {
		return
	}

	var histories []storage.RentalHistory
	var err error

	switch authUser.Role {
	case "admin":
		statusFilter := ctx.Query("status")
		startDateStr := ctx.Query("start_date")
		endDateStr := ctx.Query("end_date")

		if statusFilter != "" {
			histories, err = c.repository.GetByStatus(statusFilter)
		} else if startDateStr != "" && endDateStr != "" {
			var startDate, endDate time.Time
			parseAdminDate := func(dateStr string) (time.Time, error) {
				t, pErr := time.Parse(time.RFC3339, dateStr)
				if pErr == nil {
					return t, nil
				}
				return time.Parse("2006-01-02", dateStr)
			}
			startDate, err = parseAdminDate(startDateStr)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format for admin filter."})
				return
			}
			endDate, err = parseAdminDate(endDateStr)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format for admin filter."})
				return
			}
			histories, err = c.repository.GetRentalHistoryByDateRange(startDate, endDate)
		} else {
			histories, err = c.repository.GetAll() // Admin gets all if no filters
		}
	case "manager":
		if authUser.PersonID == uuid.Nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Manager PersonID not found in token"})
			return
		}
		managedProperties, propErr := c.propertyRepo.GetPropertiesForManager(ctx, authUser.PersonID)
		if propErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch managed properties: " + propErr.Error()})
			return
		}
		if len(managedProperties) == 0 {
			ctx.JSON(http.StatusOK, []storage.RentalHistory{}) // No properties, so no history
			return
		}

		var allRentalIDs []string
		for _, prop := range managedProperties {
			rentalsOnProp, rentalErr := c.rentalRepo.GetByPropertyID(ctx, prop.ID)
			if rentalErr != nil {
				// Log error and continue, or decide to fail all
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching rentals for property " + prop.ID.String() + ": " + rentalErr.Error()})
				return
			}
			for _, rental := range rentalsOnProp {
				allRentalIDs = append(allRentalIDs, rental.ID.String())
			}
		}

		if len(allRentalIDs) == 0 {
			ctx.JSON(http.StatusOK, []storage.RentalHistory{}) // No rentals, so no history
			return
		}
		histories, err = c.repository.GetByRentalIDs(allRentalIDs)
	case "resident":
		if authUser.PersonID == uuid.Nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Resident PersonID not found in token"})
			return
		}
		userRentals, rentalErr := c.rentalRepo.GetByRenterID(ctx, authUser.PersonID)
		if rentalErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user rentals: " + rentalErr.Error()})
			return
		}
		if len(userRentals) == 0 {
			ctx.JSON(http.StatusOK, []storage.RentalHistory{}) // No rentals, so no history
			return
		}
		var userRentalIDs []string
		for _, rental := range userRentals {
			userRentalIDs = append(userRentalIDs, rental.ID.String())
		}
		histories, err = c.repository.GetByRentalIDs(userRentalIDs)
	default:
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view rental history"})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rental histories: " + err.Error()})
		return
	}

	if histories == nil { // Ensure histories is not nil
		histories = []storage.RentalHistory{}
	}

	ctx.JSON(http.StatusOK, histories)
}

// GetByID retrieves a rental history record by ID
func (c *RentalHistoryController) GetByID(ctx *gin.Context) {
	authUser, ok := getAuthenticatedUser(ctx)
	if !ok {
		return
	}

	historyIDStr := ctx.Param("id")
	// _, err := uuid.Parse(historyIDStr) // Validate if it's a UUID, though repo might handle non-UUID string IDs
	// if err != nil {
	// 	ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid history ID format"})
	// 	return
	// }

	history, err := c.repository.GetByID(historyIDStr)
	if err != nil {
		// Distinguish between not found and other errors if repository supports it
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Rental history not found or error fetching: " + err.Error()})
		return
	}
	if history == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Rental history not found"})
		return
	}

	// Authorization checks
	isAllowed := false
	if authUser.Role == "admin" {
		isAllowed = true
	} else if authUser.Role == "resident" {
		if history.PersonID == authUser.PersonID.String() {
			isAllowed = true
		}
	} else if authUser.Role == "manager" {
		rentalID_uuid, parseErr := uuid.Parse(history.RentalID)
		if parseErr == nil {
			rental, rentalErr := c.rentalRepo.GetByID(ctx, rentalID_uuid)
			if rentalErr == nil && rental != nil {
				property, propErr := c.propertyRepo.GetByID(ctx, rental.PropertyID)
				if propErr == nil && property != nil {
					for _, managerID := range property.ManagerIDs {
						if managerID == authUser.PersonID {
							isAllowed = true
							break
						}
					}
				}
			}
		}
	}

	if !isAllowed {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view this rental history record"})
		return
	}

	ctx.JSON(http.StatusOK, history)
}

// GetByPersonID retrieves all rental history records for a specific person
func (c *RentalHistoryController) GetByPersonID(ctx *gin.Context) {
	authUser, ok := getAuthenticatedUser(ctx)
	if !ok {
		return
	}

	targetPersonIDStr := ctx.Param("personId")
	targetPersonID_uuid, err := uuid.Parse(targetPersonIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid person ID format"})
		return
	}

	isAllowed := false
	if authUser.Role == "admin" {
		isAllowed = true
	} else if authUser.Role == "manager" {
		// Managers can view history for any person, as they might be tenants of managed properties.
		// Further filtering could be done on the frontend or by ensuring the list is contextualized.
		isAllowed = true
	} else if authUser.Role == "resident" || authUser.Role == "user" {
		if targetPersonID_uuid == authUser.PersonID {
			isAllowed = true
		}
	}

	if !isAllowed {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view rental history for this person"})
		return
	}

	histories, err := c.repository.GetByPersonID(targetPersonIDStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, histories)
}

// GetByRentalID retrieves all rental history records for a specific rental
func (c *RentalHistoryController) GetByRentalID(ctx *gin.Context) {
	authUser, ok := getAuthenticatedUser(ctx)
	if !ok {
		return
	}

	targetRentalIDStr := ctx.Param("rentalId")
	targetRentalID_uuid, err := uuid.Parse(targetRentalIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID format"})
		return
	}

	isAllowed := false
	if authUser.Role == "admin" {
		isAllowed = true
	} else {
		// Fetch the rental to check its associations
		rental, rentalErr := c.rentalRepo.GetByID(ctx, targetRentalID_uuid)
		if rentalErr != nil || rental == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Rental not found or error fetching rental"})
			return
		}

		if authUser.Role == "resident" {
			if rental.RenterID == authUser.PersonID {
				isAllowed = true
			}
		} else if authUser.Role == "manager" {
			property, propErr := c.propertyRepo.GetByID(ctx, rental.PropertyID)
			if propErr == nil && property != nil {
				isManagerForThisProperty := false
				for _, managerID := range property.ManagerIDs {
					if managerID == authUser.PersonID {
						isManagerForThisProperty = true
						break
					}
				}
				if !isManagerForThisProperty {
					ctx.JSON(http.StatusForbidden, gin.H{"error": "Manager not authorized for this rental's history"})
					return
				}
			}
		}
	}

	if !isAllowed {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view rental history for this rental"})
		return
	}

	histories, err := c.repository.GetByRentalID(targetRentalIDStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, histories)
}

// GetMultipleByRentalIDsInput defines the expected input for GetMultipleByRentalIDs
type GetMultipleByRentalIDsInput struct {
	RentalIDs []string `json:"rental_ids"`
}

// GetMultipleByRentalIDs retrieves rental history for a list of rental IDs
func (c *RentalHistoryController) GetMultipleByRentalIDs(ctx *gin.Context) {
	authUser, ok := getAuthenticatedUser(ctx)
	if !ok {
		return
	}

	var input GetMultipleByRentalIDsInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if len(input.RentalIDs) == 0 {
		ctx.JSON(http.StatusOK, []storage.RentalHistory{})
		return
	}

	allowedRentalIDs := []string{}

	if authUser.Role == "admin" {
		allowedRentalIDs = input.RentalIDs
	} else if authUser.Role == "resident" {
		userRentals, err := c.rentalRepo.GetByRenterID(ctx, authUser.PersonID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user rentals: " + err.Error()})
			return
		}
		userRentalIDMap := make(map[string]bool)
		for _, r := range userRentals {
			userRentalIDMap[r.ID.String()] = true
		}
		for _, reqID := range input.RentalIDs {
			if userRentalIDMap[reqID] {
				allowedRentalIDs = append(allowedRentalIDs, reqID)
			}
		}
	} else if authUser.Role == "manager" {
		managedProperties, err := c.propertyRepo.GetPropertiesForManager(ctx, authUser.PersonID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch managed properties: " + err.Error()})
			return
		}

		rentalsOnManagedPropertiesMap := make(map[string]bool)
		for _, prop := range managedProperties {
			rentals, err := c.rentalRepo.GetByPropertyID(ctx, prop.ID)
			if err != nil {
				// Log or handle error, maybe continue to next property
				continue
			}
			for _, r := range rentals {
				rentalsOnManagedPropertiesMap[r.ID.String()] = true
			}
		}
		for _, reqID := range input.RentalIDs {
			if rentalsOnManagedPropertiesMap[reqID] {
				allowedRentalIDs = append(allowedRentalIDs, reqID)
			}
		}
	}

	if len(allowedRentalIDs) == 0 {
		ctx.JSON(http.StatusOK, []storage.RentalHistory{})
		return
	}

	histories, err := c.repository.GetByRentalIDs(allowedRentalIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rental histories: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, histories)
}

// GetByStatus retrieves all rental history records with a specific status (Admin only)
func (c *RentalHistoryController) GetByStatus(ctx *gin.Context) {
	authUser, ok := getAuthenticatedUser(ctx)
	if !ok {
		return
	}
	if authUser.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	status := ctx.Param("status")
	histories, err := c.repository.GetByStatus(status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, histories)
}

// GetByDateRange retrieves all rental history records with end dates in a specific range (Admin only)
func (c *RentalHistoryController) GetByDateRange(ctx *gin.Context) {
	authUser, ok := getAuthenticatedUser(ctx)
	if !ok {
		return
	}
	if authUser.Role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date query parameters are required"})
		return
	}

	var startDate, endDate time.Time
	var err error

	parseDate := func(dateStr string) (time.Time, error) {
		t, err := time.Parse(time.RFC3339, dateStr)
		if err == nil {
			return t, nil
		}
		return time.Parse("2006-01-02", dateStr)
	}

	startDate, err = parseDate(startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD or RFC3339 format."})
		return
	}

	endDate, err = parseDate(endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD or RFC3339 format."})
		return
	}

	histories, err := c.repository.GetRentalHistoryByDateRange(startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, histories)
}

// Create creates a new rental history record
// Assumed to be admin-only via http_controller.go routing
func (c *RentalHistoryController) Create(ctx *gin.Context) {
	var history storage.RentalHistory // Using storage.RentalHistory as it's more complete
	if err := ctx.ShouldBindJSON(&history); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add explicit admin check here if not solely relying on router middleware,
	// or if this method is registered outside admin group.
	// For now, assume http_controller.go handles admin restriction.

	createdHistory, err := c.repository.Create(&history)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, createdHistory)
}

// Update updates an existing rental history record
// Assumed to be admin-only via http_controller.go routing
func (c *RentalHistoryController) Update(ctx *gin.Context) {
	id := ctx.Param("id")

	var history storage.RentalHistory
	if err := ctx.ShouldBindJSON(&history); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Add explicit admin check here if needed.

	updatedHistory, err := c.repository.Update(id, &history)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, updatedHistory)
}

// Delete deletes a rental history record
// Assumed to be admin-only via http_controller.go routing
func (c *RentalHistoryController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	// Add explicit admin check here if needed.

	err := c.repository.Delete(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Rental history record deleted successfully"})
}
