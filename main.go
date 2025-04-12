package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/HSouheill/barrim_backend/config"
	"github.com/HSouheill/barrim_backend/controllers"
	"github.com/HSouheill/barrim_backend/routes"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	// Connect to database
	client := config.ConnectDB()

	// Create a new Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.Secure())

	companyController := controllers.NewCompanyController(client)
	// Register company routes
	routes.RegisterCompanyRoutes(e, companyController)
	// Setup routes
	routes.SetupRoutes(e, client)
	// Ensure uploads directory exists
	os.MkdirAll("uploads", 0755)
	// Add this to your Echo server setup
	e.Static("/uploads", "uploads")
	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
