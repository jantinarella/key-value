package server

import (
	"context"
	"errors"
	"testing"

	"key-value/proto/keyvalue"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockStorer implements kvstore.Storer for testing
type MockStorer struct {
	GetFunc    func(key string) (string, error)
	SetFunc    func(key, value string) error
	DeleteFunc func(key string) error
}

func (m *MockStorer) Get(key string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(key)
	}
	return "mock-value", nil
}

func (m *MockStorer) Set(key, value string) error {
	if m.SetFunc != nil {
		return m.SetFunc(key, value)
	}
	return nil
}

func (m *MockStorer) Delete(key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(key)
	}
	return nil
}

func TestKeyValueServer_Get(t *testing.T) {
	tests := []struct {
		name           string
		request        *keyvalue.GetRequest
		setupMock      func(*MockStorer)
		expectedValue  string
		expectedFound  bool
		expectedError  string
		expectGRPCCode codes.Code
	}{
		{
			name:    "successful get",
			request: &keyvalue.GetRequest{Key: "test-key"},
			setupMock: func(m *MockStorer) {
				m.GetFunc = func(key string) (string, error) {
					return "test-value", nil
				}
			},
			expectedValue: "test-value",
			expectedFound: true,
		},
		{
			name:    "key not found",
			request: &keyvalue.GetRequest{Key: "missing-key"},
			setupMock: func(m *MockStorer) {
				m.GetFunc = func(key string) (string, error) {
					return "", errors.New("key not found")
				}
			},
			expectedValue: "",
			expectedFound: false,
		},
		{
			name:           "empty key",
			request:        &keyvalue.GetRequest{Key: ""},
			setupMock:      func(m *MockStorer) {},
			expectGRPCCode: codes.InvalidArgument,
		},
		{
			name:    "store error",
			request: &keyvalue.GetRequest{Key: "error-key"},
			setupMock: func(m *MockStorer) {
				m.GetFunc = func(key string) (string, error) {
					return "", errors.New("connection failed")
				}
			},
			expectGRPCCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &MockStorer{}
			tt.setupMock(mockStore)

			server := NewKeyValueServer(mockStore)
			ctx := context.Background()

			resp, err := server.Get(ctx, tt.request)

			if tt.expectGRPCCode != codes.OK {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectGRPCCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedValue, resp.Value)
				assert.Equal(t, tt.expectedFound, resp.Found)
			}
		})
	}
}

func TestKeyValueServer_Set(t *testing.T) {
	tests := []struct {
		name            string
		request         *keyvalue.SetRequest
		setupMock       func(*MockStorer)
		expectedSuccess bool
		expectedError   string
		expectGRPCCode  codes.Code
	}{
		{
			name:    "successful set",
			request: &keyvalue.SetRequest{Key: "test-key", Value: "test-value"},
			setupMock: func(m *MockStorer) {
				m.SetFunc = func(key, value string) error {
					return nil
				}
			},
			expectedSuccess: true,
		},
		{
			name:           "empty key",
			request:        &keyvalue.SetRequest{Key: "", Value: "test-value"},
			setupMock:      func(m *MockStorer) {},
			expectGRPCCode: codes.InvalidArgument,
		},
		{
			name:    "store error",
			request: &keyvalue.SetRequest{Key: "test-key", Value: "test-value"},
			setupMock: func(m *MockStorer) {
				m.SetFunc = func(key, value string) error {
					return errors.New("storage failed")
				}
			},
			expectedSuccess: false,
			expectedError:   "storage failed",
		},
		{
			name:    "empty value allowed",
			request: &keyvalue.SetRequest{Key: "test-key", Value: ""},
			setupMock: func(m *MockStorer) {
				m.SetFunc = func(key, value string) error {
					return nil
				}
			},
			expectedSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &MockStorer{}
			tt.setupMock(mockStore)

			server := NewKeyValueServer(mockStore)
			ctx := context.Background()

			resp, err := server.Set(ctx, tt.request)

			if tt.expectGRPCCode != codes.OK {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectGRPCCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedSuccess, resp.Success)
				if tt.expectedError != "" {
					assert.Equal(t, tt.expectedError, resp.Error)
				}
			}
		})
	}
}

func TestKeyValueServer_Delete(t *testing.T) {
	tests := []struct {
		name            string
		request         *keyvalue.DeleteRequest
		setupMock       func(*MockStorer)
		expectedSuccess bool
		expectedError   string
		expectGRPCCode  codes.Code
	}{
		{
			name:    "successful delete",
			request: &keyvalue.DeleteRequest{Key: "test-key"},
			setupMock: func(m *MockStorer) {
				m.DeleteFunc = func(key string) error {
					return nil
				}
			},
			expectedSuccess: true,
		},
		{
			name:           "empty key",
			request:        &keyvalue.DeleteRequest{Key: ""},
			setupMock:      func(m *MockStorer) {},
			expectGRPCCode: codes.InvalidArgument,
		},
		{
			name:    "store error",
			request: &keyvalue.DeleteRequest{Key: "test-key"},
			setupMock: func(m *MockStorer) {
				m.DeleteFunc = func(key string) error {
					return errors.New("delete failed")
				}
			},
			expectedSuccess: false,
			expectedError:   "delete failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &MockStorer{}
			tt.setupMock(mockStore)

			server := NewKeyValueServer(mockStore)
			ctx := context.Background()

			resp, err := server.Delete(ctx, tt.request)

			if tt.expectGRPCCode != codes.OK {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectGRPCCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedSuccess, resp.Success)
				if tt.expectedError != "" {
					assert.Equal(t, tt.expectedError, resp.Error)
				}
			}
		})
	}
}
