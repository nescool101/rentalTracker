package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/storage"
)

// RentalController handles HTTP requests for rental entities
type RentalController struct {
	repository   *storage.RentalRepository
	propertyRepo *storage.PropertyRepository // Added for manager logic
}

// NewRentalController creates a new RentalController
func NewRentalController(repository *storage.RentalRepository, propertyRepo *storage.PropertyRepository) *RentalController {
	return &RentalController{
		repository:   repository,
		propertyRepo: propertyRepo,
	}
}

// GetAll retrieves rentals based on user role.
// Admins get all. Managers get rentals for their managed properties.
// Other roles are forbidden from this endpoint.
// @Summary Get rentals (role-based for admin/manager)
// @Description Get rentals. Admins get all. Managers get rentals for their properties.
// @Tags rentals
// @Accept json
// @Produce json
// @Success 200 {array} model.Rental
// @Router /rentals [get]
func (c *RentalController) GetAll(ctx *gin.Context) {
	authUser, ok := getAuthenticatedUser(ctx) // Assuming getAuthenticatedUser is available or add it
	if !ok {
		return
	}

	var rentals []model.Rental
	var err error

	switch authUser.Role {
	case "admin":
		rentals, err = c.repository.GetAll(ctx)
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
			ctx.JSON(http.StatusOK, []model.Rental{}) // No properties, so no rentals
			return
		}

		var allRentalsForManager []model.Rental
		for _, prop := range managedProperties {
			rentalsOnProp, rentalErr := c.repository.GetByPropertyID(ctx, prop.ID)
			if rentalErr != nil {
				log.Printf("Error fetching rentals for property %s: %v. Skipping.", prop.ID.String(), rentalErr)
				// Decide if one error should fail all, or just skip this property's rentals
				continue
			}
			allRentalsForManager = append(allRentalsForManager, rentalsOnProp...)
		}
		rentals = allRentalsForManager
	default: // Other roles, including residents, are forbidden from this specific GetAll endpoint.
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view all rentals via this endpoint."})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rentals: " + err.Error()})
		return
	}

	if rentals == nil {
		rentals = []model.Rental{}
	}
	ctx.JSON(http.StatusOK, rentals)
}

// GetByID retrieves a rental by ID
// @Summary Get rental by ID
// @Description Get rental by ID
// @Tags rentals
// @Accept json
// @Produce json
// @Param id path string true "Rental ID"
// @Success 200 {object} model.Rental
// @Failure 404 {object} string "Rental not found"
// @Router /rentals/{id} [get]
func (c *RentalController) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	rental, err := c.repository.GetByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rental == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Rental not found"})
		return
	}

	ctx.JSON(http.StatusOK, rental)
}

// GetByPropertyID retrieves rentals by property ID
// @Summary Get rentals by property ID
// @Description Get rentals by property ID
// @Tags rentals
// @Accept json
// @Produce json
// @Param propertyId path string true "Property ID"
// @Success 200 {array} model.Rental
// @Router /rentals/property/{propertyId} [get]
func (c *RentalController) GetByPropertyID(ctx *gin.Context) {
	propertyID, err := uuid.Parse(ctx.Param("propertyId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	rentals, err := c.repository.GetByPropertyID(ctx, propertyID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, rentals)
}

// GetByRenterID retrieves rentals by renter ID
// @Summary Get rentals by renter ID
// @Description Get rentals by renter ID
// @Tags rentals
// @Accept json
// @Produce json
// @Param renterId path string true "Renter ID"
// @Success 200 {array} model.Rental
// @Router /rentals/renter/{renterId} [get]
func (c *RentalController) GetByRenterID(ctx *gin.Context) {
	renterID, err := uuid.Parse(ctx.Param("renterId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	rentals, err := c.repository.GetByRenterID(ctx, renterID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, rentals)
}

// GetActiveRentals retrieves all active rentals
// @Summary Get active rentals
// @Description Get active rentals
// @Tags rentals
// @Accept json
// @Produce json
// @Success 200 {array} model.Rental
// @Router /rentals/active [get]
func (c *RentalController) GetActiveRentals(ctx *gin.Context) {
	rentals, err := c.repository.GetActiveRentals(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, rentals)
}

// Create adds a new rental
// @Summary Create a new rental
// @Description Create a new rental
// @Tags rentals
// @Accept json
// @Produce json
// @Param rental body model.Rental true "Rental object"
// @Success 201 {object} model.Rental
// @Router /rentals [post]
func (c *RentalController) Create(ctx *gin.Context) {
	var rental model.Rental
	if err := ctx.ShouldBindJSON(&rental); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate a new UUID if not provided
	if rental.ID == uuid.Nil {
		rental.ID = uuid.New()
	}

	createdRental, err := c.repository.Create(ctx, rental)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, createdRental)
}

// Update updates an existing rental
// @Summary Update a rental
// @Description Update a rental
// @Tags rentals
// @Accept json
// @Produce json
// @Param id path string true "Rental ID"
// @Param rental body model.Rental true "Rental object"
// @Success 200 {object} model.Rental
// @Failure 404 {object} string "Rental not found"
// @Router /rentals/{id} [put]
func (c *RentalController) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var rental model.Rental
	if err := ctx.ShouldBindJSON(&rental); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure the ID in the URL matches the ID in the body
	rental.ID = id

	updatedRental, err := c.repository.Update(ctx, rental)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if updatedRental == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Rental not found"})
		return
	}

	ctx.JSON(http.StatusOK, updatedRental)
}

// Delete removes a rental
// @Summary Delete a rental
// @Description Delete a rental
// @Tags rentals
// @Accept json
// @Produce json
// @Param id path string true "Rental ID"
// @Success 204 "No Content"
// @Router /rentals/{id} [delete]
func (c *RentalController) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = c.repository.Delete(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// RegisterRoutes registers the routes for the rental controller
func (c *RentalController) RegisterRoutes(router *gin.RouterGroup) {
	rentals := router.Group("/rentals")
	{
		rentals.GET("", c.GetAll) // This is now role-aware for Admin/Manager
		rentals.GET("/:id", c.GetByID)
		rentals.GET("/property/:propertyId", c.GetByPropertyID)
		rentals.GET("/renter/:renterId", c.GetByRenterID)
		rentals.GET("/active", c.GetActiveRentals)
		// CUD operations should be registered under an admin-only group in http_controller.go
		// rentals.POST("", c.Create)
		// rentals.PUT("/:id", c.Update)
		// rentals.DELETE("/:id", c.Delete)
	}
}
