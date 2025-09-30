package kvstore

import (
	"testing"
)

func TestInMemoryStore_Set(t *testing.T) {
	store := NewInMemoryStore()

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"basic set", "key1", "value1"},
		{"empty value", "key2", ""},
		{"empty key", "", "value3"},
		{"overwrite existing", "key1", "new_value1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Set(tt.key, tt.value)
			if err != nil {
				t.Errorf("Set() error = %v, want nil", err)
			}

			// Verify the value was stored
			if stored, exists := store.store[tt.key]; !exists || stored != tt.value {
				t.Errorf("Expected store[%s] = %s, got %s (exists: %v)", tt.key, tt.value, stored, exists)
			}
		})
	}
}

func TestInMemoryStore_Get(t *testing.T) {
	store := NewInMemoryStore()

	// Setup test data
	store.Set("existing_key", "existing_value")
	store.Set("empty_value", "")

	tests := []struct {
		name      string
		key       string
		wantValue string
		wantError bool
	}{
		{"existing key", "existing_key", "existing_value", false},
		{"empty value", "empty_value", "", false},
		{"non-existing key", "non_existing", "", true},
		{"empty key", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := store.Get(tt.key)

			if tt.wantError {
				if err == nil {
					t.Errorf("Get() error = nil, want error")
				}
				if err.Error() != "key not found" {
					t.Errorf("Get() error = %v, want 'key not found'", err)
				}
			} else {
				if err != nil {
					t.Errorf("Get() error = %v, want nil", err)
				}
				if value != tt.wantValue {
					t.Errorf("Get() value = %v, want %v", value, tt.wantValue)
				}
			}
		})
	}
}

func TestInMemoryStore_Delete(t *testing.T) {
	store := NewInMemoryStore()

	// Setup test data
	store.Set("key_to_delete", "value")
	store.Set("another_key", "another_value")

	tests := []struct {
		name      string
		key       string
		wantError bool
	}{
		{"delete existing key", "key_to_delete", false},
		{"delete non-existing key (idempotent)", "non_existing", false},
		{"delete empty key (idempotent)", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Delete(tt.key)

			if tt.wantError {
				if err == nil {
					t.Errorf("Delete() error = nil, want error")
				}
				if err.Error() != "key not found" {
					t.Errorf("Delete() error = %v, want 'key not found'", err)
				}
			} else {
				if err != nil {
					t.Errorf("Delete() error = %v, want nil", err)
				}

				// Verify the key was actually deleted
				if _, exists := store.store[tt.key]; exists {
					t.Errorf("Key %s should have been deleted but still exists", tt.key)
				}
			}
		})
	}

	// Verify other keys weren't affected
	if _, exists := store.store["another_key"]; !exists {
		t.Error("Other keys should not be affected by delete operation")
	}
}
