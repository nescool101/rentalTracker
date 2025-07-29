package controller

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nescool101/rentManager/service"
	"github.com/nescool101/rentManager/storage"
)

// EmailController handles HTTP requests for sending emails
type EmailController struct {
	userRepo     *storage.UserRepository
	personRepo   *storage.PersonRepository
	rentalRepo   *storage.RentalRepository
	propertyRepo *storage.PropertyRepository
}

// NewEmailController creates a new EmailController
func NewEmailController(userRepo *storage.UserRepository, personRepo *storage.PersonRepository, rentalRepo *storage.RentalRepository, propertyRepo *storage.PropertyRepository) *EmailController {
	return &EmailController{
		userRepo:     userRepo,
		personRepo:   personRepo,
		rentalRepo:   rentalRepo,
		propertyRepo: propertyRepo,
	}
}

// CustomEmailRequest defines the structure for the custom email request body
type CustomEmailRequest struct {
	RecipientPersonID string `json:"recipient_person_id" binding:"required"`
	Subject           string `json:"subject" binding:"required"`
	Body              string `json:"body" binding:"required"`
}

// AnnualRenewalRequest defines the structure for the annual renewal trigger
type AnnualRenewalRequest struct {
	OptionalMessage string `json:"optional_message"`
}

// RegisterRoutes sets up the email routes for an admin-protected group
// It expects an adminRouter, e.g., /api/admin, to which it will add /emails
func (ctrl *EmailController) RegisterRoutes(adminRouter *gin.RouterGroup) {
	emailRoutes := adminRouter.Group("/emails")
	{
		emailRoutes.POST("/custom", ctrl.HandleSendCustomEmail)                                 // POST /api/admin/emails/custom (if adminRouter is /api/admin)
		emailRoutes.POST("/annual-renewal-reminders", ctrl.HandleTriggerAnnualRenewalReminders) // New route
	}
}

// HandleSendCustomEmail sends a custom email to a specified person
func (ctrl *EmailController) HandleSendCustomEmail(ctx *gin.Context) {
	var req CustomEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	recipientPersonUUID, err := uuid.Parse(req.RecipientPersonID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipient_person_id format"})
		return
	}

	// Fetch the user record for the recipient to get their email
	recipientUser, err := ctrl.userRepo.GetByPersonID(ctx, recipientPersonUUID)
	if err != nil {
		log.Printf("Error fetching user for person ID %s: %v", req.RecipientPersonID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve recipient details"})
		return
	}
	if recipientUser == nil || recipientUser.Email == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Recipient user not found or email is missing"})
		return
	}

	if req.Subject == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Subject cannot be empty"})
		return
	}
	if req.Body == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Body cannot be empty"})
		return
	}

	// Call the email service function
	err = service.SendSimpleEmail(recipientUser.Email, req.Subject, req.Body)
	if err != nil {
		log.Printf("Error sending custom email to %s: %v", recipientUser.Email, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		return
	}

	log.Printf("Custom email sent successfully to %s (PersonID: %s)", recipientUser.Email, req.RecipientPersonID)
	ctx.JSON(http.StatusOK, gin.H{"message": "Custom email sent successfully to " + recipientUser.Email})
}

// HandleTriggerAnnualRenewalReminders triggers the process of sending annual renewal reminders.
func (ctrl *EmailController) HandleTriggerAnnualRenewalReminders(ctx *gin.Context) {
	var req AnnualRenewalRequest
	// BindJSON will bind an empty struct if the body is empty or not JSON, which is fine for optional fields.
	if err := ctx.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" { // Allow empty body for no optional message
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Call the service function in a goroutine so it doesn't block the HTTP response.
	go func() {
		// Create a new background context for the goroutine
		bgCtx := context.Background()
		emailsSent, err := service.SendAnnualRenewalReminders(bgCtx, ctrl.personRepo, ctrl.rentalRepo, ctrl.propertyRepo, ctrl.userRepo, req.OptionalMessage)
		if err != nil {
			log.Printf("❌ [ERROR] HandleTriggerAnnualRenewalReminders: Error in service call: %v", err)
			// Since this is a background task, we can't directly return an HTTP error for this failure.
			// Logging is the primary way to observe issues here.
		} else {
			log.Printf("ℹ️ [INFO] HandleTriggerAnnualRenewalReminders: Service call completed. Emails sent: %d", emailsSent)
		}
	}()

	ctx.JSON(http.StatusAccepted, gin.H{"message": "Annual renewal reminder process triggered in the background. Check server logs for details."})
}
