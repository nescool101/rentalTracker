package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/storage"
)

// BankAccountController handles HTTP requests for bank account entities
type BankAccountController struct {
	repository *storage.BankAccountRepository
}

// NewBankAccountController creates a new BankAccountController
func NewBankAccountController(repository *storage.BankAccountRepository) *BankAccountController {
	return &BankAccountController{
		repository: repository,
	}
}

// GetAll retrieves all bank accounts
// @Summary Get all bank accounts
// @Description Get all bank accounts
// @Tags bank-accounts
// @Accept json
// @Produce json
// @Success 200 {array} model.BankAccount
// @Router /bank-accounts [get]
func (c *BankAccountController) GetAll(ctx *gin.Context) {
	// Check for admin role
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
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Admin or manager role required to view all bank accounts"})
		return
	}

	accounts, err := c.repository.GetAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if accounts == nil {
		accounts = []model.BankAccount{}
	}

	ctx.JSON(http.StatusOK, accounts)
}

// GetByID retrieves a bank account by ID
// @Summary Get bank account by ID
// @Description Get bank account by ID
// @Tags bank-accounts
// @Accept json
// @Produce json
// @Param id path string true "Bank Account ID"
// @Success 200 {object} model.BankAccount
// @Failure 404 {object} string "Bank account not found"
// @Router /bank-accounts/{id} [get]
func (c *BankAccountController) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Authorization: Check if user is authorized to view this account
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

	// Retrieve account first to check ownership
	account, err := c.repository.GetByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if account == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Bank account not found"})
		return
	}

	// Admin can view any account, others can only view their own
	if authUser.Role != "admin" && authUser.PersonID != account.PersonID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only access your own bank accounts"})
		return
	}

	ctx.JSON(http.StatusOK, account)
}

// GetByPersonID retrieves bank accounts by person ID
// @Summary Get bank accounts by person ID
// @Description Get bank accounts by person ID
// @Tags bank-accounts
// @Accept json
// @Produce json
// @Param personId path string true "Person ID"
// @Success 200 {array} model.BankAccount
// @Router /bank-accounts/person/{personId} [get]
func (c *BankAccountController) GetByPersonID(ctx *gin.Context) {
	personID, err := uuid.Parse(ctx.Param("personId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Authorization: Check if the authenticated user is the person or an admin
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

	if authUser.Role != "admin" && authUser.PersonID != personID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only access your own bank accounts"})
		return
	}

	accounts, err := c.repository.GetByPersonID(ctx, personID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if accounts == nil {
		accounts = []model.BankAccount{}
	}

	ctx.JSON(http.StatusOK, accounts)
}

// Create adds a new bank account
// @Summary Create a new bank account
// @Description Create a new bank account. The owner of the account must be the authenticated user or an admin.
// @Tags bank-accounts
// @Accept json
// @Produce json
// @Param account body model.BankAccount true "Bank Account object"
// @Success 201 {object} model.BankAccount
// @Router /bank-accounts [post]
func (c *BankAccountController) Create(ctx *gin.Context) {
	var account model.BankAccount
	if err := ctx.ShouldBindJSON(&account); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authorization: Only admin can create accounts for others
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

	if authUser.Role != "admin" && authUser.PersonID != account.PersonID {
		log.Printf("User %s (PersonID: %s) attempting to create bank account for PersonID: %s",
			authUser.Email, authUser.PersonID, account.PersonID)
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only create bank accounts for yourself"})
		return
	}

	// Generate a new UUID if not provided
	if account.ID == uuid.Nil {
		account.ID = uuid.New()
	}

	createdAccount, err := c.repository.Create(ctx, account)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, createdAccount)
}

// Update updates an existing bank account
// @Summary Update a bank account
// @Description Update a bank account. The owner of the account must be the authenticated user or an admin.
// @Tags bank-accounts
// @Accept json
// @Produce json
// @Param id path string true "Bank Account ID"
// @Param account body model.BankAccount true "Bank Account object"
// @Success 200 {object} model.BankAccount
// @Failure 404 {object} string "Bank account not found"
// @Router /bank-accounts/{id} [put]
func (c *BankAccountController) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Check if account exists first
	existingAccount, err := c.repository.GetByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if existingAccount == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Bank account not found"})
		return
	}

	// Authorization: Check if user is authorized to update this account
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

	if authUser.Role != "admin" && authUser.PersonID != existingAccount.PersonID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own bank accounts"})
		return
	}

	var account model.BankAccount
	if err := ctx.ShouldBindJSON(&account); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure the ID in the URL matches the ID in the body
	account.ID = id

	// Ensure person ID stays the same (can't change ownership)
	account.PersonID = existingAccount.PersonID

	updatedAccount, err := c.repository.Update(ctx, account)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedAccount)
}

// Delete removes a bank account
// @Summary Delete a bank account
// @Description Delete a bank account. Admin only operation.
// @Tags bank-accounts
// @Accept json
// @Produce json
// @Param id path string true "Bank Account ID"
// @Success 204 "No Content"
// @Router /bank-accounts/{id} [delete]
func (c *BankAccountController) Delete(ctx *gin.Context) {
	// This endpoint is registered under AdminMiddleware, so only admins can access it
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Check if account exists
	existingAccount, err := c.repository.GetByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if existingAccount == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Bank account not found"})
		return
	}

	err = c.repository.Delete(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// RegisterRoutes registers the routes for the bank account controller
func (c *BankAccountController) RegisterRoutes(router *gin.RouterGroup) {
	bankAccounts := router.Group("/bank-accounts")
	{
		bankAccounts.GET("", c.GetAll)
		bankAccounts.GET("/:id", c.GetByID)
		bankAccounts.GET("/person/:personId", c.GetByPersonID)
		bankAccounts.POST("", c.Create)
		bankAccounts.PUT("/:id", c.Update)
		bankAccounts.DELETE("/:id", c.Delete)
	}
}
