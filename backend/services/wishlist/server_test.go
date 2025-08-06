package main

import (
	"testing"

	wishlistgen "github.com/theseems/wili/backend/services/wishlist/gen"
)

func TestValidateCreateWishlistItemRequest(t *testing.T) {
	tests := []struct {
		name          string
		req           wishlistgen.CreateWishlistItemRequest
		expectErrors  bool
		expectedCount int
		description   string
	}{
		{
			name: "valid_request_with_name_only",
			req: wishlistgen.CreateWishlistItemRequest{
				Type: "text",
				Data: map[string]interface{}{
					"name": "Test Item",
				},
			},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept valid request with only name (description is optional)",
		},
		{
			name: "valid_request_with_name_and_description",
			req: wishlistgen.CreateWishlistItemRequest{
				Type: "text",
				Data: map[string]interface{}{
					"name":        "Test Item",
					"description": "This is a test item",
				},
			},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept valid request with both name and description",
		},
		{
			name: "missing_name",
			req: wishlistgen.CreateWishlistItemRequest{
				Type: "text",
				Data: map[string]interface{}{
					"description": "This is a test item",
				},
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject request missing required name field",
		},
		{
			name: "empty_name",
			req: wishlistgen.CreateWishlistItemRequest{
				Type: "text",
				Data: map[string]interface{}{
					"name":        "",
					"description": "This is a test item",
				},
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject request with empty name",
		},
		{
			name: "name_too_long",
			req: wishlistgen.CreateWishlistItemRequest{
				Type: "text",
				Data: map[string]interface{}{
					"name":        string(make([]byte, 301)), // 301 characters, exceeds limit
					"description": "This is a test item",
				},
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject request with name exceeding 300 characters",
		},
		{
			name: "description_too_long",
			req: wishlistgen.CreateWishlistItemRequest{
				Type: "text",
				Data: map[string]interface{}{
					"name":        "Test Item",
					"description": string(make([]byte, 2001)), // 2001 characters, exceeds limit
				},
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject request with description exceeding 2000 characters",
		},
		{
			name: "invalid_url",
			req: wishlistgen.CreateWishlistItemRequest{
				Type: "text",
				Data: map[string]interface{}{
					"name":        "Test Item",
					"description": "This is a test item",
					"url":         "not-a-valid-url",
				},
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject request with invalid URL format",
		},
		{
			name: "valid_http_url",
			req: wishlistgen.CreateWishlistItemRequest{
				Type: "text",
				Data: map[string]interface{}{
					"name":        "Test Item",
					"description": "This is a test item",
					"url":         "http://example.com",
				},
			},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept valid HTTP URL",
		},
		{
			name: "valid_https_url",
			req: wishlistgen.CreateWishlistItemRequest{
				Type: "text",
				Data: map[string]interface{}{
					"name":        "Test Item",
					"description": "This is a test item",
					"url":         "https://example.com",
				},
			},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept valid HTTPS URL",
		},
		{
			name: "name_not_string",
			req: wishlistgen.CreateWishlistItemRequest{
				Type: "text",
				Data: map[string]interface{}{
					"name":        123,
					"description": "This is a test item",
				},
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject request with non-string name",
		},
		{
			name: "description_not_string",
			req: wishlistgen.CreateWishlistItemRequest{
				Type: "text",
				Data: map[string]interface{}{
					"name":        "Test Item",
					"description": 123,
				},
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject request with non-string description",
		},
		{
			name: "empty_type",
			req: wishlistgen.CreateWishlistItemRequest{
				Type: "",
				Data: map[string]interface{}{
					"name": "Test Item",
				},
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject request with empty type",
		},
		{
			name: "type_too_long",
			req: wishlistgen.CreateWishlistItemRequest{
				Type: string(make([]byte, 51)), // 51 characters, exceeds limit
				Data: map[string]interface{}{
					"name": "Test Item",
				},
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject request with type exceeding 50 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing: %s", tt.description)
			
			errors := ValidateCreateWishlistItemRequest(tt.req)
			
			if tt.expectErrors && len(errors) == 0 {
				t.Errorf("Expected validation errors but got none")
			}
			
			if !tt.expectErrors && len(errors) > 0 {
				t.Errorf("Expected no validation errors but got: %v", errors)
			}
			
			if tt.expectedCount > 0 && len(errors) != tt.expectedCount {
				t.Errorf("Expected %d validation errors but got %d: %v", tt.expectedCount, len(errors), errors)
			}
		})
	}
}

func TestValidateUpdateWishlistItemRequest(t *testing.T) {
	tests := []struct {
		name          string
		req           wishlistgen.UpdateWishlistItemRequest
		expectErrors  bool
		expectedCount int
		description   string
	}{
		{
			name: "valid_update_with_name_only",
			req: wishlistgen.UpdateWishlistItemRequest{
				Type: stringPtr("text"),
				Data: mapPtr(map[string]interface{}{
					"name": "Updated Item",
				}),
			},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept valid update with only name",
		},
		{
			name: "valid_update_with_name_and_description",
			req: wishlistgen.UpdateWishlistItemRequest{
				Type: stringPtr("text"),
				Data: mapPtr(map[string]interface{}{
					"name":        "Updated Item",
					"description": "Updated description",
				}),
			},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept valid update with name and description",
		},
		{
			name: "empty_update_request",
			req: wishlistgen.UpdateWishlistItemRequest{},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept empty update request (partial updates allowed)",
		},
		{
			name: "update_with_missing_name_in_data",
			req: wishlistgen.UpdateWishlistItemRequest{
				Data: mapPtr(map[string]interface{}{
					"description": "Updated description",
				}),
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject update with data but missing name",
		},
		{
			name: "update_with_invalid_url",
			req: wishlistgen.UpdateWishlistItemRequest{
				Data: mapPtr(map[string]interface{}{
					"name": "Updated Item",
					"url":  "invalid-url",
				}),
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject update with invalid URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing: %s", tt.description)
			
			errors := ValidateUpdateWishlistItemRequest(tt.req)
			
			if tt.expectErrors && len(errors) == 0 {
				t.Errorf("Expected validation errors but got none")
			}
			
			if !tt.expectErrors && len(errors) > 0 {
				t.Errorf("Expected no validation errors but got: %v", errors)
			}
			
			if tt.expectedCount > 0 && len(errors) != tt.expectedCount {
				t.Errorf("Expected %d validation errors but got %d: %v", tt.expectedCount, len(errors), errors)
			}
		})
	}
}

func TestValidateCreateWishlistRequest(t *testing.T) {
	tests := []struct {
		name          string
		req           wishlistgen.CreateWishlistRequest
		expectErrors  bool
		expectedCount int
		description   string
	}{
		{
			name: "valid_request_with_title_only",
			req: wishlistgen.CreateWishlistRequest{
				Title: "Test Wishlist",
			},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept valid request with only title",
		},
		{
			name: "valid_request_with_title_and_description",
			req: wishlistgen.CreateWishlistRequest{
				Title:       "Test Wishlist",
				Description: stringPtr("This is a test wishlist"),
			},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept valid request with title and description",
		},
		{
			name: "empty_title",
			req: wishlistgen.CreateWishlistRequest{
				Title:       "",
				Description: stringPtr("This is a test wishlist"),
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject request with empty title",
		},
		{
			name: "title_too_long",
			req: wishlistgen.CreateWishlistRequest{
				Title:       string(make([]byte, 201)), // 201 characters, exceeds limit
				Description: stringPtr("This is a test wishlist"),
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject request with title exceeding 200 characters",
		},
		{
			name: "description_too_long",
			req: wishlistgen.CreateWishlistRequest{
				Title:       "Test Wishlist",
				Description: stringPtr(string(make([]byte, 2001))), // 2001 characters, exceeds limit
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject request with description exceeding 2000 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing: %s", tt.description)
			
			errors := ValidateCreateWishlistRequest(tt.req)
			
			if tt.expectErrors && len(errors) == 0 {
				t.Errorf("Expected validation errors but got none")
			}
			
			if !tt.expectErrors && len(errors) > 0 {
				t.Errorf("Expected no validation errors but got: %v", errors)
			}
			
			if tt.expectedCount > 0 && len(errors) != tt.expectedCount {
				t.Errorf("Expected %d validation errors but got %d: %v", tt.expectedCount, len(errors), errors)
			}
		})
	}
}

func TestValidateUpdateWishlistRequest(t *testing.T) {
	tests := []struct {
		name          string
		req           wishlistgen.UpdateWishlistRequest
		expectErrors  bool
		expectedCount int
		description   string
	}{
		{
			name: "valid_update_title_only",
			req: wishlistgen.UpdateWishlistRequest{
				Title: stringPtr("Updated Wishlist"),
			},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept valid update with only title",
		},
		{
			name: "valid_update_description_only",
			req: wishlistgen.UpdateWishlistRequest{
				Description: stringPtr("Updated description"),
			},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept valid update with only description",
		},
		{
			name: "empty_update_request",
			req: wishlistgen.UpdateWishlistRequest{},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept empty update request (partial updates allowed)",
		},
		{
			name: "update_with_empty_title",
			req: wishlistgen.UpdateWishlistRequest{
				Title: stringPtr(""),
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject update with empty title",
		},
		{
			name: "update_with_title_too_long",
			req: wishlistgen.UpdateWishlistRequest{
				Title: stringPtr(string(make([]byte, 201))),
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject update with title exceeding 200 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing: %s", tt.description)
			
			errors := ValidateUpdateWishlistRequest(tt.req)
			
			if tt.expectErrors && len(errors) == 0 {
				t.Errorf("Expected validation errors but got none")
			}
			
			if !tt.expectErrors && len(errors) > 0 {
				t.Errorf("Expected no validation errors but got: %v", errors)
			}
			
			if tt.expectedCount > 0 && len(errors) != tt.expectedCount {
				t.Errorf("Expected %d validation errors but got %d: %v", tt.expectedCount, len(errors), errors)
			}
		})
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func mapPtr(m map[string]interface{}) *map[string]interface{} {
	return &m
}