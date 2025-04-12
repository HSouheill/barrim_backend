// middleware/jwt_middleware.go
package middleware

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

// JwtCustomClaims for JWT token
type JwtCustomClaims struct {
	UserID   string `json:"userId"`
	Email    string `json:"email"`
	UserType string `json:"userType"`
	jwt.StandardClaims
}

// GetJWTConfig returns JWT middleware configuration
func GetJWTConfig() middleware.JWTConfig {
	return middleware.JWTConfig{
		Claims:     &JwtCustomClaims{},
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
	}
}

func JWTMiddleware() echo.MiddlewareFunc {
	secret := os.Getenv("JWT_SECRET")

	return echomiddleware.JWTWithConfig(echomiddleware.JWTConfig{
		SigningKey: []byte(secret),
	})
}

// GenerateJWT generates new JWT token
func GenerateJWT(userID, email, userType string) (string, error) {
	// Set expiration time
	expiration := time.Now().Add(time.Hour * 24) // 24 hours

	// Set custom claims
	claims := &JwtCustomClaims{
		userID,
		email,
		userType,
		jwt.StandardClaims{
			ExpiresAt: expiration.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// GetUserFromToken extracts user information from JWT token
func GetUserFromToken(c echo.Context) *JwtCustomClaims {
	user := c.Get("user").(*jwt.Token)

	// Handle both MapClaims and JwtCustomClaims
	if claims, ok := user.Claims.(*JwtCustomClaims); ok {
		return claims
	}

	if mapClaims, ok := user.Claims.(jwt.MapClaims); ok {
		// Convert map claims to your custom claims structure
		return &JwtCustomClaims{
			UserID:   mapClaims["userId"].(string),
			UserType: mapClaims["userType"].(string),
			// Add other fields as needed
		}
	}

	return nil
}

func ExtractUserID(c echo.Context) (string, error) {
	user := c.Get("user").(*jwt.Token)
	if user == nil {
		return "", errors.New("invalid token")
	}

	claims := user.Claims.(jwt.MapClaims)
	// Try to get userId first (as per your JwtCustomClaims structure)
	if userID, ok := claims["userId"].(string); ok {
		return userID, nil
	}
	// Fallback to id if userId isn't found
	if userID, ok := claims["id"].(string); ok {
		return userID, nil
	}

	return "", errors.New("invalid user ID in token")
}
