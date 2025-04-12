// controllers/company_controller.go
package controllers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/HSouheill/barrim_backend/config"
	"github.com/HSouheill/barrim_backend/middleware"
	"github.com/HSouheill/barrim_backend/models"
	"github.com/google/uuid"
)

// CompanyController handles company-related operations
type CompanyController struct {
	DB *mongo.Client
}

// NewCompanyController creates a new company controller
func NewCompanyController(db *mongo.Client) *CompanyController {
	return &CompanyController{DB: db}
}

// GetCompanyData retrieves company data including category and subcategory
func (cc *CompanyController) GetCompanyData(c echo.Context) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user collection
	collection := config.GetCollection(cc.DB, "users")

	// Get user information from token
	claims := middleware.GetUserFromToken(c)
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid user ID",
		})
	}

	// Find user by ID and check if it's a company
	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": userID, "userType": "company"}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, models.Response{
				Status:  http.StatusNotFound,
				Message: "Company not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to find company",
		})
	}

	// Ensure CompanyInfo is not nil
	if user.CompanyInfo == nil {
		user.CompanyInfo = &models.CompanyInfo{}
	}

	// Return company data with category and subcategory
	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Company data retrieved successfully",
		Data: map[string]interface{}{
			"companyInfo": map[string]interface{}{
				"name":        user.CompanyInfo.Name,
				"Category":    user.CompanyInfo.Category,
				"subCategory": user.CompanyInfo.SubCategory,
				"logo":        user.CompanyInfo.Logo,
			},
			"location": user.Location,
		},
	})
}

// CreateBranch handles the creation of a new branch with images
func (cc *CompanyController) CreateBranch(c echo.Context) error {
	// Add request logging at the start
	log.Printf("Starting branch creation request from IP: %s", c.RealIP())

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Increased timeout
	defer cancel()

	// Get user collection
	collection := config.GetCollection(cc.DB, "users")

	// Get user information from token
	claims := middleware.GetUserFromToken(c)
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid user ID",
		})
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Failed to parse form data: " + err.Error(),
		})
	}

	// Get branch data from form
	data := form.Value["data"]
	if len(data) == 0 {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Branch data is required",
		})
	}

	// Log the raw data for debugging
	log.Printf("Raw branch data received: %s", data[0])

	// Parse branch data
	var branchData map[string]interface{}
	if err := json.Unmarshal([]byte(data[0]), &branchData); err != nil {
		log.Printf("Error unmarshaling branch data: %v", err)
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid branch data format: " + err.Error(),
		})
	}

	// Handle file uploads
	files := form.File["images"]
	var imagePaths []string

	for _, file := range files {
		// Generate unique filename
		filename := uuid.New().String() + filepath.Ext(file.Filename)
		uploadPath := filepath.Join("uploads", filename)

		// Ensure uploads directory exists
		os.MkdirAll("uploads", 0755)

		// Save file to uploads directory
		src, err := file.Open()
		if err != nil {
			log.Printf("Error opening file %s: %v", file.Filename, err)
			continue // Skip this file if there's an error
		}
		defer src.Close()

		// Create destination file
		dst, err := os.Create(uploadPath)
		if err != nil {
			log.Printf("Error creating destination file %s: %v", uploadPath, err)
			continue // Skip this file if there's an error
		}
		defer dst.Close()

		// Copy file
		if _, err = io.Copy(dst, src); err != nil {
			log.Printf("Error copying file data: %v", err)
			continue // Skip this file if there's an error
		}

		imagePaths = append(imagePaths, uploadPath)
	}

	// Create branch object with safer type assertions
	branch := models.Branch{
		ID:          primitive.NewObjectID(),
		Name:        getString(branchData, "name", ""),
		Location:    getString(branchData, "location", ""),
		Phone:       getString(branchData, "phone", ""),
		Category:    getString(branchData, "category", ""),
		SubCategory: getString(branchData, "subCategory", ""),
		Description: getString(branchData, "description", ""),
		Images:      imagePaths,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Handle latitude and longitude with more type safety
	if lat, ok := branchData["latitude"]; ok && lat != nil {
		switch v := lat.(type) {
		case float64:
			branch.Latitude = v
		case int:
			branch.Latitude = float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				branch.Latitude = f
			}
		}
	}

	if lng, ok := branchData["longitude"]; ok && lng != nil {
		switch v := lng.(type) {
		case float64:
			branch.Longitude = v
		case int:
			branch.Longitude = float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				branch.Longitude = f
			}
		}
	}

	log.Printf("Prepared branch object: %+v", branch)

	// Update user document to add the branch
	update := bson.M{
		"$push": bson.M{
			"companyInfo.branches": branch,
		},
	}

	result, err := collection.UpdateByID(ctx, userID, update)
	if err != nil {
		log.Printf("Error updating database: %v", err)
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to save branch: " + err.Error(),
		})
	}

	log.Printf("Database update result: %+v", result)

	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Branch created successfully",
		Data: map[string]interface{}{
			"_id":         branch.ID.Hex(),
			"name":        branch.Name,
			"location":    branch.Location,
			"latitude":    branch.Latitude,
			"longitude":   branch.Longitude,
			"images":      branch.Images,
			"Phone":       branch.Phone,
			"Category":    branch.Category,
			"SubCategory": branch.SubCategory,
			"Description": branch.Description,
		},
	})
}

// Helper function to safely get string values from a map
func getString(data map[string]interface{}, key, defaultValue string) string {
	if val, ok := data[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetBranches retrieves all branches for a company
// GetBranches retrieves branches for a company (can be accessed by any authenticated user)
// GetBranches retrieves branches for a company
func (cc *CompanyController) GetBranches(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get company ID from query parameters
	companyIDParam := c.QueryParam("companyId")
	var companyID primitive.ObjectID
	var err error

	if companyIDParam != "" && companyIDParam != "null" {
		// If companyId is provided in query and not "null", use that
		log.Printf("Using companyId from query params: %s", companyIDParam)
		companyID, err = primitive.ObjectIDFromHex(companyIDParam)
		if err != nil {
			log.Printf("Invalid company ID format: %s", companyIDParam)
			return c.JSON(http.StatusBadRequest, models.Response{
				Status:  http.StatusBadRequest,
				Message: "Invalid company ID format",
			})
		}
	} else {
		// Otherwise use the authenticated user's ID
		claims := middleware.GetUserFromToken(c)
		log.Printf("No valid companyId provided, using authenticated user ID: %s", claims.UserID)
		companyID, err = primitive.ObjectIDFromHex(claims.UserID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.Response{
				Status:  http.StatusBadRequest,
				Message: "Invalid user ID",
			})
		}
	}

	// Find the company/user by ID
	collection := config.GetCollection(cc.DB, "users")
	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": companyID}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, models.Response{
				Status:  http.StatusNotFound,
				Message: "Company not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to find company",
		})
	}

	// Check if user is a company and has branches
	if user.CompanyInfo == nil || len(user.CompanyInfo.Branches) == 0 {
		return c.JSON(http.StatusOK, models.Response{
			Status:  http.StatusOK,
			Message: "No branches found for this company",
			Data:    []models.Branch{}, // Return empty array
		})
	}

	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Branches retrieved successfully",
		Data:    user.CompanyInfo.Branches,
	})
}

// DeleteBranch handles the deletion of a branch by ID
func (cc *CompanyController) DeleteBranch(c echo.Context) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user collection
	collection := config.GetCollection(cc.DB, "users")

	// Get user information from token
	claims := middleware.GetUserFromToken(c)
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid user ID",
		})
	}

	// Get branch ID from URL parameter
	branchID := c.Param("id")
	if branchID == "" {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Branch ID is required",
		})
	}

	// Convert string branch ID to ObjectID
	branchObjectID, err := primitive.ObjectIDFromHex(branchID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid branch ID format",
		})
	}

	// First, find the branch to get its image paths before deleting
	var user models.User
	err = collection.FindOne(ctx, bson.M{
		"_id":                      userID,
		"companyInfo.branches._id": branchObjectID,
	}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, models.Response{
				Status:  http.StatusNotFound,
				Message: "Branch not found",
			})
		}
		log.Printf("Error finding branch: %v", err)
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to find branch: " + err.Error(),
		})
	}

	// Find the branch with the matching ID and get its images
	var imagesToDelete []string
	if user.CompanyInfo != nil {
		for _, branch := range user.CompanyInfo.Branches {
			if branch.ID == branchObjectID {
				imagesToDelete = branch.Images
				break
			}
		}
	}

	// Update user document to pull the branch
	update := bson.M{
		"$pull": bson.M{
			"companyInfo.branches": bson.M{"_id": branchObjectID},
		},
	}

	result, err := collection.UpdateByID(ctx, userID, update)
	if err != nil {
		log.Printf("Error deleting branch: %v", err)
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to delete branch: " + err.Error(),
		})
	}

	if result.ModifiedCount == 0 {
		return c.JSON(http.StatusNotFound, models.Response{
			Status:  http.StatusNotFound,
			Message: "Branch not found or already deleted",
		})
	}

	// Delete image files from filesystem
	var deletionErrors []string
	for _, imagePath := range imagesToDelete {
		if err := os.Remove(imagePath); err != nil {
			log.Printf("Error deleting image file %s: %v", imagePath, err)
			deletionErrors = append(deletionErrors, imagePath)
		} else {
			log.Printf("Successfully deleted image file: %s", imagePath)
		}
	}

	// Log deletion errors but don't fail the request
	if len(deletionErrors) > 0 {
		log.Printf("Some image files could not be deleted: %v", deletionErrors)
	}

	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Branch deleted successfully",
	})
}

// UpdateBranch handles updating an existing branch with new data and images
func (cc *CompanyController) UpdateBranch(c echo.Context) error {
	// Add request logging at the start
	log.Printf("Starting branch update request from IP: %s", c.RealIP())

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get user collection
	collection := config.GetCollection(cc.DB, "users")

	// Get user information from token
	claims := middleware.GetUserFromToken(c)
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid user ID",
		})
	}

	// Get branch ID from URL parameter
	branchID := c.Param("id")
	if branchID == "" {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Branch ID is required",
		})
	}

	// Convert string branch ID to ObjectID
	branchObjectID, err := primitive.ObjectIDFromHex(branchID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid branch ID format",
		})
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Failed to parse form data: " + err.Error(),
		})
	}

	// Get branch data from form
	data := form.Value["data"]
	if len(data) == 0 {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Branch data is required",
		})
	}

	// Log the raw data for debugging
	log.Printf("Raw branch update data received: %s", data[0])

	// Parse branch data
	var branchData map[string]interface{}
	if err := json.Unmarshal([]byte(data[0]), &branchData); err != nil {
		log.Printf("Error unmarshaling branch data: %v", err)
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid branch data format: " + err.Error(),
		})
	}

	// First, find the existing branch to get current image paths
	var user models.User
	err = collection.FindOne(ctx, bson.M{
		"_id":                      userID,
		"companyInfo.branches._id": branchObjectID,
	}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, models.Response{
				Status:  http.StatusNotFound,
				Message: "Branch not found",
			})
		}
		log.Printf("Error finding branch: %v", err)
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to find branch: " + err.Error(),
		})
	}

	// Find the existing branch and get its current images
	var existingBranch models.Branch
	var existingImagePaths []string
	if user.CompanyInfo != nil {
		for _, branch := range user.CompanyInfo.Branches {
			if branch.ID == branchObjectID {
				existingBranch = branch
				existingImagePaths = branch.Images
				break
			}
		}
	}

	// Handle new file uploads
	files := form.File["images"]
	var newImagePaths []string

	for _, file := range files {
		// Generate unique filename
		filename := uuid.New().String() + filepath.Ext(file.Filename)
		uploadPath := filepath.Join("uploads", filename)

		// Ensure uploads directory exists
		os.MkdirAll("uploads", 0755)

		// Save file to uploads directory
		src, err := file.Open()
		if err != nil {
			log.Printf("Error opening file %s: %v", file.Filename, err)
			continue // Skip this file if there's an error
		}
		defer src.Close()

		// Create destination file
		dst, err := os.Create(uploadPath)
		if err != nil {
			log.Printf("Error creating destination file %s: %v", uploadPath, err)
			continue // Skip this file if there's an error
		}
		defer dst.Close()

		// Copy file
		if _, err = io.Copy(dst, src); err != nil {
			log.Printf("Error copying file data: %v", err)
			continue // Skip this file if there's an error
		}

		newImagePaths = append(newImagePaths, uploadPath)
	}

	// Determine final image paths (keep existing unless new ones are provided)
	var finalImagePaths []string
	if len(newImagePaths) > 0 {
		finalImagePaths = newImagePaths
		// Delete old images if new ones are uploaded
		for _, oldPath := range existingImagePaths {
			if err := os.Remove(oldPath); err != nil {
				log.Printf("Warning: Failed to delete old image %s: %v", oldPath, err)
			}
		}
	} else {
		finalImagePaths = existingImagePaths
	}

	// Create updated branch object with safer type assertions
	updatedBranch := models.Branch{
		ID:          branchObjectID, // Keep the original ID
		Name:        getString(branchData, "name", existingBranch.Name),
		Location:    getString(branchData, "location", existingBranch.Location),
		Phone:       getString(branchData, "phone", existingBranch.Phone),
		Category:    getString(branchData, "category", existingBranch.Category),
		SubCategory: getString(branchData, "subCategory", existingBranch.SubCategory),
		Description: getString(branchData, "description", existingBranch.Description),
		Images:      finalImagePaths,
		CreatedAt:   existingBranch.CreatedAt, // Keep original creation time
		UpdatedAt:   time.Now(),               // Update the update timestamp
	}

	// Handle latitude and longitude with more type safety
	if lat, ok := branchData["latitude"]; ok && lat != nil {
		switch v := lat.(type) {
		case float64:
			updatedBranch.Latitude = v
		case int:
			updatedBranch.Latitude = float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				updatedBranch.Latitude = f
			}
		}
	} else {
		updatedBranch.Latitude = existingBranch.Latitude
	}

	if lng, ok := branchData["longitude"]; ok && lng != nil {
		switch v := lng.(type) {
		case float64:
			updatedBranch.Longitude = v
		case int:
			updatedBranch.Longitude = float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				updatedBranch.Longitude = f
			}
		}
	} else {
		updatedBranch.Longitude = existingBranch.Longitude
	}

	log.Printf("Prepared updated branch object: %+v", updatedBranch)

	// Update the branch in the database
	// First, remove the old branch
	pull := bson.M{
		"$pull": bson.M{
			"companyInfo.branches": bson.M{"_id": branchObjectID},
		},
	}

	_, err = collection.UpdateByID(ctx, userID, pull)
	if err != nil {
		log.Printf("Error removing old branch: %v", err)
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update branch: " + err.Error(),
		})
	}

	// Then, add the updated branch
	push := bson.M{
		"$push": bson.M{
			"companyInfo.branches": updatedBranch,
		},
	}

	result, err := collection.UpdateByID(ctx, userID, push)
	if err != nil {
		log.Printf("Error updating branch: %v", err)
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update branch: " + err.Error(),
		})
	}

	log.Printf("Database update result: %+v", result)

	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Branch updated successfully",
		Data: map[string]interface{}{
			"_id":         updatedBranch.ID.Hex(),
			"name":        updatedBranch.Name,
			"location":    updatedBranch.Location,
			"latitude":    updatedBranch.Latitude,
			"longitude":   updatedBranch.Longitude,
			"images":      updatedBranch.Images,
			"Phone":       updatedBranch.Phone,
			"Category":    updatedBranch.Category,
			"SubCategory": updatedBranch.SubCategory,
			"Description": updatedBranch.Description,
		},
	})
}

// GetCompanyBranches retrieves branches for a specific company by ID
func (cc *CompanyController) GetCompanyBranches(c echo.Context) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get company ID from URL parameter
	companyID := c.Param("id")
	if companyID == "" {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Company ID is required",
		})
	}

	// Convert string company ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(companyID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid company ID format",
		})
	}

	// Find user by ID without restricting to company type
	collection := config.GetCollection(cc.DB, "users")
	var user models.User
	err = collection.FindOne(ctx, bson.M{
		"_id": objectID,
	}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, models.Response{
				Status:  http.StatusNotFound,
				Message: "Company not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to find company",
		})
	}

	// Ensure CompanyInfo exists
	if user.CompanyInfo == nil {
		// Return empty branches if not a company user
		return c.JSON(http.StatusOK, models.Response{
			Status:  http.StatusOK,
			Message: "No branch data available",
			Data:    []models.Branch{}, // Return empty array
		})
	}

	// Return branches
	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Branches retrieved successfully",
		Data:    user.CompanyInfo.Branches,
	})
}

func (cc *CompanyController) UpdateCompanyData(c echo.Context) error {
	// Get user ID from JWT token
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.Response{
			Status:  http.StatusUnauthorized,
			Message: "Invalid or expired token: " + err.Error(),
		})
	}

	// Parse the request body
	var updateData map[string]interface{}
	if err := c.Bind(&updateData); err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid request format: " + err.Error(),
		})
	}

	// Convert userID string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Response{
			Status:  http.StatusBadRequest,
			Message: "Invalid user ID format: " + err.Error(),
		})
	}

	// Create update document with the fields to update
	update := bson.M{
		"$set": bson.M{
			"phone":     updateData["phone"],
			"whatsapp":  updateData["whatsapp"],
			"website":   updateData["website"],
			"facebook":  updateData["facebook"],
			"instagram": updateData["instagram"],
			"updatedAt": time.Now(),
		},
	}

	// Update the document in MongoDB
	filter := bson.M{"_id": objectID}

	// Use the correct database and collection names
	collection := cc.DB.Database("barrim").Collection("users")

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update company data: " + err.Error(),
		})
	}

	if result.ModifiedCount == 0 {
		// Check if the user exists
		var user models.User
		err = collection.FindOne(context.Background(), filter).Decode(&user)
		if err != nil {
			return c.JSON(http.StatusNotFound, models.Response{
				Status:  http.StatusNotFound,
				Message: "User not found",
			})
		}

		// User exists but no changes were made
		return c.JSON(http.StatusOK, models.Response{
			Status:  http.StatusOK,
			Message: "No changes made to company data",
		})
	}

	return c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Message: "Company data updated successfully",
	})
}
