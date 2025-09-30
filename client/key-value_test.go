package client

import (
	"context"
	"testing"
	"time"

	"key-value/proto/keyvalue"
	"key-value/shared/models"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockKeyValueServiceClient implements the gRPC client interface for testing
type MockKeyValueServiceClient struct {
	GetFunc    func(ctx context.Context, in *keyvalue.GetRequest, opts ...grpc.CallOption) (*keyvalue.GetResponse, error)
	SetFunc    func(ctx context.Context, in *keyvalue.SetRequest, opts ...grpc.CallOption) (*keyvalue.SetResponse, error)
	DeleteFunc func(ctx context.Context, in *keyvalue.DeleteRequest, opts ...grpc.CallOption) (*keyvalue.DeleteResponse, error)
	HealthFunc func(ctx context.Context, in *keyvalue.HealthRequest, opts ...grpc.CallOption) (*keyvalue.HealthResponse, error)
}

func (m *MockKeyValueServiceClient) Get(ctx context.Context, in *keyvalue.GetRequest, opts ...grpc.CallOption) (*keyvalue.GetResponse, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, in, opts...)
	}
	return &keyvalue.GetResponse{Value: "mock-value", Found: true}, nil
}

func (m *MockKeyValueServiceClient) Set(ctx context.Context, in *keyvalue.SetRequest, opts ...grpc.CallOption) (*keyvalue.SetResponse, error) {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, in, opts...)
	}
	return &keyvalue.SetResponse{Success: true}, nil
}

func (m *MockKeyValueServiceClient) Delete(ctx context.Context, in *keyvalue.DeleteRequest, opts ...grpc.CallOption) (*keyvalue.DeleteResponse, error) {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, in, opts...)
	}
	return &keyvalue.DeleteResponse{Success: true}, nil
}

func (m *MockKeyValueServiceClient) Health(ctx context.Context, in *keyvalue.HealthRequest, opts ...grpc.CallOption) (*keyvalue.HealthResponse, error) {
	if m.HealthFunc != nil {
		return m.HealthFunc(ctx, in, opts...)
	}
	return &keyvalue.HealthResponse{Status: "healthy", Timestamp: time.Now().Unix()}, nil
}

func TestKVStoreClient_Get(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		setupMock      func(*MockKeyValueServiceClient)
		expectedValue  string
		expectedFound  bool
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful get",
			key:  "test-key",
			setupMock: func(m *MockKeyValueServiceClient) {
				m.GetFunc = func(ctx context.Context, in *keyvalue.GetRequest, opts ...grpc.CallOption) (*keyvalue.GetResponse, error) {
					return &keyvalue.GetResponse{
						Value: "test-value",
						Found: true,
					}, nil
				}
			},
			expectedValue: "test-value",
			expectedFound: true,
		},
		{
			name: "key not found",
			key:  "missing-key",
			setupMock: func(m *MockKeyValueServiceClient) {
				m.GetFunc = func(ctx context.Context, in *keyvalue.GetRequest, opts ...grpc.CallOption) (*keyvalue.GetResponse, error) {
					return &keyvalue.GetResponse{
						Value: "",
						Found: false,
					}, nil
				}
			},
			expectedValue: "",
			expectedFound: false,
		},
		{
			name: "grpc error",
			key:  "error-key",
			setupMock: func(m *MockKeyValueServiceClient) {
				m.GetFunc = func(ctx context.Context, in *keyvalue.GetRequest, opts ...grpc.CallOption) (*keyvalue.GetResponse, error) {
					return nil, status.Errorf(codes.Internal, "service unavailable")
				}
			},
			expectError:    true,
			expectedErrMsg: "failed to get key error-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockKeyValueServiceClient{}
			tt.setupMock(mockClient)

			// Create client with mock
			client := &KVStoreClient{
				client: mockClient,
				addr:   "mock-address",
			}

			ctx := context.Background()
			value, found, err := client.Get(ctx, tt.key)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, value)
				assert.Equal(t, tt.expectedFound, found)
			}
		})
	}
}

func TestKVStoreClient_Set(t *testing.T) {
	tests := []struct {
		name           string
		kv             models.KeyValue
		setupMock      func(*MockKeyValueServiceClient)
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful set",
			kv:   models.KeyValue{Key: "test-key", Value: "test-value"},
			setupMock: func(m *MockKeyValueServiceClient) {
				m.SetFunc = func(ctx context.Context, in *keyvalue.SetRequest, opts ...grpc.CallOption) (*keyvalue.SetResponse, error) {
					return &keyvalue.SetResponse{
						Success: true,
					}, nil
				}
			},
		},
		{
			name: "set operation failed",
			kv:   models.KeyValue{Key: "test-key", Value: "test-value"},
			setupMock: func(m *MockKeyValueServiceClient) {
				m.SetFunc = func(ctx context.Context, in *keyvalue.SetRequest, opts ...grpc.CallOption) (*keyvalue.SetResponse, error) {
					return &keyvalue.SetResponse{
						Success: false,
						Error:   "storage failed",
					}, nil
				}
			},
			expectError:    true,
			expectedErrMsg: "set operation failed: storage failed",
		},
		{
			name: "grpc error",
			kv:   models.KeyValue{Key: "test-key", Value: "test-value"},
			setupMock: func(m *MockKeyValueServiceClient) {
				m.SetFunc = func(ctx context.Context, in *keyvalue.SetRequest, opts ...grpc.CallOption) (*keyvalue.SetResponse, error) {
					return nil, status.Errorf(codes.Internal, "service unavailable")
				}
			},
			expectError:    true,
			expectedErrMsg: "failed to set key test-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockKeyValueServiceClient{}
			tt.setupMock(mockClient)

			// Create client with mock
			client := &KVStoreClient{
				client: mockClient,
				addr:   "mock-address",
			}

			ctx := context.Background()
			err := client.Set(ctx, tt.kv)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKVStoreClient_Delete(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		setupMock      func(*MockKeyValueServiceClient)
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful delete",
			key:  "test-key",
			setupMock: func(m *MockKeyValueServiceClient) {
				m.DeleteFunc = func(ctx context.Context, in *keyvalue.DeleteRequest, opts ...grpc.CallOption) (*keyvalue.DeleteResponse, error) {
					return &keyvalue.DeleteResponse{
						Success: true,
					}, nil
				}
			},
		},
		{
			name: "delete operation failed",
			key:  "test-key",
			setupMock: func(m *MockKeyValueServiceClient) {
				m.DeleteFunc = func(ctx context.Context, in *keyvalue.DeleteRequest, opts ...grpc.CallOption) (*keyvalue.DeleteResponse, error) {
					return &keyvalue.DeleteResponse{
						Success: false,
						Error:   "delete failed",
					}, nil
				}
			},
			expectError:    true,
			expectedErrMsg: "delete operation failed: delete failed",
		},
		{
			name: "grpc error",
			key:  "test-key",
			setupMock: func(m *MockKeyValueServiceClient) {
				m.DeleteFunc = func(ctx context.Context, in *keyvalue.DeleteRequest, opts ...grpc.CallOption) (*keyvalue.DeleteResponse, error) {
					return nil, status.Errorf(codes.Internal, "service unavailable")
				}
			},
			expectError:    true,
			expectedErrMsg: "failed to delete key test-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockKeyValueServiceClient{}
			tt.setupMock(mockClient)

			// Create client with mock
			client := &KVStoreClient{
				client: mockClient,
				addr:   "mock-address",
			}

			ctx := context.Background()
			err := client.Delete(ctx, tt.key)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKVStoreClient_Health(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockKeyValueServiceClient)
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "healthy service",
			setupMock: func(m *MockKeyValueServiceClient) {
				m.HealthFunc = func(ctx context.Context, in *keyvalue.HealthRequest, opts ...grpc.CallOption) (*keyvalue.HealthResponse, error) {
					return &keyvalue.HealthResponse{
						Status:    "healthy",
						Timestamp: time.Now().Unix(),
					}, nil
				}
			},
		},
		{
			name: "unhealthy service",
			setupMock: func(m *MockKeyValueServiceClient) {
				m.HealthFunc = func(ctx context.Context, in *keyvalue.HealthRequest, opts ...grpc.CallOption) (*keyvalue.HealthResponse, error) {
					return &keyvalue.HealthResponse{
						Status:    "unhealthy",
						Timestamp: time.Now().Unix(),
					}, nil
				}
			},
			expectError:    true,
			expectedErrMsg: "service is not healthy: unhealthy",
		},
		{
			name: "grpc error",
			setupMock: func(m *MockKeyValueServiceClient) {
				m.HealthFunc = func(ctx context.Context, in *keyvalue.HealthRequest, opts ...grpc.CallOption) (*keyvalue.HealthResponse, error) {
					return nil, status.Errorf(codes.Internal, "service unavailable")
				}
			},
			expectError:    true,
			expectedErrMsg: "health check failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockKeyValueServiceClient{}
			tt.setupMock(mockClient)

			// Create client with mock
			client := &KVStoreClient{
				client: mockClient,
				addr:   "mock-address",
			}

			ctx := context.Background()
			err := client.Health(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
