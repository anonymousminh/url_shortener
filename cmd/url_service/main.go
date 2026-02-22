package main

import (
	"net/http"
	"net/url"
	"sync/atomic"

	"github.com/anonymousminh/url_shortener/pkg/base62"

	"github.com/labstack/echo/v5"
)

// Try in-memory first
var (
	urlStore = make(map[string]string)
	counter  uint64
)

// Request/Response Structs
type CreateURLRequest struct {
	URL string `json:"url"`
}

type CreateURLResponse struct {
	URL string `json:"short_url"`
}

func main() {
	e := echo.New()

	// API Endpoints
	e.POST("/api/v1/urls", createShortURLHandler)
	e.GET("/api/v1/urls/:short_code", GetOriginalURLHandler)

	// Start URL Service in port 8081
	if err := e.Start(":8081"); err != nil && err != http.ErrServerClosed {
		e.Logger.Error("Shutting down the URL Service", "error", err)
	}
}

// Handlers
func createShortURLHandler(c *echo.Context) error {
	// 1. Bind the request body
	req := new(CreateURLRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// 2. Validate the URL
	if _, err := url.ParseRequestURI(req.URL); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid URL format")
	}

	// 3. Generate a new ID
	newID := atomic.AddUint64(&counter, 1)

	// 4. Encode the ID to a short code
	shortCode := base62.Encode(newID)

	// 5. Store the mapping (in memory as for now)
	urlStore[shortCode] = req.URL

	// 6. Create the response
	resp := CreateURLResponse{
		URL: "http://localhost:8080/" + shortCode,
	}

	return c.JSON(http.StatusCreated, resp)
}

func GetOriginalURLHandler(c *echo.Context) error {
	shortCode := c.Param("short_code")

	originalURL, found := urlStore[shortCode]
	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "Short URL not found")
	}

	return c.JSON(http.StatusOK, map[string]string{"original_url": originalURL})
}
