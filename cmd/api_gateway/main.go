package main

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {

	// Create a new Echo instance
	e := echo.New()

	// Configure the Rate Limter
	rateLimitConfig := middleware.RateLimiterConfig{
		// Store our own defined rate limit
		Store: middleware.NewRateLimiterMemoryStore(20.0),
		// DenyHandler to handle the error message if denied
		DenyHandler: func(c *echo.Context, identifier string, err error) error {
			return &echo.HTTPError{
				Code:    middleware.ErrRateLimitExceeded.Code,
				Message: middleware.ErrRateLimitExceeded.Message,
			}
		},
	}

	// Register middleware
	e.Use(middleware.RateLimiterWithConfig(rateLimitConfig))

	// Define a simple route called health
	e.GET("/health", func(c *echo.Context) error {
		return c.String(http.StatusOK, "API Gateway is healthy")
	})

	// Start server on port 8080
	if err := e.Start(":8080"); err != nil {
		e.Logger.Error("Failed to start the server", "error", err)
	}
}
