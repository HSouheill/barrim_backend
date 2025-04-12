package routes

import (
	"github.com/HSouheill/barrim_backend/controllers"
	"github.com/HSouheill/barrim_backend/middleware"
	"github.com/labstack/echo/v4"
)

func RegisterCompanyRoutes(e *echo.Echo, companyController *controllers.CompanyController) {
	// Protected routes - require authentication
	companyGroup := e.Group("/api/company")
	companyGroup.Use(middleware.JWTMiddleware())

	companyGroup.GET("/data", companyController.GetCompanyData)
	companyGroup.PUT("/data", companyController.UpdateCompanyData)
	companyGroup.POST("/branches", companyController.CreateBranch)
	companyGroup.GET("/branches", companyController.GetBranches)
	companyGroup.DELETE("/branches/:id", companyController.DeleteBranch)
	companyGroup.PUT("/branches/:id", companyController.UpdateBranch)

}
