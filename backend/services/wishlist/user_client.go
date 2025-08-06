package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// UserClient wraps HTTP calls to the user service
type UserClient struct {
	baseURL    string
	httpClient *http.Client
}

// ValidateTokenRequest represents the request to validate a JWT token
type ValidateTokenRequest struct {
	Token string `json:"token"`
}

// ValidateTokenResponse represents the response from token validation
type ValidateTokenResponse struct {
	Valid bool     `json:"valid"`
	User  UserInfo `json:"user"`
}

// UserInfo represents user information returned from validation
type UserInfo struct {
	Id          openapi_types.UUID `json:"id"`
	DisplayName string             `json:"displayName"`
	AvatarUrl   *string            `json:"avatarUrl"`
}

// NewUserClient creates a new user service client
func NewUserClient(baseURL string) *UserClient {
	return &UserClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{},
	}
}

// ValidateToken calls the user service to validate a JWT token and extract user info
func (c *UserClient) ValidateToken(ctx context.Context, token string) (*UserInfo, error) {
	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")

	log.Printf("[USER_CLIENT] Validating token with user service at %s", c.baseURL)

	reqBody := ValidateTokenRequest{
		Token: token,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/auth/validate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[USER_CLIENT] Failed to connect to user service: %v", err)
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[USER_CLIENT] User service returned status %d", resp.StatusCode)
		return nil, fmt.Errorf("user service returned status: %d", resp.StatusCode)
	}

	var validateResp ValidateTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&validateResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !validateResp.Valid {
		log.Printf("[USER_CLIENT] Token validation failed")
		return nil, fmt.Errorf("invalid token")
	}

	log.Printf("[USER_CLIENT] Token validated successfully for user: %s", validateResp.User.DisplayName)
	return &validateResp.User, nil
}
