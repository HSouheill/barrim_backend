package routes

import (
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/HSouheill/barrim_backend/controllers"
	customMiddleware "github.com/HSouheill/barrim_backend/middleware"
)

// SetupRoutes configures all API routes
func SetupRoutes(e *echo.Echo, db *mongo.Client) {
	// Create controllers
	authController := controllers.NewAuthController(db)
	userController := controllers.NewUserController(db)
	passwordController := controllers.NewPasswordController(db)

	// Public routes
	e.POST("/api/auth/signup", authController.Signup)
	e.POST("/api/auth/login", authController.Login)
	e.POST("api/auth/google", authController.GoogleLogin)

	// Public routes
	e.GET("/api/service-providers", userController.SearchServiceProviders)

	e.POST("/api/auth/forget-password", passwordController.ForgetPassword)
	e.POST("/api/auth/verify-otp", passwordController.VerifyOTP)
	e.POST("/api/auth/reset-password", passwordController.ResetPassword)
	e.GET("/uploads/:filename", controllers.ServeImage)

	// Protected routes
	r := e.Group("/api")
	r.Use(middleware.JWTWithConfig(customMiddleware.GetJWTConfig()))

	// User routes
	r.GET("/users", userController.GetAllUsers)
	r.GET("/users/profile", userController.GetProfile)
	r.PUT("/users/profile", userController.UpdateProfile)
	r.PUT("/users/location", userController.UpdateLocation) // Existing route for updating location
	r.DELETE("/users", userController.DeleteUser)
	r.POST("/upload-logo", userController.UploadCompanyLogo)
	r.POST("/upload-profile-photo", userController.UploadProfilePhoto)
	r.POST("/update-availability", userController.UpdateAvailability)
	// In your routes.go file, add this line to the protected routes section:
	r.GET("/user/companies", userController.GetCompaniesWithLocations)
	// Add the new save-locations route
	r.POST("/save-locations", userController.UpdateLocation) // Reuse the UpdateLocation method

	// Company-specific routes
	company := r.Group("/api/company")
	company.Use(customMiddleware.RequireUserType("company", "wholesaler"))
	company.POST("/logo", userController.UploadCompanyLogo)

	// Service provider specific routes
	serviceProvider := r.Group("/service-provider")
	serviceProvider.Use(customMiddleware.RequireUserType("serviceProvider"))
	serviceProvider.POST("/availability", userController.UpdateAvailability)
	serviceProvider.POST("/photo", userController.UploadProfilePhoto)
}

// Helper function to check place types
func containsType(types []string, targetType string) bool {
	for _, t := range types {
		if t == targetType {
			return true
		}
	}
	return false
}

// ServeImage handles serving uploaded images
func ServeImage(c echo.Context) error {
	filename := c.Param("filename")
	return c.File(filepath.Join("uploads", filename))
}
