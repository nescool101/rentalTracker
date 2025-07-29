package controller

import (
	"encoding/base64"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/storage"
)

// ManagerRegistrationController handles HTTP requests for manager registration
type ManagerRegistrationController struct {
	repositoryFactory *storage.RepositoryFactory
}

// NewManagerRegistrationController creates a new ManagerRegistrationController
func NewManagerRegistrationController(repositoryFactory *storage.RepositoryFactory) *ManagerRegistrationController {
	return &ManagerRegistrationController{
		repositoryFactory: repositoryFactory,
	}
}

// Register handles the registration of a new manager
// @Summary Register a new manager
// @Description Register a new manager with property, pricing, and bank account details
// @Tags manager
// @Accept json
// @Produce json
// @Param registration body model.ManagerRegistrationRequest true "Manager registration data"
// @Success 201 {object} model.ManagerRegistrationResponse
// @Failure 400 {object} string "Bad request"
// @Failure 500 {object} string "Internal server error"
// @Router /register/manager [post]
func (c *ManagerRegistrationController) Register(ctx *gin.Context) {
	var registrationRequest model.ManagerRegistrationRequest

	// Bind the JSON request
	if err := ctx.ShouldBindJSON(&registrationRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get repositories
	personRepo := c.repositoryFactory.GetPersonRepository()
	propertyRepo := c.repositoryFactory.GetPropertyRepository()
	userRepo := c.repositoryFactory.GetUserRepository()
	rentalRepo := c.repositoryFactory.GetRentalRepository()
	pricingRepo := c.repositoryFactory.GetPricingRepository()
	bankAccountRepo := c.repositoryFactory.GetBankAccountRepository()

	// 1. Create Person
	personID := uuid.New()
	person := model.Person{
		ID:       personID,
		FullName: registrationRequest.FullName,
		Phone:    registrationRequest.Phone,
		NIT:      registrationRequest.NIT,
	}

	_, err := personRepo.Create(ctx, person)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating person: " + err.Error()})
		return
	}

	// 2. Create Property
	propertyID := uuid.New()
	property := model.Property{
		ID:        propertyID,
		Address:   registrationRequest.PropertyAddress,
		AptNumber: registrationRequest.PropertyAptNumber,
		City:      registrationRequest.PropertyCity,
		State:     registrationRequest.PropertyState,
		ZipCode:   registrationRequest.PropertyZipCode,
		Type:      registrationRequest.PropertyType,
		// Note: the manager will be the resident at this point
		ResidentID: personID,
		ManagerIDs: []uuid.UUID{personID},
	}

	_, err = propertyRepo.Create(ctx, property)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating property: " + err.Error()})
		return
	}

	// 3. Create Bank Account (in a real implementation)
	// In this simplified version, we're just generating the ID
	// since we don't have a complete BankAccount implementation
	bankAccountID := uuid.New()
	bankAccount := model.BankAccount{
		ID:            bankAccountID,
		PersonID:      personID,
		BankName:      registrationRequest.BankName,
		AccountType:   registrationRequest.AccountType,
		AccountNumber: registrationRequest.AccountNumber,
		AccountHolder: registrationRequest.AccountHolder,
	}

	_, err = bankAccountRepo.Create(ctx, bankAccount)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating bank account: " + err.Error()})
		return
	}

	// 4. Create User with pending status
	userID := uuid.New()

	// Encode the password in base64
	passwordBytes := []byte(registrationRequest.Password)
	passwordBase64 := base64.StdEncoding.EncodeToString(passwordBytes)

	user := model.User{
		ID:             userID,
		Email:          registrationRequest.Email,
		PasswordBase64: passwordBase64,
		Role:           "manager", // Set role to manager
		PersonID:       personID,
		Status:         "pending", // New managers start as pending
	}

	_, err = userRepo.Create(ctx, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user: " + err.Error()})
		return
	}

	// 5. Create Rental (self-rental initially)
	rentalID := uuid.New()

	// Convert normal time.Time to FlexibleTime by type casting
	startDate := model.FlexibleTime(registrationRequest.StartDate)
	endDate := model.FlexibleTime(registrationRequest.EndDate)

	rental := model.Rental{
		ID:            rentalID,
		PropertyID:    propertyID,
		RenterID:      personID, // The manager is initially the renter
		BankAccountID: bankAccountID,
		StartDate:     startDate,
		EndDate:       endDate,
		PaymentTerms:  registrationRequest.PaymentTerms,
		UnpaidMonths:  0,
	}

	_, err = rentalRepo.Create(ctx, rental)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating rental: " + err.Error()})
		return
	}

	// 6. Create Pricing
	pricingID := uuid.New()
	pricing := model.Pricing{
		ID:                   pricingID,
		RentalID:             rentalID,
		MonthlyRent:          registrationRequest.MonthlyRent,
		SecurityDeposit:      registrationRequest.SecurityDeposit,
		UtilitiesIncluded:    registrationRequest.UtilitiesIncluded,
		TenantResponsibleFor: registrationRequest.TenantResponsibleFor,
		LateFee:              registrationRequest.LateFee,
		DueDay:               registrationRequest.DueDay,
	}

	_, err = pricingRepo.Create(ctx, pricing)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating pricing: " + err.Error()})
		return
	}

	// 7. Create PersonRole with role='manager'
	personRoleID := uuid.New()
	// In a real implementation, you'd get the manager role ID from the database
	// Here we're just creating a direct association
	// Convert role string to uuid.UUID
	roleID, err := uuid.Parse("a1f876db-4d06-448c-96c2-40a10cd54b46") // Example role ID
	if err != nil {
		log.Printf("Error parsing role ID: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing role ID"})
		return
	}

	personRole := model.PersonRole{
		ID:       personRoleID,
		PersonID: personID,
		RoleID:   roleID,
	}

	personRoleRepo := c.repositoryFactory.GetPersonRoleRepository()
	_, err = personRoleRepo.Create(ctx, personRole)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating person role: " + err.Error()})
		return
	}

	// 8. Return success response
	response := model.ManagerRegistrationResponse{
		Success:    true,
		UserID:     userID,
		PersonID:   personID,
		PropertyID: propertyID,
		Message:    "Manager registration successful. An administrator will review and approve your account.",
	}

	ctx.JSON(http.StatusCreated, response)
}

// RegisterRoutes sets up the manager registration routes
func (c *ManagerRegistrationController) RegisterRoutes(router *gin.RouterGroup) {
	register := router.Group("/register")
	{
		register.POST("/manager", c.Register)
	}
}
