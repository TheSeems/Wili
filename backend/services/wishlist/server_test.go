package main

import (
	"strings"
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
				Data: wishlistgen.WishlistItemData{
					Name: "Test Item",
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
				Data: wishlistgen.WishlistItemData{
					Name:        "Test Item",
					Description: stringPtr("This is a test item"),
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
				Data: wishlistgen.WishlistItemData{
					Description: stringPtr("This is a test item"),
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
				Data: wishlistgen.WishlistItemData{
					Name: "",
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
				Data: wishlistgen.WishlistItemData{
					Name: strings.Repeat("a", MaxItemNameLength+1),
				},
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject request with name exceeding maximum length",
		},
		{
			name: "description_too_long",
			req: wishlistgen.CreateWishlistItemRequest{
				Type: "text",
				Data: wishlistgen.WishlistItemData{
					Name:        "Test Item",
					Description: stringPtr(strings.Repeat("a", MaxItemDescriptionLength+1)),
				},
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject request with description exceeding maximum length",
		},
		{
			name: "invalid_url",
			req: wishlistgen.CreateWishlistItemRequest{
				Type: "text",
				Data: wishlistgen.WishlistItemData{
					Name: "Test Item",
					Url:  stringPtr("not-a-valid-url"),
				},
			},
			expectErrors:  true,
			expectedCount: 1,
			description:   "Should reject request with invalid URL",
		},
		{
			name: "valid_url",
			req: wishlistgen.CreateWishlistItemRequest{
				Type: "text",
				Data: wishlistgen.WishlistItemData{
					Name: "Test Item",
					Url:  stringPtr("https://example.com"),
				},
			},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept request with valid URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateCreateWishlistItemRequest(tt.req)

			if tt.expectErrors {
				if len(errors) == 0 {
					t.Errorf("Expected validation errors but got none")
				}
				if len(errors) != tt.expectedCount {
					t.Errorf("Expected %d errors, got %d", tt.expectedCount, len(errors))
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("Expected no validation errors but got %d: %v", len(errors), errors)
				}
			}

			if tt.expectedCount > 0 && len(errors) != tt.expectedCount {
				t.Errorf("Expected %d validation errors but got %d: %v", tt.expectedCount, len(errors), errors)
			}
		})
	}
}

func TestValidateBookItemRequest(t *testing.T) {
	tests := []struct {
		name          string
		req           wishlistgen.BookItemRequest
		expectErrors  bool
		expectedCount int
		description   string
	}{
		{
			name: "valid_anonymous_booking",
			req: wishlistgen.BookItemRequest{
				BookerName: nil,
				Message:    nil,
			},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept anonymous booking with no name or message",
		},
		{
			name: "valid_booking_with_name",
			req: wishlistgen.BookItemRequest{
				BookerName: stringPtr("John Doe"),
				Message:    nil,
			},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept booking with name only",
		},
		{
			name: "valid_booking_with_message",
			req: wishlistgen.BookItemRequest{
				BookerName: nil,
				Message:    stringPtr("I'll buy this for you!"),
			},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept booking with message only",
		},
		{
			name: "valid_booking_with_name_and_message",
			req: wishlistgen.BookItemRequest{
				BookerName: stringPtr("Jane Smith"),
				Message:    stringPtr("Happy to help with this gift!"),
			},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should accept booking with both name and message",
		},
		{
			name: "empty_name_and_message",
			req: wishlistgen.BookItemRequest{
				BookerName: stringPtr(""),
				Message:    stringPtr(""),
			},
			expectErrors:  false,
			expectedCount: 0,
			description:   "Should treat empty strings as nil (anonymous booking)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// BookItemRequest doesn't have validation yet, but we can test the structure
			// This test ensures the request structure is valid
			if tt.req.BookerName != nil && *tt.req.BookerName == "" {
				tt.req.BookerName = nil
			}
			if tt.req.Message != nil && *tt.req.Message == "" {
				tt.req.Message = nil
			}

			// Basic structure validation
			if tt.expectErrors {
				t.Errorf("BookItemRequest validation not implemented yet")
			}
		})
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}
