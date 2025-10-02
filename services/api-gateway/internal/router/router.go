package router

import (
	"key-value/client"
	"key-value/services/api-gateway/internal/config"
	"key-value/services/api-gateway/internal/handlers"
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, config *config.Config) error {

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "I am alive in "+config.Environment)
	})

	// Protected routes with API key middleware
	v1 := e.Group("/v1",
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				requestAPIKey := c.Request().Header.Get("x-api-key")
				if requestAPIKey != config.APIKey {
					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid API key"})
				}
				return next(c)
			}
		})

	// Create a new KVStoreClient
	kvstoreClient, err := client.NewKVStoreClient(config.KVServiceAddr)
	if err != nil {
		e.Logger.Errorf("Failed to create KVStoreClient: %v", err)
		return err
	}

	// Initialize handlers
	handler := handlers.NewHandler(kvstoreClient)

	// Value endpoints
	v1.GET("/values/:key", handler.GetValueByKey)
	v1.PUT("/values", handler.UpdateValue)
	v1.DELETE("/values/:key", handler.DeleteValue)

	return nil
}
