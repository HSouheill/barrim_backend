// controllers/auth_controller.go
package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/HSouheill/barrim_backend/config"
	"github.com/HSouheill/barrim_backend/middleware"
	"github.com/HSouheill/barrim_backend/models"
	"github.com/HSouheill/barrim_backend/utils"
)

// AuthController contains authentication logic
type AuthController struct {
	DB *mongo.Client
}

// NewAuthController creates a new auth controller
func NewAuthController(db *mongo.Client) *AuthController {
	return &AuthController{DB: db}
}

// Signup handler
func (ac *AuthController) Signup(c echo.Context) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user collection
	collection := config.GetCollection(ac.DB, "users")

	// Check if the request is multipart
	contentType := c.Request().Header.Get("Content-Type")
	var signupReq models.SignupRequest
	var logoPath string // Path where the logo will be saved

	// Handle multipart form data (for file uploads)
	if contentType != "" && len(contentType) >= 9 && contentType[:9] == "multipart" {
		// Parse multipart form with 10MB max memory
		if err := c.Request().ParseMultipartForm(10 << 20); err != nil {
			return c.JSON(http.StatusBadRequest, models.Response{
				Status:  http.StatusBadRequest,
				Message: "Failed to parse multipart form: " + err.Error(),
			})
		}

		// Get the data field and unmarshal it
		dataField := c.FormValue("data")
		if dataField == "" {
			return c.JSON(http.StatusBadRequest, models.Response{
				Status:  http.StatusBadRequest,
				Message: "Missing data field in multipart form",
			})
		}

		if err := json.Unmarshal([]byte(dataField), &signupReq); err != nil {
			return c.JSON(http.StatusBadRequest, models.Response{
				Status:  http.StatusBadRequest,
				Message: "Invalid JSON in data field: " + err.Error(),
			})
		}

		// Handle file upload if present
		file, fileHeader, err := c.Request().FormFile("logo")
		if err == nil && fileHeader != nil {
			defer file.Close()

			// Create uploads directory if it doesn't exist
			uploadsDir := "uploads"
			if err := os.MkdirAll(uploadsDir, 0755); err != nil {
				return c.JSON(http.StatusInternalServerError, models.Response{
					Status:  http.StatusInternalServerError,
					Message: "Failed to create uploads directory: " + err.Error(),
				})
			}

			// Generate unique filename
			fileExt := filepath.Ext(fileHeader.Filename)
			newFilename := primitive.NewObjectID().Hex() + fileExt
			logoPath = filepath.Join(uploadsDir, newFilename)

			// Create the file
			dst, err := os.Create(logoPath)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, models.Response{
					Status:  http.StatusInternalServerError,
					Message: "Failed to create file: " + err.Error(),
				})
			}
			defer dst.Close()

			// Copy the uploaded file to the destination
			if _, err := io.Copy(dst, file); err != nil {
				return c.JSON(http.StatusInternalServerError, models.Response{
					Status:  http.StatusInternalServerError,
					Message: "Failed to save file: " + err.Error(),
				})
			}

			log.Printf("File uploaded successfully: %s", logoPath)
		}
	} else {
		// Regular JSON binding for non-multipart requests
		if err := c.Bind(&signupReq); err != nil {
			return c.JSON(http.StatusBadRequest, models.Response{
				Status:  http.StatusBadRequest,
				Message: "Invalid request body",
			})
		}
	}

	// Log signup request details
	log.Printf("\n=== New %s Signup Request ===\n", signupReq.UserType)
	log.Printf("Email: %s\n", signupReq.Email)
	log.Printf("Full Name: %s\n", signupReq.FullName)
	log.Printf("Phone: %s\n", signupReq.Phone)
	log.Printf("User Type: %s\n", signupReq.UserType)

	// Log specific fields based on user type
	switch signupReq.UserType {
	case "company":
		if signupReq.CompanyInfo != nil {
			log.Printf("Company Name: %s\n", signupReq.CompanyInfo.Name)
			log.Printf("Category Type: %s\n", signupReq.CompanyInfo.Category)
			if signupReq.CompanyInfo.CustomCategory != "" {
				log.Printf("Custom Category: %s\n", signupReq.CompanyInfo.CustomCategory)
			}
			if logoPath != "" {
				log.Printf("Logo Path: %s\n", logoPath)
			}
		}
		if signupReq.Location != nil {
			log.Printf("Location: %s, %s, %s\n",
				signupReq.Location.City,
				signupReq.Location.Country,
				signupReq.Location.PostalCode)
		}

	case "wholesaler":
		if signupReq.WholesalerInfo != nil {
			log.Printf("Business Name: %s\n", signupReq.WholesalerInfo.BusinessName)
			log.Printf("Category Type: %s\n", signupReq.WholesalerInfo.Category)
			if signupReq.WholesalerInfo.ReferralCode != "" {
				log.Printf("Referral Code: %s\n", signupReq.WholesalerInfo.ReferralCode)
			}
			if logoPath != "" {
				log.Printf("Logo Path: %s\n", logoPath)
			}
		}
		if signupReq.Location != nil {
			log.Printf("Location: %s, %s, %s\n",
				signupReq.Location.City,
				signupReq.Location.Country,
				signupReq.Location.PostalCode)
		}

	case "serviceProvider":
		if signupReq.ServiceProviderInfo != nil {
			log.Printf("Service Type: %s\n", signupReq.ServiceProviderInfo.ServiceType)
			if signupReq.ServiceProviderInfo.CustomServiceType != "" {
				log.Printf("Custom Service Type: %s\n", signupReq.ServiceProviderInfo.CustomServiceType)
			}
			log.Printf("Years Experience: %d\n", signupReq.ServiceProviderInfo.YearsExperience)
		}
		if signupReq.Location != nil {
			log.Printf("Location: %s, %s, %s\n",
				signupReq.Location.City,
				signupReq.Location.Country,
				signupReq.Location.PostalCode)
		}
	}

	fmt.Printf("=============================\n\n")

	// Validate required fields
	if signupReq.Email == "" || signupReq.Password == "" || signupReq.FullName == "" ||
		signupReq.UserType == "" {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Missing required fields",
		})
	}

	// Validate user type
	validUserTypes := map[string]bool{
		"user":            true,
		"company":         true,
		"wholesaler":      true,
		"serviceProvider": true,
	}
	if !validUserTypes[signupReq.UserType] {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid user type",
		})
	}

	// Additional validations depending on user type...
	// (existing validation code unchanged)

	// Check if user already exists
	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{"email": signupReq.Email}).Decode(&existingUser)
	if err == nil {
		return c.JSON(http.StatusConflict, models.Response{
			Status:  http.StatusConflict,
			Message: "User with this email already exists",
		})
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(signupReq.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to hash password",
		})
	}

	// Create new user
	now := time.Now()
	newUser := models.User{
		Email:               signupReq.Email,
		Password:            hashedPassword,
		FullName:            signupReq.FullName,
		UserType:            signupReq.UserType,
		DateOfBirth:         signupReq.DateOfBirth,
		Gender:              signupReq.Gender,
		Phone:               signupReq.Phone,
		ReferralCode:        signupReq.ReferralCode,
		InterestedDeals:     signupReq.InterestedDeals,
		Location:            signupReq.Location,
		CompanyInfo:         signupReq.CompanyInfo,
		ServiceProviderInfo: signupReq.ServiceProviderInfo,
		WholesalerInfo:      signupReq.WholesalerInfo,
		LogoPath:            logoPath, // Add the logo path to the user document
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	// Insert user to database
	result, err := collection.InsertOne(ctx, newUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to create user: " + err.Error(),
		})
	}

	// Generate JWT token
	token, err := middleware.GenerateJWT(result.InsertedID.(primitive.ObjectID).Hex(), newUser.Email, newUser.UserType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to generate token",
		})
	}

	// Return the token and user info
	return c.JSON(http.StatusCreated, models.Response{
		Status:  http.StatusCreated,
		Message: "User created successfully",
		Data: map[string]interface{}{
			"token": token,
			"user": map[string]interface{}{
				"id":       result.InsertedID,
				"email":    newUser.Email,
				"fullName": newUser.FullName,
				"userType": newUser.UserType,
				"logoPath": logoPath, // Include logo path in response
			},
		},
	})
}

// Login handler
func (ac *AuthController) Login(c echo.Context) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user collection
	collection := config.GetCollection(ac.DB, "users")

	// Parse request body
	var loginReq models.LoginRequest
	if err := c.Bind(&loginReq); err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid request body",
		})
	}

	// Find user by email
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": loginReq.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusUnauthorized, models.Response{
				Status:  http.StatusUnauthorized,
				Message: "Invalid email or password",
			})
		}
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to find user",
		})
	}

	// Check password
	err = utils.CheckPassword(loginReq.Password, user.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.Response{
			Status:  http.StatusUnauthorized,
			Message: "Invalid email or password",
		})
	}

	// Generate JWT token
	token, err := middleware.GenerateJWT(user.ID.Hex(), user.Email, user.UserType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to generate token",
		})
	}

	// Return the token and user info
	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Login successful",
		Data: map[string]interface{}{
			"token": token,
			"user": map[string]interface{}{
				"id":       user.ID,
				"email":    user.Email,
				"fullName": user.FullName,
				"userType": user.UserType,
			},
		},
	})
}

// GoogleUser represents the user data received from Google authentication
type GoogleUser struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	PhotoURL    string `json:"photoURL"`
	UID         string `json:"uid"`
}

// GoogleLogin handles Google authentication
func (ac *AuthController) GoogleLogin(c echo.Context) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user collection
	collection := config.GetCollection(ac.DB, "users")

	// Parse request body
	var googleUser GoogleUser
	if err := c.Bind(&googleUser); err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if googleUser.Email == "" || googleUser.UID == "" {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Email and UID are required",
		})
	}

	// Check if user exists
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": googleUser.Email}).Decode(&user)

	// Initialize userData for response
	var userData map[string]interface{}

	if err != nil {
		if err == mongo.ErrNoDocuments {
			// User doesn't exist, create new user
			now := time.Now()
			newUser := models.User{
				Email:      googleUser.Email,
				FullName:   googleUser.DisplayName,
				UserType:   "user", // Default user type
				GoogleUID:  googleUser.UID,
				ProfilePic: googleUser.PhotoURL,
				CreatedAt:  now,
				UpdatedAt:  now,
			}

			// Insert user to database
			result, err := collection.InsertOne(ctx, newUser)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, models.Response{
					Status:  http.StatusInternalServerError,
					Message: "Failed to create user",
				})
			}

			// Generate JWT token
			token, err := middleware.GenerateJWT(result.InsertedID.(primitive.ObjectID).Hex(), newUser.Email, newUser.UserType)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, models.Response{
					Status:  http.StatusInternalServerError,
					Message: "Failed to generate token",
				})
			}

			// Set user data for response
			userData = map[string]interface{}{
				"token": token,
				"user": map[string]interface{}{
					"id":       result.InsertedID,
					"email":    newUser.Email,
					"fullName": newUser.FullName,
					"userType": newUser.UserType,
				},
			}
		} else {
			return c.JSON(http.StatusInternalServerError, models.Response{
				Status:  http.StatusInternalServerError,
				Message: "Database error",
			})
		}
	} else {
		// User exists, update Google info
		update := bson.M{
			"$set": bson.M{
				"googleUID":  googleUser.UID,
				"fullName":   googleUser.DisplayName,
				"profilePic": googleUser.PhotoURL,
				"updatedAt":  time.Now(),
			},
		}

		_, err = collection.UpdateOne(ctx, bson.M{"email": googleUser.Email}, update)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.Response{
				Status:  http.StatusInternalServerError,
				Message: "Failed to update user",
			})
		}

		// Generate JWT token
		token, err := middleware.GenerateJWT(user.ID.Hex(), user.Email, user.UserType)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.Response{
				Status:  http.StatusInternalServerError,
				Message: "Failed to generate token",
			})
		}

		// Set user data for response
		userData = map[string]interface{}{
			"token": token,
			"user": map[string]interface{}{
				"id":       user.ID,
				"email":    user.Email,
				"fullName": user.FullName,
				"userType": user.UserType,
			},
		}
	}

	// Return success response
	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Login successful",
		Data:    userData,
	})
}

// ServeImage serves image files from the uploads directory
func ServeImage(c echo.Context) error {
	// Get the image path from URL parameter
	filename := c.Param("filename")

	// Sanitize the filename to prevent directory traversal attacks
	filename = filepath.Base(filename)

	// Construct the full path to the image
	path := filepath.Join("uploads", filename)

	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Image not found",
		})
	}

	// Serve the file
	return c.File(path)
}
