package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/anonymousminh/url_shortener/internal/storage"
	"github.com/anonymousminh/url_shortener/pkg/base62"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

type application struct {
	store storage.Storer
}

type CreateURLRequest struct {
	URL string `json:"url"`
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	dbpool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to create database connection pool: %v", err)
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	defer dbpool.Close()

	app := &application{
		store: storage.NewPostgresStorage(dbpool),
	}

	e := echo.New()

	e.POST("/api/v1/urls", app.createShortURLHandler)
	e.GET("/api/v1/urls/:short_code", app.getOriginalURLHandler)

	if err := e.Start(":8081"); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Shutting down the URL Service: %v", err)
	}
}

func (app *application) createShortURLHandler(c *echo.Context) error {
	req := new(CreateURLRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid URL format")
	}

	id, err := app.store.SaveURL(c.Request().Context(), req.URL)
	if err != nil {
		log.Printf("Failed to save URL: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save URL")
	}

	shortCode := base62.Encode(uint64(id))

	if err := app.store.UpdateShortCode(c.Request().Context(), id, shortCode); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not update short code")
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"short_url":    fmt.Sprintf("http://localhost:8081/%s", shortCode),
		"original_url": req.URL,
	})
}

func (app *application) getOriginalURLHandler(c *echo.Context) error {
	shortCode := c.Param("short_code")

	originalURL, err := app.store.GetURLByShortCode(c.Request().Context(), shortCode)
	if err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Short URL not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"original_url": originalURL,
	})
}
