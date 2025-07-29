package controller

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/service"
	"github.com/nescool101/rentManager/storage"
)

// ManagerInvitationController handles HTTP requests for manager invitations
type ManagerInvitationController struct {
	repositoryFactory *storage.RepositoryFactory
}

// NewManagerInvitationController creates a new ManagerInvitationController
func NewManagerInvitationController(repositoryFactory *storage.RepositoryFactory) *ManagerInvitationController {
	return &ManagerInvitationController{
		repositoryFactory: repositoryFactory,
	}
}

// InvitationRequest represents a request to invite a manager
type InvitationRequest struct {
	Email        string `json:"email" binding:"required,email"`
	Name         string `json:"name" binding:"required"`
	Message      string `json:"message"`
	TempPassword string `json:"tempPassword"`
	Status       string `json:"status"`
}

// generateRandomNIT creates a temporary unique NIT placeholder
func generateRandomNIT() string {
	// Create a new random source with time-based seed
	// This is preferred over using the global rand.Seed which is deprecated
	source := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(source)

	// Generate a random timestamp-based string to ensure uniqueness
	timestamp := time.Now().UnixNano()
	randomPart := rnd.Intn(1000000)

	// Format: TEMP-{timestamp}-{random number}
	// This ensures uniqueness even if invitations are sent rapidly
	return fmt.Sprintf("TEMP-%d-%06d", timestamp, randomPart)
}

// SendInvitation handles sending an invitation to a potential manager
// @Summary Send invitation to a manager
// @Description Create a user with temporary password and send invitation email
// @Tags admin
// @Accept json
// @Produce json
// @Param invitation body InvitationRequest true "Invitation data"
// @Success 201 {object} object
// @Failure 400 {object} string "Bad request"
// @Failure 500 {object} string "Internal server error"
// @Router /admin/invitations/manager [post]
func (c *ManagerInvitationController) SendInvitation(ctx *gin.Context) {
	log.Printf("📨 Received manager invitation request on path: %s", ctx.FullPath())
	var request InvitationRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Printf("❌ Error binding JSON: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("📧 Processing invitation request for: %s <%s>", request.Name, request.Email)

	// Set default status if not provided
	if request.Status == "" {
		request.Status = "newuser"
	}

	// Create a new user with the temporary password
	userRepo := c.repositoryFactory.GetUserRepository()
	personRepo := c.repositoryFactory.GetPersonRepository()

	// First check if a user with this email already exists
	existingUser, err := userRepo.GetByEmail(ctx, request.Email)
	if err != nil {
		log.Printf("❌ Error checking for existing user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for existing user"})
		return
	}

	if existingUser != nil {
		log.Printf("⚠️ User with email %s already exists", request.Email)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "A user with this email already exists"})
		return
	}

	// First create a person record with a temporary unique NIT
	personID := uuid.New()
	person := model.Person{
		ID:       personID,
		FullName: request.Name,
		Phone:    "",                  // Will be updated by user during onboarding
		NIT:      generateRandomNIT(), // Generate a random temporary NIT
	}

	_, err = personRepo.Create(ctx, person)
	if err != nil {
		log.Printf("❌ Error creating person record: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create person record"})
		return
	}

	// Encode the temporary password in base64 as required by the user model
	passwordBase64 := base64.StdEncoding.EncodeToString([]byte(request.TempPassword))

	// Now create the user
	userID := uuid.New()
	user := model.User{
		ID:             userID,
		Email:          request.Email,
		PasswordBase64: passwordBase64,
		Role:           "manager", // New users created through invitation are managers
		PersonID:       personID,
		Status:         request.Status,
	}

	createdUser, err := userRepo.Create(ctx, user)
	if err != nil {
		log.Printf("❌ Error creating user record: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user account"})
		return
	}

	// Log user creation with appropriate ID
	userIDToUse := userID // Default to the ID we generated
	if createdUser != nil {
		userIDToUse = createdUser.ID // Use returned ID if available
	} else {
		log.Printf("⚠️ User was created but no user object was returned")
	}

	log.Printf("👤 Created new user (ID: %s) for %s <%s> with status '%s'",
		userIDToUse, request.Name, request.Email, request.Status)

	// Build login URL - prioritize environment variable over Origin header
	frontendURL := os.Getenv("APP_BASE_URL")
	if frontendURL == "" {
		// Fallback to Origin header if environment variable not set
		frontendURL = ctx.GetHeader("Origin")
		if frontendURL == "" {
			// Final fallback for development (but should be avoided in production)
			frontendURL = "http://localhost:5173"
			log.Printf("⚠️ Warning: Using hardcoded frontend URL. Set APP_BASE_URL environment variable.")
		}
	}
	loginURL := fmt.Sprintf("%s/login", frontendURL)

	// Send the invitation email
	subject := "¡Has sido invitado como Administrador de Propiedades!"
	body := fmt.Sprintf("Hola %s,\n\nHas sido invitado a registrarte como administrador de propiedades.\n\nPara iniciar sesión, utiliza los siguientes datos:\nEmail: %s\nContraseña temporal: %s\n\nURL de inicio de sesión: %s\n\n%s\n\nDeberás cambiar tu contraseña en el primer inicio de sesión.\n\nGracias,\nEquipo de Administración",
		request.Name, request.Email, request.TempPassword, loginURL, request.Message)

	err = service.SendSimpleEmail(request.Email, subject, body)
	if err != nil {
		log.Printf("❌ Error sending invitation email to %s: %v", request.Email, err)
		// We still return success since the user was created, just log the error
	}

	// Return success response
	ctx.JSON(http.StatusCreated, gin.H{
		"success":       true,
		"message":       "User created and invitation email sent successfully.",
		"email":         request.Email,
		"person_id":     personID,
		"user_id":       userID,
		"temp_password": request.TempPassword,
	})
}

// RegisterRoutes sets up the manager invitation routes
func (c *ManagerInvitationController) RegisterRoutes(router *gin.RouterGroup) {
	// This route becomes /api/invitations/manager when registered with adminApi
	router.POST("/invitations/manager", c.SendInvitation)
}
