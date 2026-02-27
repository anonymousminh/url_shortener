package storage

import "context"

type Storer interface {

	// saveURL saves the original URL to the storage and returns the ID of the saved URL
	SaveURL(ctx context.Context, originalURL string) (int64, error)

	// UpdateShortCode updates the short code of the URL in the storage
	UpdateShortCode(ctx context.Context, id int64, shortCode string) error

	// GetURLByShortCode gets the original URL from the storage by the short code
	GetURLByShortCode(ctx context.Context, shortCode string) (string, error)
}
