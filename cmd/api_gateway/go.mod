module github.com/anonymousminh/url_shortener/cmd/api_gateway

go 1.25.0

require github.com/anonymousminh/url_shortener/internal/logger v0.0.0

require (
	github.com/labstack/echo/v5 v5.0.4 // indirect
	golang.org/x/time v0.14.0 // indirect
)

replace github.com/anonymousminh/url_shortener/internal/logger => ../../internal/logger
