package client

import (
	"context"
	"fmt"
	"key-value/proto/keyvalue"
	"key-value/shared/models"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// KVStoreClient wraps the gRPC client for the key-value service
type KVStoreClient struct {
	client keyvalue.KeyValueServiceClient
	conn   *grpc.ClientConn
	addr   string
}

// NewKVStoreClient creates a new client connection to the key-value service
func NewKVStoreClient(address string) (*KVStoreClient, error) {
	// Set up connection options
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(address, options...) // Use WithInsecure for development, use credentials for production
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	client := keyvalue.NewKeyValueServiceClient(conn)

	return &KVStoreClient{
		client: client,
		conn:   conn,
		addr:   address,
	}, nil
}

// Get retrieves a value by key
func (c *KVStoreClient) Get(ctx context.Context, key string) (string, bool, error) {
	req := &keyvalue.GetRequest{
		Key: key,
	}

	resp, err := c.client.Get(ctx, req)
	if err != nil {
		return "", false, fmt.Errorf("failed to get key %s: %w", key, err)
	}

	return resp.Value, resp.Found, nil
}

// Set stores a key-value pair
func (c *KVStoreClient) Set(ctx context.Context, kv models.KeyValue) error {
	req := &keyvalue.SetRequest{
		Key:   kv.Key,
		Value: kv.Value,
	}

	resp, err := c.client.Set(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", kv.Key, err)
	}

	if !resp.Success {
		return fmt.Errorf("set operation failed: %s", resp.Error)
	}

	return nil
}

// Delete removes a key-value pair
func (c *KVStoreClient) Delete(ctx context.Context, key string) error {
	req := &keyvalue.DeleteRequest{
		Key: key,
	}

	resp, err := c.client.Delete(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	if !resp.Success {
		return fmt.Errorf("delete operation failed: %s", resp.Error)
	}

	return nil
}

// Health provides a health check endpoint
func (c *KVStoreClient) Health(ctx context.Context) error {
	req := &keyvalue.HealthRequest{}

	resp, err := c.client.Health(ctx, req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.Status != "healthy" {
		return fmt.Errorf("service is not healthy: %s", resp.Status)
	}

	return nil
}

func (c *KVStoreClient) Close() error {
	return c.conn.Close()
}
