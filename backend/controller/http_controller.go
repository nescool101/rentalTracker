package controller

import (
	"log" // Standard Go log package
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nescool101/rentManager/middleware"
	"github.com/nescool101/rentManager/service"
	"github.com/nescool101/rentManager/storage"
)

func StartHTTPServer() error {
	router := gin.Default()

	// Configure CORS to allow requests from the frontend
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true // In production, you should specify your frontend domain
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	// Initialize Supabase client
	supabaseClient, err := storage.InitializeSupabaseClient()
	if err != nil {
		return err
	}

	// Create repository factory
	repoFactory := storage.NewRepositoryFactory(supabaseClient)

	// Create controllers
	personRepo := repoFactory.GetPersonRepository()
	propertyRepo := repoFactory.GetPropertyRepository()
	rentalRepo := repoFactory.GetRentalRepository()
	userRepo := repoFactory.GetUserRepository()
	rentPaymentRepo := repoFactory.GetRentPaymentRepository()
	rentalHistoryRepo := repoFactory.GetRentalHistoryRepository()
	maintRepo := repoFactory.GetMaintenanceRequestRepository()
	pricingRepo := repoFactory.GetPricingRepository()
	bankAccountRepo := repoFactory.GetBankAccountRepository()

	personController := NewPersonController(personRepo, propertyRepo, rentalRepo, bankAccountRepo, userRepo)
	propertyController := NewPropertyController(propertyRepo)
	rentalController := NewRentalController(rentalRepo, propertyRepo)
	userController := NewUserController(userRepo)
	rentPaymentController := NewRentPaymentController(rentPaymentRepo, rentalRepo, propertyRepo)
	rentalHistoryController := NewRentalHistoryController(rentalHistoryRepo, rentalRepo, propertyRepo, personRepo)
	maintenanceRequestController := NewMaintenanceRequestController(maintRepo, propertyRepo, rentalRepo)
	pricingController := NewPricingController(pricingRepo)
	emailController := NewEmailController(userRepo, personRepo, rentalRepo, propertyRepo)
	contractController := NewContractController(personRepo, propertyRepo, pricingRepo)
	signingRepo := repoFactory.GetContractSigningRepository()
	contractSigningController := NewContractSigningController(personRepo, propertyRepo, pricingRepo, userRepo, contractController, signingRepo)
	managerRegistrationController := NewManagerRegistrationController(repoFactory)
	managerInvitationController := NewManagerInvitationController(repoFactory)
	bankAccountController := NewBankAccountController(bankAccountRepo)
	fileUploadController := NewFileUploadController(userRepo, personRepo)

	// Public API routes (no auth required)
	publicApi := router.Group("/api")
	{
		// Public routes - login doesn't require authentication
		users := publicApi.Group("/users")
		users.POST("/login", userController.Login)

		// Public contract signing routes
		contractSigningController.RegisterPublicRoutes(publicApi)

		// Public file upload routes (with token validation)
		fileUploadController.RegisterPublicRoutes(publicApi)
	}

	// Protected API routes (requires authentication)
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// Register user routes
		userController.RegisterRoutes(api)

		// Register person routes
		personController.RegisterRoutes(api)

		// Register property routes (only GET routes for authenticated users)
		properties := api.Group("/properties")
		{
			properties.GET("", propertyController.GetAll)
			properties.GET("/:id", propertyController.GetByID)
			properties.GET("/resident/:residentId", propertyController.GetByResident)
			properties.GET("/manager/:managerId", propertyController.GetByManagerID)
			properties.GET("/user/:userId", propertyController.GetByUserID)
			properties.POST("", propertyController.Create)
		}

		// Register rental routes - partial access for all users
		// Create, Update, Delete are admin only (defined below)
		rentals := api.Group("/rentals")
		rentals.GET("", rentalController.GetAll)
		rentals.GET("/:id", rentalController.GetByID)
		rentals.GET("/by-property/:property_id", rentalController.GetByPropertyID)
		rentals.GET("/by-renter/:renter_id", rentalController.GetByRenterID)

		// Register maintenance request routes
		maintenanceRequests := api.Group("/maintenance-requests")
		{
			maintenanceRequests.GET("", maintenanceRequestController.GetAll)
			maintenanceRequests.GET("/:id", maintenanceRequestController.GetByID)
			maintenanceRequests.GET("/property/:propertyId", maintenanceRequestController.GetByPropertyID)
			maintenanceRequests.POST("/property-ids", maintenanceRequestController.GetByPropertyIDs)
			maintenanceRequests.GET("/renter/:renterId", maintenanceRequestController.GetByRenterID)
			maintenanceRequests.GET("/status/:status", maintenanceRequestController.GetByStatus)
			maintenanceRequests.POST("", maintenanceRequestController.Create)
			// Note: Update and Delete are admin-only, registered below
		}

		// Register rent payment routes
		payments := api.Group("/payments")
		{
			payments.GET("", rentPaymentController.GetAll)
			payments.GET("/:id", rentPaymentController.GetByID)
			payments.GET("/rental/:rentalId", rentPaymentController.GetByRentalID)
			payments.GET("/rental-ids", rentPaymentController.GetByRentalIDs)
			// Note: date-range and late endpoints are admin-only, registered below
		}

		// Register rental history routes
		rentalHistoryController.RegisterRoutes(api)

		// Register bank account routes
		bankAccountController.RegisterRoutes(api)

		// Register manager registration route - now requires authentication
		managerRegistrationController.RegisterRoutes(api)

		// Register authenticated file upload routes (for regular users)
		fileUploadController.RegisterAuthenticatedUploadRoutes(api)

		// Admin-only routes
		adminApi := api.Group("/admin")
		adminApi.Use(middleware.AdminMiddleware())
		{
			// Admin-only property endpoints
			adminProperties := adminApi.Group("/properties")
			{
				adminProperties.PUT("/:id", propertyController.Update)
				adminProperties.DELETE("/:id", propertyController.Delete)
			}

			// Admin-only payment endpoints
			adminPayments := adminApi.Group("/payments")
			{
				adminPayments.GET("/date-range", rentPaymentController.GetByDateRange)
				adminPayments.GET("/late", rentPaymentController.GetLatePayments)
				adminPayments.POST("", rentPaymentController.Create)
				adminPayments.PUT("/:id", rentPaymentController.Update)
				adminPayments.DELETE("/:id", rentPaymentController.Delete)
			}

			// Admin-only rental history CUD endpoints
			adminRentalHistory := adminApi.Group("/rental-history")
			{
				// GET /status and /date-range are handled by rentalHistoryController.RegisterRoutes(api)
				// and have internal admin checks. No need to duplicate here.
				adminRentalHistory.POST("", rentalHistoryController.Create)
				adminRentalHistory.PUT("/:id", rentalHistoryController.Update)
				adminRentalHistory.DELETE("/:id", rentalHistoryController.Delete)
			}

			// Admin-only maintenance request endpoints
			adminMaintenanceRequests := adminApi.Group("/maintenance-requests")
			{
				adminMaintenanceRequests.PUT("/:id", maintenanceRequestController.Update)
				adminMaintenanceRequests.DELETE("/:id", maintenanceRequestController.Delete)
			}

			// Admin-only Rental CUD routes
			adminRentals := adminApi.Group("/rentals")
			{
				adminRentals.POST("", rentalController.Create)
				adminRentals.PUT("/:id", rentalController.Update)
				adminRentals.DELETE("/:id", rentalController.Delete)
			}

			// Admin-only Pricing CUD routes
			pricingRoutesGroup := adminApi.Group("/pricing")
			pricingController.RegisterRoutes(pricingRoutesGroup)

			// Admin-only Email routes
			emailController.RegisterRoutes(adminApi)

			// Admin-only Contract routes
			contractController.RegisterRoutes(adminApi)

			// Admin-only Contract Signing routes that require authentication
			contractSigningController.RegisterAuthRoutes(adminApi)

			// Admin-only Manager Invitation routes - explicitly set up without using RegisterRoutes
			adminApi.POST("/invitations/manager", managerInvitationController.SendInvitation)

			// Admin-only File Upload routes (for generating upload links)
			fileUploadController.RegisterRoutes(adminApi)

		}
	}

	// Serve static files (frontend)
	router.Static("/assets", "./static/assets")
	router.StaticFile("/", "./static/index.html")
	router.NoRoute(func(c *gin.Context) {
		// Serve index.html for any non-API routes (SPA routing)
		if !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.File("./static/index.html")
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
		}
	})

	// Add health check endpoint
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Legacy routes (temporary, should be migrated)
	router.GET("/payers", getPayers)
	router.GET("/validate_email", validateEmailHandler(repoFactory))

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	return router.Run(":" + port)
}

func getPayers(c *gin.Context) {
	payers := service.GetAllPayers()
	c.JSON(http.StatusOK, gin.H{"payers": payers})
}

func validateEmailHandler(repoFactory *storage.RepositoryFactory) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("ℹ️ [API] /validate_email endpoint triggered.")
		personRepo := repoFactory.GetPersonRepository()
		rentalRepo := repoFactory.GetRentalRepository()
		propertyRepo := repoFactory.GetPropertyRepository()
		userRepo := repoFactory.GetUserRepository()
		pricingRepo := repoFactory.GetPricingRepository()

		// Run NotifyAll in a goroutine so it doesn't block the HTTP response
		go service.NotifyAll(personRepo, rentalRepo, propertyRepo, userRepo, pricingRepo)

		c.JSON(http.StatusOK, gin.H{"status": "Email notification process triggered in background."})
	}
}
