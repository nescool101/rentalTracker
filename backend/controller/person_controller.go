package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/storage"
)

// PersonController handles HTTP requests for person entities
type PersonController struct {
	repository      *storage.PersonRepository
	propertyRepo    *storage.PropertyRepository
	rentalRepo      *storage.RentalRepository
	bankAccountRepo *storage.BankAccountRepository
	userRepo        *storage.UserRepository
}

// NewPersonController creates a new PersonController
func NewPersonController(repository *storage.PersonRepository, propertyRepo *storage.PropertyRepository, rentalRepo *storage.RentalRepository, bankAccountRepo *storage.BankAccountRepository, userRepo *storage.UserRepository) *PersonController {
	return &PersonController{
		repository:      repository,
		propertyRepo:    propertyRepo,
		rentalRepo:      rentalRepo,
		bankAccountRepo: bankAccountRepo,
		userRepo:        userRepo,
	}
}

// GetAll retrieves all persons (Admin role) or persons related to a Manager's properties.
// @Summary Get all persons (role-based)
// @Description Get all persons. Admins get all. Managers get persons (renters and self) associated with their managed properties.
// @Tags persons
// @Accept json
// @Produce json
// @Success 200 {array} model.Person
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Router /persons [get]
func (c *PersonController) GetAll(ctx *gin.Context) {
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

	var persons []model.Person
	var err error

	if authUser.Role == "admin" {
		persons, err = c.repository.GetAll(ctx)
	} else if authUser.Role == "manager" {
		if authUser.PersonID == uuid.Nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Manager PersonID not found in token"})
			return
		}
		// Fetch properties managed by this manager
		managedProperties, propErr := c.propertyRepo.GetPropertiesForManager(ctx, authUser.PersonID)
		if propErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch managed properties: " + propErr.Error()})
			return
		}
		if len(managedProperties) == 0 {
			// Manager manages no properties, return at least the manager themself if they are a person
			managerPerson, personErr := c.repository.GetByID(ctx, authUser.PersonID)
			if personErr == nil && managerPerson != nil {
				persons = []model.Person{*managerPerson}
			} else {
				persons = []model.Person{}
			}
			ctx.JSON(http.StatusOK, persons)
			return
		}

		personIDsToFetch := make(map[uuid.UUID]bool)
		personIDsToFetch[authUser.PersonID] = true // Include the manager themselves

		for _, prop := range managedProperties {
			// Fetch rentals for each property
			rentalsOnProp, rentalErr := c.rentalRepo.GetByPropertyID(ctx, prop.ID)
			if rentalErr != nil {
				// Log and continue, or return error. For now, log and attempt to proceed.
				log.Printf("Error fetching rentals for property %s: %v", prop.ID.String(), rentalErr)
				continue
			}
			for _, rental := range rentalsOnProp {
				if rental.RenterID != uuid.Nil {
					personIDsToFetch[rental.RenterID] = true
				}
			}
		}

		if len(personIDsToFetch) > 0 {
			var idsSlice []uuid.UUID
			for id := range personIDsToFetch {
				idsSlice = append(idsSlice, id)
			}
			persons, err = c.repository.GetByIDs(ctx, idsSlice)
		} else {
			// Should not happen if manager personID was added, but as a fallback
			persons = []model.Person{}
		}
	} else {
		// For other roles (e.g. resident), GET /persons is forbidden as per original logic
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view all persons"})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve persons: " + err.Error()})
		return
	}
	if persons == nil {
		persons = []model.Person{}
	}
	ctx.JSON(http.StatusOK, persons)
}

// GetByID retrieves a person by ID
// @Summary Get person by ID
// @Description Get person by ID. Admins/Managers can get any. Residents can get their own.
// @Tags persons
// @Accept json
// @Produce json
// @Param id path string true "Person ID"
// @Success 200 {object} model.Person
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 404 {object} string "Person not found"
// @Router /persons/{id} [get]
func (c *PersonController) GetByID(ctx *gin.Context) {
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

	personIDStr := ctx.Param("id")
	personID, err := uuid.Parse(personIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Authorization checks
	isAllowed := false
	if authUser.Role == "admin" || authUser.Role == "manager" {
		isAllowed = true
	} else if (authUser.Role == "resident" || authUser.Role == "user") && authUser.PersonID == personID {
		isAllowed = true
	}

	if !isAllowed {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view this person"})
		return
	}

	person, err := c.repository.GetByID(ctx, personID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if person == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
		return
	}

	ctx.JSON(http.StatusOK, person)
}

// GetByRole retrieves persons by role
// @Summary Get persons by role
// @Description Get persons by role
// @Tags persons
// @Accept json
// @Produce json
// @Param role path string true "Role name"
// @Success 200 {array} model.Person
// @Router /persons/role/{role} [get]
func (c *PersonController) GetByRole(ctx *gin.Context) {
	role := ctx.Param("role")
	if role == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Role is required"})
		return
	}

	persons, err := c.repository.GetByRole(ctx, role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, persons)
}

// Create adds a new person
// @Summary Create a new person
// @Description Create a new person. Allowed for Admin and Manager roles.
// @Tags persons
// @Accept json
// @Produce json
// @Param person body model.Person true "Person object"
// @Success 201 {object} model.Person
// @Failure 400 {object} string "Invalid input"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Router /persons [post]
func (c *PersonController) Create(ctx *gin.Context) {
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

	if authUser.Role != "admin" && authUser.Role != "manager" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to create a person"})
		return
	}

	var person model.Person
	if err := ctx.ShouldBindJSON(&person); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if person.ID == uuid.Nil {
		person.ID = uuid.New()
	}

	createdPerson, err := c.repository.Create(ctx, person)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, createdPerson)
}

// Update updates an existing person
// @Summary Update a person
// @Description Update a person. Allowed for Admin role and for users updating their own information.
// @Tags persons
// @Accept json
// @Produce json
// @Param id path string true "Person ID"
// @Param person body model.Person true "Person object"
// @Success 200 {object} model.Person
// @Failure 400 {object} string "Invalid input or ID format"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 404 {object} string "Person not found"
// @Router /persons/{id} [put]
func (c *PersonController) Update(ctx *gin.Context) {
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

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Check if user is authorized:
	// 1. Admins can update any person
	// 2. Users can update their own person record
	isAuthorized := false
	if authUser.Role == "admin" {
		isAuthorized = true
	} else if authUser.PersonID == id {
		// User is updating their own profile
		isAuthorized = true
	}

	if !isAuthorized {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this person"})
		return
	}

	var person model.Person
	if err := ctx.ShouldBindJSON(&person); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	person.ID = id

	updatedPerson, err := c.repository.Update(ctx, person)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedPerson)
}

// Delete removes a person
// @Summary Delete a person
// @Description Delete a person. Allowed for Admin role only. Cascade deletes all associated bank accounts first.
// @Tags persons
// @Accept json
// @Produce json
// @Param id path string true "Person ID"
// @Success 204 "No Content"
// @Failure 400 {object} string "Invalid ID format"
// @Failure 401 {object} string "Unauthorized"
// @Failure 403 {object} string "Forbidden"
// @Failure 500 {object} string "Internal server error"
// @Router /persons/{id} [delete]
func (c *PersonController) Delete(ctx *gin.Context) {
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
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete a person"})
		return
	}

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Check if person exists
	existingPerson, err := c.repository.GetByID(ctx, id)
	if err != nil {
		log.Printf("Error checking if person exists: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking person existence"})
		return
	}
	if existingPerson == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
		return
	}

	// First, delete all associated users (cascade delete)
	log.Printf("Deleting users for person ID: %s", id.String())
	associatedUser, err := c.userRepo.GetByPersonID(ctx, id)
	if err != nil {
		log.Printf("Error retrieving user for person %s: %v", id.String(), err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving associated user"})
		return
	}

	// Delete the associated user if exists
	if associatedUser != nil {
		log.Printf("Deleting user ID: %s for person ID: %s", associatedUser.ID.String(), id.String())
		err = c.userRepo.Delete(ctx, associatedUser.ID)
		if err != nil {
			log.Printf("Error deleting user %s for person %s: %v", associatedUser.ID.String(), id.String(), err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting associated user"})
			return
		}
		log.Printf("Successfully deleted user %s for person %s", associatedUser.ID.String(), id.String())
	} else {
		log.Printf("No associated user found for person %s", id.String())
	}

	// Second, delete all associated bank accounts (cascade delete)
	log.Printf("Deleting bank accounts for person ID: %s", id.String())
	bankAccounts, err := c.bankAccountRepo.GetByPersonID(ctx, id)
	if err != nil {
		log.Printf("Error retrieving bank accounts for person %s: %v", id.String(), err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving associated bank accounts"})
		return
	}

	// Delete each bank account
	for _, account := range bankAccounts {
		log.Printf("Deleting bank account ID: %s for person ID: %s", account.ID.String(), id.String())
		err = c.bankAccountRepo.Delete(ctx, account.ID)
		if err != nil {
			log.Printf("Error deleting bank account %s for person %s: %v", account.ID.String(), id.String(), err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting associated bank account"})
			return
		}
	}

	log.Printf("Successfully deleted %d bank account(s) for person %s", len(bankAccounts), id.String())

	// Now delete the person
	log.Printf("Deleting person with ID: %s", id.String())
	err = c.repository.Delete(ctx, id)
	if err != nil {
		log.Printf("Error deleting person %s: %v", id.String(), err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Successfully deleted person %s and %d associated bank account(s)", id.String(), len(bankAccounts))
	ctx.Status(http.StatusNoContent)
}

// RegisterRoutes registers the routes for the person controller
// Routes registered under the general authenticated group /api
func (c *PersonController) RegisterRoutes(router *gin.RouterGroup) {
	persons := router.Group("/persons")
	{
		persons.GET("", c.GetAll)
		persons.GET("/:id", c.GetByID)
		persons.GET("/role/:role", c.GetByRole)
		persons.POST("", c.Create)
		persons.PUT("/:id", c.Update)
		persons.DELETE("/:id", c.Delete)
	}
}
