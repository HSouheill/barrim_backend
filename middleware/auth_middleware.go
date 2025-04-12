// middleware/auth_middleware.go
package middleware

import (
	"net/http"

	"github.com/HSouheill/barrim_backend/models"
	"github.com/labstack/echo/v4"
)

// RequireUserType checks if the authenticated user has one of the allowed user types
func RequireUserType(allowedTypes ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := GetUserFromToken(c)

			for _, allowedType := range allowedTypes {
				if claims.UserType == allowedType {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, models.Response{
				Status:  http.StatusForbidden,
				Message: "Access denied for your user type",
			})
		}
	}
}
