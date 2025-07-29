package controller

import (
	"net/http"

	"encoding/base64"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"log"

	"github.com/nescool101/rentManager/auth"
	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/storage"
)

// UserController handles HTTP requests for user entities
type UserController struct {
	repository *storage.UserRepository
}

// NewUserController creates a new UserController
func NewUserController(repository *storage.UserRepository) *UserController {
	return &UserController{
		repository: repository,
	}
}

// GetAll retrieves all users
// @Summary Get all users
// @Description Get all users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} model.User
// @Router /users [get]
func (c *UserController) GetAll(ctx *gin.Context) {
	users, err := c.repository.GetAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, users)
}

// GetByID retrieves a user by ID
// @Summary Get user by ID
// @Description Get user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} model.User
// @Failure 404 {object} string "User not found"
// @Router /users/{id} [get]
func (c *UserController) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	user, err := c.repository.GetByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// GetByEmail retrieves a user by email
// @Summary Get user by email
// @Description Get user by email
// @Tags users
// @Accept json
// @Produce json
// @Param email query string true "Email address"
// @Success 200 {object} model.User
// @Failure 404 {object} string "User not found"
// @Router /users/email [get]
func (c *UserController) GetByEmail(ctx *gin.Context) {
	email := ctx.Query("email")
	if email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	user, err := c.repository.GetByEmail(ctx, email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// Login authenticates a user
// @Summary Login user
// @Description Authenticate user by email and password
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body object true "User credentials"
// @Success 200 {object} object
// @Failure 401 {object} string "Authentication failed"
// @Router /users/login [post]
func (c *UserController) Login(ctx *gin.Context) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Login attempt for email: %s", credentials.Email)

	// Get user by email
	user, err := c.repository.GetByEmail(ctx, credentials.Email)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		log.Printf("User not found: %s", credentials.Email)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	log.Printf("User found: %s, Status: %s", credentials.Email, user.Status)

	// Check user status - pending and disabled users cannot log in
	if user.Status == "pending" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "EL ESTADO DE TU USUARIO ES INACTIVO, ESTAMOS ESPERANDO TU PAGO O APROBACION EN EL SISTEMA PARA QUE PUEDAS ACCEDER"})
		return
	}

	if user.Status == "disabled" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Tu cuenta ha sido deshabilitada. Contacta a soporte para más información."})
		return
	}

	// Note: "newuser" status is allowed to login and will be redirected to the stepper component

	// Password checking logic
	// The frontend always sends base64 encoded passwords, so we compare directly with stored password
	log.Printf("Comparing passwords - Received password: %s (length: %d)", credentials.Password, len(credentials.Password))
	log.Printf("Stored password hash: %s (length: %d)", user.PasswordBase64, len(user.PasswordBase64))

	passwordMatch := credentials.Password == user.PasswordBase64

	if !passwordMatch {
		// Just for debugging - can be removed in production
		log.Printf("Password mismatch for user: %s - Comparison returned false", credentials.Email)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	tokenString, err := auth.GenerateToken(user)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	log.Printf("Login successful for user: %s with status: %s", credentials.Email, user.Status)

	// Return user data with success flag
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"role":      user.Role,
			"person_id": user.PersonID,
			"status":    user.Status,
		},
		"token": tokenString,
	})
}

// Helper function to check if a string is base64 encoded
func isBase64Encoded(s string) bool {
	// First try to decode
	_, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Printf("String is not base64 encoded: %v", err)
		return false
	}

	// Additional check: base64 encoded strings have a specific length pattern
	// and only contain a limited set of characters
	if len(s)%4 != 0 {
		log.Printf("String length %d is not divisible by 4, not base64", len(s))
		return false
	}

	// Check for base64 character set
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=') {
			log.Printf("String contains invalid base64 character: %c", c)
			return false
		}
	}

	log.Printf("String appears to be valid base64")
	return true
}

// Create adds a new user
// @Summary Create a new user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body model.User true "User object"
// @Success 201 {object} model.User
// @Router /users [post]
func (c *UserController) Create(ctx *gin.Context) {
	var user model.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate a new UUID if not provided
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	// Set default status to 'pending' if not specified
	if user.Status == "" {
		user.Status = "pending"
	}

	// Handle password encoding
	if user.PasswordBase64 != "" {
		// Only encode if not already encoded
		if !isBase64Encoded(user.PasswordBase64) {
			encodedPassword := base64.StdEncoding.EncodeToString([]byte(user.PasswordBase64))
			user.PasswordBase64 = encodedPassword
		}
	}

	createdUser, err := c.repository.Create(ctx, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, createdUser)
}

// Update modifies an existing user
// @Summary Update a user
// @Description Update a user's information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body model.User true "User object"
// @Success 200 {object} model.User
// @Failure 404 {object} string "User not found"
// @Router /users/{id} [put]
func (c *UserController) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	log.Printf("Attempting to update user with ID: %s", id.String())

	// Check if user exists
	existingUser, err := c.repository.GetByID(ctx, id)
	if err != nil {
		log.Printf("Error retrieving existing user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if existingUser == nil {
		log.Printf("User not found with ID: %s", id.String())
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Bind new data
	var user model.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		log.Printf("Invalid user data: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure ID is consistent
	user.ID = id

	// Handle password - ensure it's properly base64 encoded
	if user.PasswordBase64 != "" {
		log.Printf("Password update requested for user %s", user.Email)

		// For new users or direct password updates, decode and re-encode to ensure proper format
		decodedPassword, err := base64.StdEncoding.DecodeString(user.PasswordBase64)
		if err != nil {
			// If it's not valid base64, treat it as plain text and encode it
			log.Printf("Password is not valid base64, treating as plain text and encoding")
			user.PasswordBase64 = base64.StdEncoding.EncodeToString([]byte(user.PasswordBase64))
		} else {
			// It's valid base64, but let's re-encode it to ensure consistency
			log.Printf("Password is valid base64, re-encoding for consistency")
			user.PasswordBase64 = base64.StdEncoding.EncodeToString(decodedPassword)
		}

		log.Printf("Final password length after processing: %d", len(user.PasswordBase64))
	} else {
		// If no password provided, use the existing one
		log.Printf("No password update requested, keeping existing password")
		user.PasswordBase64 = existingUser.PasswordBase64
	}

	// Update user
	updatedUser, err := c.repository.Update(ctx, user)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("User %s updated successfully", user.Email)
	ctx.JSON(http.StatusOK, updatedUser)
}

// Delete removes a user
// @Summary Delete a user
// @Description Delete a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 204 "No Content"
// @Failure 404 {object} string "User not found"
// @Router /users/{id} [delete]
func (c *UserController) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Check if user exists
	existingUser, err := c.repository.GetByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if existingUser == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Delete user
	err = c.repository.Delete(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// ChangePassword changes a user's password with current password verification
// @Summary Change user password
// @Description Change user password with current password verification
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param passwords body object true "Password change data"
// @Success 200 {object} object
// @Failure 400 {object} string "Bad request"
// @Failure 401 {object} string "Invalid current password"
// @Failure 404 {object} string "User not found"
// @Router /users/{id}/change-password [put]
func (c *UserController) ChangePassword(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var passwordChangeRequest struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	if err := ctx.ShouldBindJSON(&passwordChangeRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Password change request for user ID: %s", id.String())

	// Get existing user
	existingUser, err := c.repository.GetByID(ctx, id)
	if err != nil {
		log.Printf("Error retrieving user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if existingUser == nil {
		log.Printf("User not found with ID: %s", id.String())
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verify current password
	currentPasswordEncoded := base64.StdEncoding.EncodeToString([]byte(passwordChangeRequest.CurrentPassword))
	if currentPasswordEncoded != existingUser.PasswordBase64 {
		log.Printf("Current password verification failed for user: %s", existingUser.Email)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Contraseña actual incorrecta"})
		return
	}

	// Encode new password
	newPasswordEncoded := base64.StdEncoding.EncodeToString([]byte(passwordChangeRequest.NewPassword))

	// Update user with new password
	existingUser.PasswordBase64 = newPasswordEncoded
	updatedUser, err := c.repository.Update(ctx, *existingUser)
	if err != nil {
		log.Printf("Error updating user password: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Password changed successfully for user: %s", existingUser.Email)
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Contraseña actualizada exitosamente",
		"user": gin.H{
			"id":    updatedUser.ID,
			"email": updatedUser.Email,
			"role":  updatedUser.Role,
		},
	})
}

// RegisterRoutes sets up the user routes
func (c *UserController) RegisterRoutes(router *gin.RouterGroup) {
	users := router.Group("/users")
	{
		users.GET("", c.GetAll)
		users.GET("/:id", c.GetByID)
		users.GET("/email", c.GetByEmail)
		users.POST("", c.Create)
		users.PUT("/:id", c.Update)
		users.PUT("/:id/change-password", c.ChangePassword)
		users.DELETE("/:id", c.Delete)
	}
}
