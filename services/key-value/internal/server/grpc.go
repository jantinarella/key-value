package server

import (
	"context"
	"key-value/services/key-value/internal/kvstore"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"key-value/proto/keyvalue"
)

// KeyValueServer implements the gRPC KeyValueService
type KeyValueServer struct {
	keyvalue.UnimplementedKeyValueServiceServer
	store kvstore.Storer
}

// NewKeyValueServer creates a new gRPC server instance
func NewKeyValueServer(store kvstore.Storer) *KeyValueServer {
	return &KeyValueServer{
		store: store,
	}
}

// Get retrieves a value by key
func (s *KeyValueServer) Get(ctx context.Context, req *keyvalue.GetRequest) (*keyvalue.GetResponse, error) {
	if req.Key == "" {
		return nil, status.Errorf(codes.InvalidArgument, "key cannot be empty")
	}

	value, err := s.store.Get(req.Key)
	if err != nil {
		// Check if it's a "key not found" error
		if err.Error() == "key not found" {
			return &keyvalue.GetResponse{
				Value: "",
				Found: false,
				Error: "",
			}, nil
		}
		// Other errors
		return nil, status.Errorf(codes.Internal, "service failed to get value: %v", err)
	}

	return &keyvalue.GetResponse{
		Value: value,
		Found: true,
		Error: "",
	}, nil
}

// Set stores a key-value pair
func (s *KeyValueServer) Set(ctx context.Context, req *keyvalue.SetRequest) (*keyvalue.SetResponse, error) {
	if req.Key == "" {
		return nil, status.Errorf(codes.InvalidArgument, "key cannot be empty")
	}

	err := s.store.Set(req.Key, req.Value)
	if err != nil {
		return &keyvalue.SetResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &keyvalue.SetResponse{
		Success: true,
		Error:   "",
	}, nil
}

// Delete removes a key-value pair
func (s *KeyValueServer) Delete(ctx context.Context, req *keyvalue.DeleteRequest) (*keyvalue.DeleteResponse, error) {
	if req.Key == "" {
		return nil, status.Errorf(codes.InvalidArgument, "key cannot be empty")
	}

	err := s.store.Delete(req.Key)
	if err != nil {
		return &keyvalue.DeleteResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &keyvalue.DeleteResponse{
		Success: true,
		Error:   "",
	}, nil
}

// Health provides a health check endpoint
func (s *KeyValueServer) Health(ctx context.Context, req *keyvalue.HealthRequest) (*keyvalue.HealthResponse, error) {
	return &keyvalue.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
	}, nil
}
