package handlers

import (
	"key-value/shared/models"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// GetValueByKey retrieves a KeyValue by key
func (h *Handler) GetValueByKey(c echo.Context) error {
	key := c.Param("key")
	value, found, err := h.kvstoreClient.Get(c.Request().Context(), key)
	if err != nil {
		log.Printf("Failed to get value: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get value"})
	}
	if !found {
		log.Printf("Key not found: %s", key)
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "Key not found"})
	}

	return c.JSON(http.StatusOK, models.KeyValue{
		Key:   key,
		Value: value,
	})
}

// UpdateValue updates a KeyValue pair writing over the existing value if present
func (h *Handler) UpdateValue(c echo.Context) error {
	keyValue := models.KeyValue{}
	if err := c.Bind(&keyValue); err != nil {
		log.Printf("Failed to bind request body: %v", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	// Validate the reqest has a key
	if keyValue.Key == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Key is required"})
	}

	// Update the value
	err := h.kvstoreClient.Set(c.Request().Context(), keyValue)
	if err != nil {
		log.Printf("Failed to update value: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update value " + err.Error()})
	}

	return c.JSON(http.StatusOK, models.KeyValue{
		Key:   keyValue.Key,
		Value: keyValue.Value,
	})
}

// DeleteValue deletes a KeyValue pair if the value does not exist, it is a no-op
func (h *Handler) DeleteValue(c echo.Context) error {
	key := c.Param("key")
	if key == "" {
		log.Printf("Key is required")
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Key is required"})
	}

	// Delete the value
	err := h.kvstoreClient.Delete(c.Request().Context(), key)
	if err != nil {
		log.Printf("Failed to delete value: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete value"})
	}

	return c.NoContent(http.StatusNoContent)
}
