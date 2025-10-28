package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	wishlistgen "github.com/theseems/wili/backend/services/wishlist/gen"
)

const (
	MaxWishlistTitleLength       = 200
	MaxWishlistDescriptionLength = 2000
	MaxItemNameLength            = 300
	MaxItemDescriptionLength     = 2000
	MinItemNameLength            = 1
	MinWishlistTitleLength       = 1
)

// ValidationError represents a validation error with field-specific details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", v.Field, v.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "no validation errors"
	}

	var messages []string
	for _, err := range v {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// ValidateCreateWishlistRequest validates a create wishlist request
func ValidateCreateWishlistRequest(req wishlistgen.CreateWishlistRequest) ValidationErrors {
	var errors ValidationErrors

	if err := validateStringField("title", req.Title, MinWishlistTitleLength, MaxWishlistTitleLength, true); err != nil {
		errors = append(errors, *err)
	}

	if req.Description != nil {
		if err := validateStringField("description", *req.Description, 0, MaxWishlistDescriptionLength, false); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// ValidateUpdateWishlistRequest validates an update wishlist request
func ValidateUpdateWishlistRequest(req wishlistgen.UpdateWishlistRequest) ValidationErrors {
	var errors ValidationErrors

	if req.Title != nil {
		if err := validateStringField("title", *req.Title, MinWishlistTitleLength, MaxWishlistTitleLength, true); err != nil {
			errors = append(errors, *err)
		}
	}

	// Validate description if provided
	if req.Description != nil {
		if err := validateStringField("description", *req.Description, 0, MaxWishlistDescriptionLength, false); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// ValidateCreateWishlistItemRequest validates a create wishlist item request
func ValidateCreateWishlistItemRequest(req wishlistgen.CreateWishlistItemRequest) ValidationErrors {
	var errors ValidationErrors

	if err := validateStringField("type", req.Type, 1, 50, true); err != nil {
		errors = append(errors, *err)
	}

	if dataErrors := validateItemData(req.Data); len(dataErrors) > 0 {
		errors = append(errors, dataErrors...)
	}

	return errors
}

// ValidateUpdateWishlistItemRequest validates an update wishlist item request
func ValidateUpdateWishlistItemRequest(req wishlistgen.UpdateWishlistItemRequest) ValidationErrors {
	var errors ValidationErrors

	if req.Type != nil {
		if err := validateStringField("type", *req.Type, 1, 50, true); err != nil {
			errors = append(errors, *err)
		}
	}

	// Validate data payload if provided
	if req.Data != nil {
		if dataErrors := validateItemData(*req.Data); len(dataErrors) > 0 {
			errors = append(errors, dataErrors...)
		}
	}

	return errors
}

// validateItemData validates the item data payload structure
func validateItemData(data wishlistgen.WishlistItemData) ValidationErrors {
	var errors ValidationErrors

	// Validate name (required)
	if err := validateStringField("data.name", data.Name, MinItemNameLength, MaxItemNameLength, true); err != nil {
		errors = append(errors, *err)
	}

	// Validate description (optional)
	if data.Description != nil {
		if err := validateStringField("data.description", *data.Description, 0, MaxItemDescriptionLength, false); err != nil {
			errors = append(errors, *err)
		}
	}

	// Validate URL (optional)
	if data.Url != nil {
		if !isValidURL(*data.Url) {
			errors = append(errors, ValidationError{
				Field:   "data.url",
				Message: "url must be a valid HTTP/HTTPS URL",
			})
		}
	}

	return errors
}

func validateStringField(fieldName, value string, minLength, maxLength int, required bool) *ValidationError {
	if required && strings.TrimSpace(value) == "" {
		return &ValidationError{
			Field:   fieldName,
			Message: "field is required and cannot be empty",
		}
	}

	if !required && strings.TrimSpace(value) == "" {
		return nil
	}

	if minLength > 0 && utf8.RuneCountInString(value) < minLength {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("must be at least %d characters long", minLength),
		}
	}

	if maxLength > 0 && utf8.RuneCountInString(value) > maxLength {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("must not exceed %d characters", maxLength),
		}
	}

	return nil
}

func isValidURL(urlStr string) bool {
	urlStr = strings.TrimSpace(urlStr)
	return strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://")
}
