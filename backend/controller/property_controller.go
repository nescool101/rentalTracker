package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"log"

	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/storage"
)

// PropertyController handles HTTP requests for property entities
type PropertyController struct {
	repository *storage.PropertyRepository
}

// NewPropertyController creates a new PropertyController
func NewPropertyController(repository *storage.PropertyRepository) *PropertyController {
	return &PropertyController{
		repository: repository,
	}
}

// GetAll retrieves properties based on user role
// @Summary Get properties (role-based)
// @Description Get properties. Admins get all. Managers get their managed properties. Residents get their resident properties.
// @Tags properties
// @Accept json
// @Produce json
// @Success 200 {array} model.Property
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Router /properties [get]
func (c *PropertyController) GetAll(ctx *gin.Context) {
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

	var properties []model.Property
	var err error

	switch authUser.Role {
	case "admin":
		properties, err = c.repository.GetAll(ctx)
	case "manager":
		if authUser.PersonID == uuid.Nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Manager PersonID not found in token"})
			return
		}
		properties, err = c.repository.GetPropertiesForManager(ctx, authUser.PersonID)
	case "resident":
		if authUser.PersonID == uuid.Nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Resident PersonID not found in token"})
			return
		}
		// Assuming resident's properties are linked via property.resident_id which is authUser.PersonID
		// If it's via rentals, this logic would need to use GetByUserID or similar
		properties, err = c.repository.GetByResident(ctx, authUser.PersonID)
	default:
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view these properties"})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve properties: " + err.Error()})
		return
	}

	if properties == nil { // Ensure properties is not nil, make it an empty slice if no results
		properties = []model.Property{}
	}

	ctx.JSON(http.StatusOK, properties)
}

// GetByID retrieves a property by ID
// @Summary Get property by ID
// @Description Get property by ID
// @Tags properties
// @Accept json
// @Produce json
// @Param id path string true "Property ID"
// @Success 200 {object} model.Property
// @Failure 404 {object} string "Property not found"
// @Router /properties/{id} [get]
func (c *PropertyController) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	property, err := c.repository.GetByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if property == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	ctx.JSON(http.StatusOK, property)
}

// GetByResident retrieves properties by resident ID
// @Summary Get properties by resident ID
// @Description Get properties by resident ID
// @Tags properties
// @Accept json
// @Produce json
// @Param residentId path string true "Resident ID"
// @Success 200 {array} model.Property
// @Router /properties/resident/{residentId} [get]
func (c *PropertyController) GetByResident(ctx *gin.Context) {
	residentID, err := uuid.Parse(ctx.Param("residentId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Authorization check
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

	// Admins can access any resident's properties.
	// Other users (residents) can only access their own.
	// Assuming "user" role is equivalent to "resident" for this check.
	if authUser.Role != "admin" && authUser.Role != "manager" { // Managers should not use this endpoint directly for other residents
		if authUser.PersonID != residentID {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only access your own resident properties"})
			return
		}
	}

	properties, err := c.repository.GetByResident(ctx, residentID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, properties)
}

// GetByManagerID retrieves properties by manager ID
// @Summary Get properties by manager ID
// @Description Get properties by manager ID
// @Tags properties
// @Accept json
// @Produce json
// @Param managerId path string true "Manager ID"
// @Success 200 {array} model.Property
// @Router /properties/manager/{managerId} [get]
func (c *PropertyController) GetByManagerID(ctx *gin.Context) {
	managerID, err := uuid.Parse(ctx.Param("managerId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Authorization: Check if the authenticated user is the manager or an admin
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

	if authUser.Role != "admin" && authUser.PersonID != managerID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only access properties you manage"})
		return
	}

	properties, err := c.repository.GetPropertiesForManager(ctx, managerID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, properties)
}

// GetByUserID retrieves properties that a user is renting
// @Summary Get properties for a specific user
// @Description Get properties that a user is renting
// @Tags properties
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {array} model.Property
// @Router /properties/user/{userId} [get]
func (c *PropertyController) GetByUserID(ctx *gin.Context) {
	// Extract userId from path parameter
	userID, err := uuid.Parse(ctx.Param("userId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Get the user from the context
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

	// If the user is an admin, they can see any user's properties
	// If not, they can only see their own properties
	if authUser.Role != "admin" && authUser.ID != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only access your own properties"})
		return
	}

	// Get properties for this user based on rentals
	properties, err := c.repository.GetByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, properties)
}

// Create adds a new property
// @Summary Create a new property
// @Description Create a new property
// @Tags properties
// @Accept json
// @Produce json
// @Param property body model.Property true "Property object"
// @Success 201 {object} model.Property
// @Router /properties [post]
func (c *PropertyController) Create(ctx *gin.Context) {
	var property model.Property
	if err := ctx.ShouldBindJSON(&property); err != nil {
		log.Printf("Error binding property JSON: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log the incoming property data
	log.Printf("Creating property: %+v", property)
	log.Printf("Manager IDs received: %v", property.ManagerIDs)

	// Authorization check
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

	log.Printf("Auth user: %s (ID: %s, PersonID: %s, Role: %s)",
		authUser.Email, authUser.ID, authUser.PersonID, authUser.Role)

	// If user is manager, they can only create properties they manage
	if authUser.Role == "manager" {
		// Verify manager is listed in ManagerIDs
		managerFound := false
		for _, managerID := range property.ManagerIDs {
			if managerID == authUser.PersonID {
				managerFound = true
				break
			}
		}

		if !managerFound {
			log.Printf("Manager %s (PersonID: %s) attempting to create property without themselves as manager",
				authUser.Email, authUser.PersonID)
			// Add manager to ManagerIDs
			property.ManagerIDs = append(property.ManagerIDs, authUser.PersonID)
			log.Printf("Added authenticated manager's PersonID to ManagerIDs: %s", authUser.PersonID)
		}
	} else if authUser.Role != "admin" {
		// Only managers and admins can create properties
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Only managers and administrators can create properties"})
		return
	}

	// Generate a new UUID if not provided
	if property.ID == uuid.Nil {
		property.ID = uuid.New()
		log.Printf("Generated new property ID: %s", property.ID)
	}

	createdProperty, err := c.repository.Create(ctx, property)
	if err != nil {
		log.Printf("Error creating property in repository: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Property created successfully: %+v", createdProperty)
	log.Printf("Property manager IDs: %v", createdProperty.ManagerIDs)

	ctx.JSON(http.StatusCreated, createdProperty)
}

// Update updates an existing property
// @Summary Update a property
// @Description Update a property
// @Tags properties
// @Accept json
// @Produce json
// @Param id path string true "Property ID"
// @Param property body model.Property true "Property object"
// @Success 200 {object} model.Property
// @Failure 404 {object} string "Property not found"
// @Router /properties/{id} [put]
func (c *PropertyController) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Get existing property to check permissions
	existingProperty, err := c.repository.GetByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if existingProperty == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	// Authorization check
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

	// Managers can only update their own properties - check if they are in ManagerIDs
	if authUser.Role == "manager" {
		managerFound := false
		for _, managerID := range existingProperty.ManagerIDs {
			if managerID == authUser.PersonID {
				managerFound = true
				break
			}
		}

		if !managerFound {
			log.Printf("Manager %s (PersonID: %s) attempting to update property they don't manage",
				authUser.Email, authUser.PersonID)
			ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only update properties you manage"})
			return
		}
	} else if authUser.Role != "admin" && authUser.Role != "manager" {
		// Only admins and managers can update properties
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Only managers and administrators can update properties"})
		return
	}

	var property model.Property
	if err := ctx.ShouldBindJSON(&property); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure the ID in the URL matches the ID in the body
	property.ID = id

	// If manager, ensure they stay in the ManagerIDs list
	if authUser.Role == "manager" {
		// Preserve current managers and ensure current user is included
		preserveManagerID := authUser.PersonID
		managerFound := false

		for _, managerID := range property.ManagerIDs {
			if managerID == preserveManagerID {
				managerFound = true
				break
			}
		}

		if !managerFound {
			// Add manager to the list to prevent them from removing themselves
			property.ManagerIDs = append(property.ManagerIDs, preserveManagerID)
			log.Printf("Ensured manager %s remains in ManagerIDs for property %s",
				preserveManagerID, property.ID)
		}
	}

	updatedProperty, err := c.repository.Update(ctx, property)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedProperty)
}

// Delete removes a property
// @Summary Delete a property
// @Description Delete a property
// @Tags properties
// @Accept json
// @Produce json
// @Param id path string true "Property ID"
// @Success 204 "No Content"
// @Router /properties/{id} [delete]
func (c *PropertyController) Delete(ctx *gin.Context) {
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

// RegisterRoutes registers the routes for the property controller
func (c *PropertyController) RegisterRoutes(router *gin.RouterGroup) {
	properties := router.Group("/properties")
	{
		properties.GET("", c.GetAll)
		properties.GET("/:id", c.GetByID)
		properties.GET("/resident/:residentId", c.GetByResident)
		properties.GET("/manager/:managerId", c.GetByManagerID)
		properties.GET("/user/:userId", c.GetByUserID)
		properties.POST("", c.Create)
		properties.PUT("/:id", c.Update)
		properties.DELETE("/:id", c.Delete)
	}
}
