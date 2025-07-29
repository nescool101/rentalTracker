package controller

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/service"
	"github.com/nescool101/rentManager/storage"
)

// ContractController handles contract-related operations
type ContractController struct {
	personRepo   *storage.PersonRepository
	propertyRepo *storage.PropertyRepository
	pricingRepo  *storage.PricingRepository
}

// NewContractController creates a new ContractController
func NewContractController(personRepo *storage.PersonRepository, propertyRepo *storage.PropertyRepository, pricingRepo *storage.PricingRepository) *ContractController {
	return &ContractController{
		personRepo:   personRepo,
		propertyRepo: propertyRepo,
		pricingRepo:  pricingRepo,
	}
}

// GenerateContractRequest defines the request structure for generating a contract
type GenerateContractRequest struct {
	RenterID         string    `json:"renter_id" binding:"required"`
	OwnerID          string    `json:"owner_id" binding:"required"`
	PropertyID       string    `json:"property_id" binding:"required"`
	CoSignerID       string    `json:"cosigner_id"` // Optional
	WitnessID        string    `json:"witness_id"`  // Optional
	StartDate        time.Time `json:"start_date" binding:"required"`
	EndDate          time.Time `json:"end_date" binding:"required"`
	ContractDuration string    `json:"contract_duration"`
	MonthlyRent      float64   `json:"monthly_rent" binding:"required"`
	RequiresDeposit  bool      `json:"requires_deposit"`
	DepositAmount    float64   `json:"deposit_amount"`
	DepositText      string    `json:"deposit_text"`
	AdditionalInfo   string    `json:"additional_info"`
}

// RegisterRoutes registers the contract routes
func (ctrl *ContractController) RegisterRoutes(router *gin.RouterGroup) {
	contractRoutes := router.Group("/contracts")
	{
		contractRoutes.POST("/generate", ctrl.HandleGenerateContract)
	}
}

// HandleGenerateContract generates a contract PDF
func (ctrl *ContractController) HandleGenerateContract(c *gin.Context) {
	var req GenerateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Parse IDs
	renterID, err := uuid.Parse(req.RenterID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid renter ID"})
		return
	}

	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner ID"})
		return
	}

	propertyID, err := uuid.Parse(req.PropertyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid property ID"})
		return
	}

	// Get renter
	renter, err := ctrl.personRepo.GetByID(c, renterID)
	if err != nil {
		log.Printf("Error getting renter: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get renter details"})
		return
	}
	if renter == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Renter not found"})
		return
	}

	// Get owner
	owner, err := ctrl.personRepo.GetByID(c, ownerID)
	if err != nil {
		log.Printf("Error getting owner: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get owner details"})
		return
	}
	if owner == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Owner not found"})
		return
	}

	// Get property
	property, err := ctrl.propertyRepo.GetByID(c, propertyID)
	if err != nil {
		log.Printf("Error getting property: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get property details"})
		return
	}
	if property == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	// Create pricing from request data instead of fetching it
	pricing := model.Pricing{
		ID:              uuid.New(),
		MonthlyRent:     req.MonthlyRent,
		SecurityDeposit: req.DepositAmount,
	}

	// Get cosigner if provided
	var cosigner *model.Person
	if req.CoSignerID != "" {
		coSignerID, err := uuid.Parse(req.CoSignerID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cosigner ID"})
			return
		}

		cosigner, err = ctrl.personRepo.GetByID(c, coSignerID)
		if err != nil {
			log.Printf("Error getting cosigner: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cosigner details"})
			return
		}
		if cosigner == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cosigner not found"})
			return
		}
	}

	// Get witness if provided
	var witness *model.Person
	if req.WitnessID != "" {
		witnessID, err := uuid.Parse(req.WitnessID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid witness ID"})
			return
		}

		witness, err = ctrl.personRepo.GetByID(c, witnessID)
		if err != nil {
			log.Printf("Error getting witness: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get witness details"})
			return
		}
		if witness == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Witness not found"})
			return
		}
	}

	// Create contract data
	contractData := service.ContractPDF{
		Renter:         renter,
		Owner:          owner,
		Property:       property,
		Pricing:        &pricing,
		CoSigner:       cosigner,
		Witness:        witness,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		AdditionalInfo: req.AdditionalInfo,
		CreationDate:   time.Now(),
		DepositText:    req.DepositText,
	}

	// Generate a contract ID
	contractID := uuid.New().String()

	// Generate PDF
	pdfBytes, err := service.GenerateContractPDF(contractData)
	if err != nil {
		log.Printf("Error generating contract PDF: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate contract"})
		return
	}

	// Set response headers for PDF download
	fileName := "contrato_arrendamiento.pdf"
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/pdf")
	c.Header("X-Contract-ID", contractID)
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
