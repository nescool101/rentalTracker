package controller

import (
	"net/http"
	"time"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/storage"
)

// MaintenanceRequestController handles HTTP requests for maintenance requests
type MaintenanceRequestController struct {
	repository         *storage.MaintenanceRequestRepository
	propertyRepository *storage.PropertyRepository
	rentalRepository   *storage.RentalRepository
}

// NewMaintenanceRequestController creates a new maintenance request controller
func NewMaintenanceRequestController(
	repository *storage.MaintenanceRequestRepository,
	propertyRepo *storage.PropertyRepository,
	rentalRepo *storage.RentalRepository,
) *MaintenanceRequestController {
	return &MaintenanceRequestController{
		repository:         repository,
		propertyRepository: propertyRepo,
		rentalRepository:   rentalRepo,
	}
}

// RegisterRoutes registers all routes for the maintenance request controller
func (c *MaintenanceRequestController) RegisterRoutes(router *gin.RouterGroup) {
	maintenance := router.Group("/maintenance-requests")
	{
		maintenance.GET("", c.GetAll)
		maintenance.GET("/:id", c.GetByID)
		maintenance.GET("/property/:propertyId", c.GetByPropertyID)
		maintenance.POST("/property-ids", c.GetByPropertyIDs)
		maintenance.GET("/renter/:renterId", c.GetByRenterID)
		maintenance.GET("/status/:status", c.GetByStatus)
		maintenance.POST("", c.Create)
		// The following routes are registered in admin section of http_controller.go
		// maintenance.PUT("/:id", c.Update)
		// maintenance.DELETE("/:id", c.Delete)
	}
}

// GetAll retrieves all maintenance requests
func (c *MaintenanceRequestController) GetAll(ctx *gin.Context) {
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
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view all maintenance requests"})
		return
	}

	requests, err := c.repository.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, requests)
}

// GetByID retrieves a maintenance request by ID
func (c *MaintenanceRequestController) GetByID(ctx *gin.Context) {
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

	requestID := ctx.Param("id")
	request, err := c.repository.GetByID(requestID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Maintenance request not found: " + err.Error()})
		return
	}

	if authUser.Role == "admin" {
		ctx.JSON(http.StatusOK, request)
		return
	}

	if authUser.Role == "manager" {
		managedProperties, err := c.propertyRepository.GetPropertiesForManager(ctx, authUser.PersonID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify manager properties"})
			return
		}
		for _, p := range managedProperties {
			if p.ID.String() == request.PropertyID {
				ctx.JSON(http.StatusOK, request)
				return
			}
		}
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Manager not authorized for this request"})
		return
	}

	if authUser.Role == "resident" || authUser.Role == "user" {
		if request.RenterID == authUser.PersonID.String() {
			ctx.JSON(http.StatusOK, request)
			return
		}
		rentals, err := c.rentalRepository.GetByRenterID(ctx, authUser.PersonID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify resident rentals"})
			return
		}
		for _, r := range rentals {
			if r.PropertyID.String() == request.PropertyID {
				ctx.JSON(http.StatusOK, request)
				return
			}
		}
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Resident/User not authorized for this request"})
		return
	}

	ctx.JSON(http.StatusForbidden, gin.H{"error": "User role not authorized"})
}

// GetByPropertyIDs retrieves maintenance requests for a list of property IDs
func (c *MaintenanceRequestController) GetByPropertyIDs(ctx *gin.Context) {
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

	var requestedPropertyIDs []string
	if err := ctx.ShouldBindJSON(&requestedPropertyIDs); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid list of property IDs: " + err.Error()})
		return
	}

	if len(requestedPropertyIDs) == 0 {
		ctx.JSON(http.StatusOK, []storage.MaintenanceRequest{})
		return
	}

	var authorizedPropertyIDs []string
	if authUser.Role == "admin" {
		authorizedPropertyIDs = requestedPropertyIDs
	} else if authUser.Role == "manager" {
		managedProperties, err := c.propertyRepository.GetPropertiesForManager(ctx, authUser.PersonID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify manager properties"})
			return
		}
		managedPropertyIDMap := make(map[string]bool)
		for _, p := range managedProperties {
			managedPropertyIDMap[p.ID.String()] = true
		}
		for _, reqID := range requestedPropertyIDs {
			if managedPropertyIDMap[reqID] {
				authorizedPropertyIDs = append(authorizedPropertyIDs, reqID)
			}
		}
	} else if authUser.Role == "resident" || authUser.Role == "user" {
		if authUser.PersonID == uuid.Nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "User PersonID not found in token"})
			return
		}
		userAssociatedPropertyIDs := make(map[string]bool)

		rentals, err := c.rentalRepository.GetByRenterID(ctx, authUser.PersonID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify user rentals"})
			return
		}
		for _, r := range rentals {
			userAssociatedPropertyIDs[r.PropertyID.String()] = true
		}

		residentProperties, err := c.propertyRepository.GetByResident(ctx, authUser.PersonID)
		if err != nil {
			log.Printf("Error fetching direct resident properties for user %s: %v", authUser.PersonID, err)
		}
		for _, p := range residentProperties {
			userAssociatedPropertyIDs[p.ID.String()] = true
		}

		for _, reqID := range requestedPropertyIDs {
			if userAssociatedPropertyIDs[reqID] {
				authorizedPropertyIDs = append(authorizedPropertyIDs, reqID)
			}
		}
	} else {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized for this operation"})
		return
	}

	if len(authorizedPropertyIDs) == 0 {
		ctx.JSON(http.StatusOK, []storage.MaintenanceRequest{})
		return
	}

	requests, err := c.repository.GetByPropertyIDs(authorizedPropertyIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch maintenance requests: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, requests)
}

// GetByPropertyID retrieves all maintenance requests for a specific property
func (c *MaintenanceRequestController) GetByPropertyID(ctx *gin.Context) {
	propertyID := ctx.Param("propertyId")
	requests, err := c.repository.GetByPropertyID(propertyID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, requests)
}

// GetByRenterID retrieves all maintenance requests from a specific renter
func (c *MaintenanceRequestController) GetByRenterID(ctx *gin.Context) {
	renterID := ctx.Param("renterId")
	requests, err := c.repository.GetByRenterID(renterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, requests)
}

// GetByStatus retrieves all maintenance requests with a specific status
func (c *MaintenanceRequestController) GetByStatus(ctx *gin.Context) {
	status := ctx.Param("status")
	requests, err := c.repository.GetByStatus(status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, requests)
}

// Create creates a new maintenance request
func (c *MaintenanceRequestController) Create(ctx *gin.Context) {
	var modelRequest model.MaintenanceRequest
	if err := ctx.ShouldBindJSON(&modelRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	authUser, ok := userInterface.(*model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User data invalid in context"})
		return
	}

	if authUser.Role == "resident" {
		if authUser.PersonID == uuid.Nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "User creating request does not have an associated PersonID"})
			return
		}
		modelRequest.RenterID = authUser.PersonID

		if modelRequest.PropertyID == uuid.Nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "PropertyID is required for resident maintenance requests"})
			return
		}

	} else if authUser.Role == "manager" {
		if modelRequest.PropertyID == uuid.Nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "PropertyID is required for manager maintenance requests"})
			return
		}
	}

	if modelRequest.PropertyID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "PropertyID is required"})
		return
	}
	if modelRequest.Description == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Description is required"})
		return
	}

	if modelRequest.RequestDate.Time().IsZero() {
		modelRequest.RequestDate = model.FlexibleTime(time.Now())
	}

	if modelRequest.Status == "" {
		modelRequest.Status = "Pending"
	}

	storageRequest := storage.MaintenanceRequest{
		PropertyID:  modelRequest.PropertyID.String(),
		RenterID:    modelRequest.RenterID.String(),
		RequestDate: modelRequest.RequestDate,
		Status:      modelRequest.Status,
		Description: modelRequest.Description,
	}

	createdRequest, err := c.repository.Create(&storageRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create maintenance request: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, createdRequest)
}

// Update updates an existing maintenance request
func (c *MaintenanceRequestController) Update(ctx *gin.Context) {
	id := ctx.Param("id")

	var request storage.MaintenanceRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedRequest, err := c.repository.Update(id, &request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedRequest)
}

// Delete deletes a maintenance request
func (c *MaintenanceRequestController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	err := c.repository.Delete(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Maintenance request deleted successfully"})
}
