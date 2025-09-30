package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"key-value/shared/models"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// MockKVStoreClient implements a mock for testing
type MockKVStoreClient struct {
	GetFunc    func(ctx context.Context, key string) (string, bool, error)
	SetFunc    func(ctx context.Context, kv models.KeyValue) error
	DeleteFunc func(ctx context.Context, key string) error
	HealthFunc func(ctx context.Context) error
	CloseFunc  func() error
}

func (m *MockKVStoreClient) Get(ctx context.Context, key string) (string, bool, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	return "mock-value", true, nil
}

func (m *MockKVStoreClient) Set(ctx context.Context, kv models.KeyValue) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, kv)
	}
	return nil
}

func (m *MockKVStoreClient) Delete(ctx context.Context, key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, key)
	}
	return nil
}

func (m *MockKVStoreClient) Health(ctx context.Context) error {
	if m.HealthFunc != nil {
		return m.HealthFunc(ctx)
	}
	return nil
}

func (m *MockKVStoreClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func TestHandler_GetValueByKey(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		setupMock      func(*MockKVStoreClient)
		expectedStatus int
		expectedKey    string
		expectedValue  string
		expectError    bool
	}{
		{
			name: "successful get",
			key:  "test-key",
			setupMock: func(m *MockKVStoreClient) {
				m.GetFunc = func(ctx context.Context, key string) (string, bool, error) {
					return "test-value", true, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedKey:    "test-key",
			expectedValue:  "test-value",
		},
		{
			name: "key not found",
			key:  "missing-key",
			setupMock: func(m *MockKVStoreClient) {
				m.GetFunc = func(ctx context.Context, key string) (string, bool, error) {
					return "", false, nil
				}
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name: "client error",
			key:  "error-key",
			setupMock: func(m *MockKVStoreClient) {
				m.GetFunc = func(ctx context.Context, key string) (string, bool, error) {
					return "", false, errors.New("connection failed")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := &MockKVStoreClient{}
			tt.setupMock(mockClient)

			// Create handler with mock client
			handler := NewHandler(mockClient)

			// Set up Echo context
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("key")
			c.SetParamValues(tt.key)

			// Call handler
			err := handler.GetValueByKey(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Check response
			if tt.expectError {
				var response map[string]string
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			} else {
				var response map[string]string
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedKey, response["key"])
				assert.Equal(t, tt.expectedValue, response["value"])
			}
		})
	}
}

func TestHandler_UpdateValue(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockKVStoreClient)
		expectedStatus int
		expectedError  string
		expectedKey    string
		expectedValue  string
	}{
		{
			name: "successful update",
			requestBody: map[string]string{
				"key":   "test-key",
				"value": "test-value",
			},
			setupMock: func(m *MockKVStoreClient) {
				m.SetFunc = func(ctx context.Context, kv models.KeyValue) error {
					return nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedKey:    "test-key",
			expectedValue:  "test-value",
		},
		{
			name: "missing key",
			requestBody: map[string]string{
				"value": "test-value",
			},
			setupMock: func(m *MockKVStoreClient) {
				m.SetFunc = func(ctx context.Context, kv models.KeyValue) error {
					return nil
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Key is required",
		},
		{
			name: "empty key",
			requestBody: map[string]string{
				"key":   "",
				"value": "test-value",
			},
			setupMock: func(m *MockKVStoreClient) {
				m.SetFunc = func(ctx context.Context, kv models.KeyValue) error {
					return nil
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Key is required",
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			setupMock:      func(m *MockKVStoreClient) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "client error",
			requestBody: map[string]string{
				"key":   "test-key",
				"value": "test-value",
			},
			setupMock: func(m *MockKVStoreClient) {
				m.SetFunc = func(ctx context.Context, kv models.KeyValue) error {
					return errors.New("connection failed")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to update value connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := &MockKVStoreClient{}
			tt.setupMock(mockClient)

			// Create handler with mock client
			handler := NewHandler(mockClient)

			// Set up request
			e := echo.New()
			var reqBody []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Call handler
			err = handler.UpdateValue(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Check response
			if tt.expectedError != "" {
				var response map[string]string
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			} else {
				var response map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedKey, response["key"])
				assert.Equal(t, tt.expectedValue, response["value"])
			}
		})
	}
}

func TestHandler_DeleteValue(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		setupMock      func(*MockKVStoreClient)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful delete",
			key:  "test-key",
			setupMock: func(m *MockKVStoreClient) {
				m.DeleteFunc = func(ctx context.Context, key string) error {
					return nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "empty key validation",
			key:            "",
			setupMock:      func(m *MockKVStoreClient) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Key is required",
		},
		{
			name: "client error",
			key:  "error-key",
			setupMock: func(m *MockKVStoreClient) {
				m.DeleteFunc = func(ctx context.Context, key string) error {
					return errors.New("connection failed")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to delete value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := &MockKVStoreClient{}
			tt.setupMock(mockClient)

			// Create handler with mock client
			handler := NewHandler(mockClient)

			// Set up Echo context
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("key")
			c.SetParamValues(tt.key)

			// Call handler
			err := handler.DeleteValue(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Check error response
			if tt.expectedError != "" {
				var response map[string]string
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}
		})
	}
}
