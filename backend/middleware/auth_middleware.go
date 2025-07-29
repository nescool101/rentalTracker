package middleware

import (
	"net/http"
	"strings"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/nescool101/rentManager/auth"
	"github.com/nescool101/rentManager/model"
)

// AuthMiddleware validates JWT tokens and adds user information to the request context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if the header has the Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate the token
		user, err := auth.ExtractUserFromToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Check user status - only block disabled accounts
		if user.Status == "disabled" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Your account is disabled. Please contact support."})
			c.Abort()
			return
		}

		// Allow newuser status - they will be redirected by the frontend when appropriate
		log.Printf("User authenticated: %s (Role: %s, Status: %s)", user.Email, user.Role, user.Status)

		// Set the user in the context
		c.Set("user", user)

		// Continue
		c.Next()
	}
}

// AdminMiddleware ensures the user has admin role
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user from the context (set by AuthMiddleware)
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		user, ok := userInterface.(*model.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User data invalid"})
			c.Abort()
			return
		}

		// Check if the user has admin role
		if user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		// Continue
		c.Next()
	}
}

// ManagerMiddleware ensures the user has either admin or manager role
// Also allows managers with 'newuser' status to access specific routes
func ManagerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user from the context (set by AuthMiddleware)
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		user, ok := userInterface.(*model.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User data invalid"})
			c.Abort()
			return
		}

		// Check if the user has either admin or manager role
		if user.Role != "admin" && user.Role != "manager" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Manager or admin access required"})
			c.Abort()
			return
		}

		// Continue
		c.Next()
	}
}
