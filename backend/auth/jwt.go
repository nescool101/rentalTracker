package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nescool101/rentManager/model"
)

// JWTSecretKey is the secret key used to sign JWT tokens
// In production, this should be set via environment variables
const JWTSecretKey = "your-super-secret-key-change-this-in-production"

// CustomClaims represents the claims in our JWT
type CustomClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	PersonID string `json:"person_id"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(user *model.User) (string, error) {
	// Token expiration time - 24 hours
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create the JWT claims
	claims := CustomClaims{
		UserID:   user.ID.String(),
		Email:    user.Email,
		Role:     user.Role,
		PersonID: user.PersonID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "rentManager",
			Subject:   user.ID.String(),
		},
	}

	// Create the token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(JWTSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*CustomClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Validate the algorithm is what we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(JWTSecretKey), nil
		},
	)

	if err != nil {
		return nil, err
	}

	// Validate the token and return the claims
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ExtractUserFromToken extracts user information from a JWT token
func ExtractUserFromToken(tokenString string) (*model.User, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, err
	}

	personID, err := uuid.Parse(claims.PersonID)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:       userID,
		Email:    claims.Email,
		Role:     claims.Role,
		PersonID: personID,
	}

	return user, nil
}
