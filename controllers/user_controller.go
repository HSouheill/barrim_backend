// controllers/user_controller.go
package controllers

import (
	"context"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/HSouheill/barrim_backend/config"
	"github.com/HSouheill/barrim_backend/middleware"
	"github.com/HSouheill/barrim_backend/models"
)

// UserController contains user management logic
type UserController struct {
	DB *mongo.Client
}

// NewUserController creates a new user controller
func NewUserController(db *mongo.Client) *UserController {
	return &UserController{DB: db}
}

// GetProfile handler gets the current user's profile
func (uc *UserController) GetProfile(c echo.Context) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user collection
	collection := config.GetCollection(uc.DB, "users")

	// Get user information from token
	claims := middleware.GetUserFromToken(c)
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid user ID",
		})
	}

	// Find user by ID
	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, models.Response{
				Status:  http.StatusNotFound,
				Message: "User not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to find user",
		})
	}

	// Remove password from response
	user.Password = ""

	// Return user profile
	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Profile retrieved successfully",
		Data:    user,
	})
}

// UpdateLocation handler updates a user's location
func (uc *UserController) UpdateLocation(c echo.Context) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user collection
	collection := config.GetCollection(uc.DB, "users")

	// Get user information from token
	claims := middleware.GetUserFromToken(c)
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid user ID",
		})
	}

	// Parse request body
	var locReq models.UpdateLocationRequest
	if err := c.Bind(&locReq); err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid request body",
		})
	}

	// Validate location
	if locReq.Location == nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Location is required",
		})
	}

	// Update user location
	update := bson.M{
		"$set": bson.M{
			"location":  locReq.Location,
			"updatedAt": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update location",
		})
	}

	if result.MatchedCount == 0 {
		return c.JSON(http.StatusNotFound, models.Response{
			Status:  http.StatusNotFound,
			Message: "User not found",
		})
	}

	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Location updated successfully",
	})
}

// UpdateProfile handler updates a user's profile
func (uc *UserController) UpdateProfile(c echo.Context) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user collection
	collection := config.GetCollection(uc.DB, "users")

	// Get user information from token
	claims := middleware.GetUserFromToken(c)
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid user ID",
		})
	}

	// Parse request body
	var user models.User
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid request body",
		})
	}

	// Build update document
	update := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	// Only allow updating specific fields
	if user.FullName != "" {
		update["$set"].(bson.M)["fullname"] = user.FullName
	}

	if user.Gender != "" {
		update["$set"].(bson.M)["gender"] = user.Gender
	}

	if user.DateOfBirth != "" {
		update["$set"].(bson.M)["dateOfBirth"] = user.DateOfBirth
	}

	if user.CompanyInfo != nil {
		update["$set"].(bson.M)["companyInfo"] = user.CompanyInfo
	}

	if user.Location != nil {
		update["$set"].(bson.M)["location"] = user.Location
	}

	if user.ServiceProviderInfo != nil {
		// If service type is "Other", ensure custom service type is provided
		if user.ServiceProviderInfo.ServiceType == "Other" &&
			user.ServiceProviderInfo.CustomServiceType == "" {
			return c.JSON(http.StatusBadRequest, models.Response{
				Status:  http.StatusBadRequest,
				Message: "Please specify your service type",
			})
		}

		update["$set"].(bson.M)["serviceProviderInfo"] = user.ServiceProviderInfo
	}

	if user.CompanyInfo != nil {
		// If updating Category type to "Other", ensure customCategory is provided
		if user.CompanyInfo.Category == "Other" && user.CompanyInfo.Category == "" {
			return c.JSON(http.StatusBadRequest, models.Response{
				Status:  http.StatusBadRequest,
				Message: "Please specify your Category type",
			})
		}

		update["$set"].(bson.M)["companyInfo"] = user.CompanyInfo
	}

	if user.Location != nil {
		update["$set"].(bson.M)["location"] = user.Location
	}

	// Update user
	result, err := collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update profile",
		})
	}

	if result.MatchedCount == 0 {
		return c.JSON(http.StatusNotFound, models.Response{
			Status:  http.StatusNotFound,
			Message: "User not found",
		})
	}

	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Profile updated successfully",
	})
}

// DeleteUser handler deletes the current user
func (uc *UserController) DeleteUser(c echo.Context) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user collection
	collection := config.GetCollection(uc.DB, "users")

	// Get user information from token
	claims := middleware.GetUserFromToken(c)
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid user ID",
		})
	}

	// Delete user
	result, err := collection.DeleteOne(ctx, bson.M{"_id": userID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to delete user",
		})
	}

	if result.DeletedCount == 0 {
		return c.JSON(http.StatusNotFound, models.Response{
			Status:  http.StatusNotFound,
			Message: "User not found",
		})
	}

	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "User deleted successfully",
	})
}

// GetAllUsers handler returns all users with pagination
func (uc *UserController) GetAllUsers(c echo.Context) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20 // default limit
	}
	skip := (page - 1) * limit

	// Get user collection
	collection := config.GetCollection(uc.DB, "users")

	// Set up options to exclude password field and apply pagination
	opts := options.Find().
		SetProjection(bson.M{"password": 0}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}}) // Sort by creation date, newest first

	// Find all users
	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to fetch users",
		})
	}
	defer cursor.Close(ctx)

	// Decode all users
	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to decode users",
		})
	}

	// Get total count for pagination info
	totalCount, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to count users",
		})
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Return users with pagination info
	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Users retrieved successfully",
		Data: map[string]interface{}{
			"users": users,
			"pagination": map[string]interface{}{
				"totalCount": totalCount,
				"page":       page,
				"limit":      limit,
				"totalPages": totalPages,
			},
		},
	})
}

// UploadCompanyLogo handler uploads a company logo
func (uc *UserController) UploadCompanyLogo(c echo.Context) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user information from token
	claims := middleware.GetUserFromToken(c)
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid user ID",
		})
	}

	// Get user collection
	collection := config.GetCollection(uc.DB, "users")

	// Check if user exists and is a company
	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return c.JSON(http.StatusNotFound, models.Response{
			Status:  http.StatusNotFound,
			Message: "User not found",
		})
	}

	if user.UserType != "company" {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Only company accounts can upload a logo",
		})
	}

	// Get file from form
	file, err := c.FormFile("logo")
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "No file uploaded or invalid file",
		})
	}

	// Validate file type (only images allowed)
	if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Only image files are allowed",
		})
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to open uploaded file",
		})
	}
	defer src.Close()

	// Generate unique filename
	filename := userID.Hex() + "_" + time.Now().Format("20060102150405") + filepath.Ext(file.Filename)

	// Define storage path (you'll need to implement your own storage logic here)
	// This is just an example assuming you have a specific directory for file uploads
	storagePath := "./uploads/logos/" + filename

	// Create destination file
	dst, err := os.Create(storagePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to create destination file",
		})
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err = io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to save uploaded file",
		})
	}

	// Update the user's companyInfo with the logo path
	logoURL := "/uploads/logos/" + filename // The URL from which the logo can be accessed

	// Update user in database
	update := bson.M{
		"$set": bson.M{
			"companyInfo.logo": logoURL,
			"updatedAt":        time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update company logo",
		})
	}

	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Company logo uploaded successfully",
		Data: map[string]string{
			"logoURL": logoURL,
		},
	})
}

// UploadProfilePhoto handler uploads a profile photo for service providers
func (uc *UserController) UploadProfilePhoto(c echo.Context) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user information from token
	claims := middleware.GetUserFromToken(c)
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid user ID",
		})
	}

	// Get user collection
	collection := config.GetCollection(uc.DB, "users")

	// Check if user exists and is a service provider
	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return c.JSON(http.StatusNotFound, models.Response{
			Status:  http.StatusNotFound,
			Message: "User not found",
		})
	}

	if user.UserType != "serviceProvider" {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Only service providers can upload a profile photo",
		})
	}

	// Get file from form
	file, err := c.FormFile("photo")
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "No file uploaded or invalid file",
		})
	}

	// Validate file type (only images allowed)
	if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Only image files are allowed",
		})
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to open uploaded file",
		})
	}
	defer src.Close()

	// Generate unique filename
	filename := userID.Hex() + "_" + time.Now().Format("20060102150405") + filepath.Ext(file.Filename)

	// Define storage path
	storagePath := "./uploads/profiles/" + filename

	// Create destination file
	dst, err := os.Create(storagePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to create destination file",
		})
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err = io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to save uploaded file",
		})
	}

	// Update the user's serviceProviderInfo with the photo path
	photoURL := "/uploads/profiles/" + filename

	// Initialize ServiceProviderInfo if it doesn't exist
	update := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	if user.ServiceProviderInfo == nil {
		update["$set"].(bson.M)["serviceProviderInfo"] = models.ServiceProviderInfo{
			ProfilePhoto: photoURL,
		}
	} else {
		update["$set"].(bson.M)["serviceProviderInfo.profilePhoto"] = photoURL
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update profile photo",
		})
	}

	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Profile photo uploaded successfully",
		Data: map[string]string{
			"photoURL": photoURL,
		},
	})
}

// UpdateAvailability updates a service provider's availability calendar
func (uc *UserController) UpdateAvailability(c echo.Context) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user information from token
	claims := middleware.GetUserFromToken(c)
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid user ID",
		})
	}

	// Get user collection
	collection := config.GetCollection(uc.DB, "users")

	// Check if user exists and is a service provider
	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return c.JSON(http.StatusNotFound, models.Response{
			Status:  http.StatusNotFound,
			Message: "User not found",
		})
	}

	if user.UserType != "serviceProvider" {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Only service providers can update availability",
		})
	}

	// Parse request body
	var availabilityReq struct {
		AvailableDays  []string `json:"availableDays"`
		AvailableHours []string `json:"availableHours"`
	}

	if err := c.Bind(&availabilityReq); err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid request body",
		})
	}

	// Validate availability data
	if len(availabilityReq.AvailableDays) == 0 || len(availabilityReq.AvailableHours) == 0 {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Available days and hours are required",
		})
	}

	// Update the service provider's availability
	update := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	if user.ServiceProviderInfo == nil {
		update["$set"].(bson.M)["serviceProviderInfo"] = models.ServiceProviderInfo{
			AvailableDays:  availabilityReq.AvailableDays,
			AvailableHours: availabilityReq.AvailableHours,
		}
	} else {
		update["$set"].(bson.M)["serviceProviderInfo.availableDays"] = availabilityReq.AvailableDays
		update["$set"].(bson.M)["serviceProviderInfo.availableHours"] = availabilityReq.AvailableHours
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update availability",
		})
	}

	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Availability updated successfully",
	})
}

// SearchServiceProviders allows searching for service providers by type and location
func (uc *UserController) SearchServiceProviders(c echo.Context) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user collection
	collection := config.GetCollection(uc.DB, "users")

	// Get pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20 // default limit
	}
	skip := (page - 1) * limit

	// Get filter parameters
	serviceType := c.QueryParam("serviceType")
	city := c.QueryParam("city")
	country := c.QueryParam("country")

	// Build filter
	filter := bson.M{"userType": "serviceProvider"}

	if serviceType != "" {
		filter["serviceProviderInfo.serviceType"] = serviceType
	}

	if city != "" {
		filter["location.city"] = city
	}

	if country != "" {
		filter["location.country"] = country
	}

	// Set up options to exclude password field and apply pagination
	opts := options.Find().
		SetProjection(bson.M{"password": 0}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	// Find service providers matching the criteria
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to fetch service providers",
		})
	}
	defer cursor.Close(ctx)

	// Decode all service providers
	var providers []models.User
	if err := cursor.All(ctx, &providers); err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to decode service providers",
		})
	}

	// Get total count for pagination info
	totalCount, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to count service providers",
		})
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Return service providers with pagination info
	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Service providers retrieved successfully",
		Data: map[string]interface{}{
			"serviceProviders": providers,
			"pagination": map[string]interface{}{
				"totalCount": totalCount,
				"page":       page,
				"limit":      limit,
				"totalPages": totalPages,
			},
		},
	})
}

// GetCompaniesWithLocations handler returns all companies with location data
func (uc *UserController) GetCompaniesWithLocations(c echo.Context) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user collection
	collection := config.GetCollection(uc.DB, "users")

	// Set up filter for companies with location data
	filter := bson.M{
		"userType": "company",
		"location": bson.M{"$exists": true, "$ne": nil},
	}

	// Set up options to exclude sensitive data
	opts := options.Find().
		SetProjection(bson.M{
			"password":           0,
			"resetPasswordToken": 0,
			"otpInfo":            0,
		})

	// Find companies
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to fetch companies",
		})
	}
	defer cursor.Close(ctx)

	// Decode all companies
	var companies []models.User
	if err := cursor.All(ctx, &companies); err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to decode companies",
		})
	}

	// Return companies
	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Companies retrieved successfully",
		Data:    companies,
	})
}
