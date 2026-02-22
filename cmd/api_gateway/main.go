package main

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	e := echo.New()

	e.HTTPErrorHandler = customErrorHandler

	rateLimitConfig := middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStore(20.0),
		DenyHandler: func(c *echo.Context, identifier string, err error) error {
			return &echo.HTTPError{
				Code:    middleware.ErrRateLimitExceeded.Code,
				Message: middleware.ErrRateLimitExceeded.Message,
			}
		},
	}

	e.Use(middleware.RateLimiterWithConfig(rateLimitConfig))

	urlServiceURL, err := url.Parse("http://localhost:8081")
	if err != nil {
		e.Logger.Error("Could not parse url-service url")
	}

	e.POST("/api/v1/urls", echo.WrapHandler(httputil.NewSingleHostReverseProxy(urlServiceURL)))
	e.GET("/:short_code", redirectHandler(urlServiceURL.String()))

	e.GET("/health", func(c *echo.Context) error {
		return c.String(http.StatusOK, "API Gateway is healthy")
	})

	if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
		e.Logger.Error("Failed to start the server", "error", err)
	}
}

func customErrorHandler(c *echo.Context, err error) {
	if r, _ := echo.UnwrapResponse(c.Response()); r != nil && r.Committed {
		return
	}

	code := http.StatusInternalServerError
	message := "Internal Server Error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message
	}

	c.Echo().Logger.Error("request error", "error", err)
	c.JSON(code, map[string]string{"message": message})
}

func redirectHandler(urlServiceBaseURL string) echo.HandlerFunc {
	return func(c *echo.Context) error {
		shortCode := c.Param("short_code")

		req, _ := http.NewRequest("GET", urlServiceBaseURL+"/api/v1/urls/"+shortCode, nil)
		client := new(http.Client)
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			return c.String(http.StatusNotFound, "URL not found")
		}
		defer resp.Body.Close()

		var data map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return c.String(http.StatusInternalServerError, "Could not decode response")
		}

		return c.Redirect(http.StatusMovedPermanently, data["original_url"])
	}
}
