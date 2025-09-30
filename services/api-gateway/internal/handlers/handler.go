package handlers

import (
	"context"
	"key-value/client"
	"key-value/shared/models"
)

// KVStoreInterface defines the interface for key-value store operations
type KVStoreInterface interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, kv models.KeyValue) error
	Delete(ctx context.Context, key string) error
	Health(ctx context.Context) error
	Close() error
}

type Handler struct {
	kvstoreClient KVStoreInterface
}

func NewHandler(kvstoreClient KVStoreInterface) *Handler {
	return &Handler{
		kvstoreClient: kvstoreClient,
	}
}

func NewHandlerWithClient(kvstoreClient *client.KVStoreClient) *Handler {
	return &Handler{
		kvstoreClient: kvstoreClient,
	}
}
