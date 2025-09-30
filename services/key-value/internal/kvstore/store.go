package kvstore

import (
	"errors"
	"sync"
)

// Storer interface defines the methods for the key-value store
type Storer interface {
	Get(key string) (string, error)
	Set(key string, value string) error
	Delete(key string) error
}

// InMemoryStore implements the Storer interface with a thread safe map
type InMemoryStore struct {
	mutex sync.RWMutex
	store map[string]string
}

// NewInMemoryStore creates a new InMemoryStore
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		mutex: sync.RWMutex{},
		store: make(map[string]string),
	}
}

// Get retrieves a value by key
func (s *InMemoryStore) Get(key string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if _, ok := s.store[key]; !ok {
		return "", errors.New("key not found")
	}
	return s.store[key], nil
}

// Set stores a key-value pair as a upsert operation
func (s *InMemoryStore) Set(key string, value string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.store[key] = value
	return nil
}

// Delete removes a key-value pair if the key does not exist, it is a no-op
func (s *InMemoryStore) Delete(key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.store, key)
	return nil
}
