package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	openapi_types "github.com/oapi-codegen/runtime/types"

	wishlistgen "github.com/theseems/wili/backend/services/wishlist/gen"
)

type WishlistServer struct {
	repo       *MongoRepo
	userClient *UserClient
	logger     *Logger
}

func NewWishlistServer(repo *MongoRepo, userClient *UserClient) *WishlistServer {
	return &WishlistServer{
		repo:       repo,
		userClient: userClient,
		logger:     NewLogger("WISHLIST"),
	}
}

// Helper function to extract user ID from JWT token via user service
func (s *WishlistServer) extractUserID(r *http.Request) (openapi_types.UUID, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		s.logger.LogUnauthorized(r, "token_extraction", "missing authorization header")
		return openapi_types.UUID{}, fmt.Errorf("missing authorization header")
	}

	// Call user service to validate token and get user info
	userInfo, err := s.userClient.ValidateToken(r.Context(), authHeader)
	if err != nil {
		s.logger.LogUnauthorized(r, "token_validation", fmt.Sprintf("token validation failed: %v", err))
		return openapi_types.UUID{}, fmt.Errorf("token validation failed: %w", err)
	}

	return userInfo.Id, nil
}

func (s *WishlistServer) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (s *WishlistServer) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (s *WishlistServer) writeValidationErrors(w http.ResponseWriter, errors ValidationErrors) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := map[string]interface{}{
		"error":   "Validation failed",
		"details": errors,
	}

	json.NewEncoder(w).Encode(response)
}

// List wishlists of the authenticated user
func (s *WishlistServer) GetWishlists(w http.ResponseWriter, r *http.Request) {
	s.logger.LogRequest(r, nil, "get_wishlists")

	userID, err := s.extractUserID(r)
	if err != nil {
		s.writeError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	wishlists, err := s.repo.GetWishlistsByUser(r.Context(), userID)
	if err != nil {
		s.logger.LogError(&userID, "get_wishlists", err, "failed to retrieve wishlists from database")
		s.writeError(w, http.StatusInternalServerError, "Failed to retrieve wishlists")
		return
	}

	s.logger.LogSuccess(&userID, "get_wishlists", fmt.Sprintf("retrieved %d wishlists", len(wishlists)))
	response := map[string]interface{}{
		"wishlists": wishlists,
	}
	s.writeJSON(w, http.StatusOK, response)
}

// Create a new wishlist for the authenticated user
func (s *WishlistServer) PostWishlists(w http.ResponseWriter, r *http.Request) {
	s.logger.LogRequest(r, nil, "create_wishlist")

	userID, err := s.extractUserID(r)
	if err != nil {
		s.writeError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	var req wishlistgen.CreateWishlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.LogBadRequest(&userID, "create_wishlist", fmt.Sprintf("malformed JSON: %v", err))
		s.writeError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	// Validate request
	if validationErrors := ValidateCreateWishlistRequest(req); len(validationErrors) > 0 {
		s.logger.LogValidationError(&userID, "create_wishlist", validationErrors)
		s.writeValidationErrors(w, validationErrors)
		return
	}

	wishlist, err := s.repo.CreateWishlist(r.Context(), userID, req)
	if err != nil {
		s.logger.LogError(&userID, "create_wishlist", err, "failed to create wishlist in database")
		s.writeError(w, http.StatusInternalServerError, "Failed to create wishlist")
		return
	}

	s.logger.LogSuccess(&userID, "create_wishlist", fmt.Sprintf("created wishlist '%s'", req.Title))
	s.writeJSON(w, http.StatusCreated, wishlist)
}

// Get a wishlist by ID (public endpoint)
func (s *WishlistServer) GetWishlistsWishlistId(w http.ResponseWriter, r *http.Request, wishlistId openapi_types.UUID) {
	s.logger.LogRequest(r, nil, "get_wishlist")

	wishlist, err := s.repo.GetWishlistByID(r.Context(), wishlistId)
	if err != nil {
		s.logger.LogError(nil, "get_wishlist", err, fmt.Sprintf("failed to retrieve wishlist %s from database", wishlistId.String()))
		s.writeError(w, http.StatusInternalServerError, "Failed to retrieve wishlist")
		return
	}

	if wishlist == nil {
		s.logger.LogNotFound(nil, "wishlist", wishlistId.String())
		s.writeError(w, http.StatusNotFound, "Wishlist not found")
		return
	}

	s.logger.LogSuccess(nil, "get_wishlist", fmt.Sprintf("retrieved wishlist '%s' (%s)", wishlist.Title, wishlistId.String()))
	s.writeJSON(w, http.StatusOK, wishlist)
}

// Delete a wishlist (owner only)
func (s *WishlistServer) DeleteWishlistsWishlistId(w http.ResponseWriter, r *http.Request, wishlistId openapi_types.UUID) {
	s.logger.LogRequest(r, nil, "delete_wishlist")

	userID, err := s.extractUserID(r)
	if err != nil {
		s.writeError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	err = s.repo.DeleteWishlist(r.Context(), wishlistId, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.logger.LogNotFound(&userID, "wishlist", wishlistId.String())
			s.writeError(w, http.StatusNotFound, "Wishlist not found or not owned by user")
			return
		}
		s.logger.LogError(&userID, "delete_wishlist", err, fmt.Sprintf("failed to delete wishlist %s", wishlistId.String()))
		s.writeError(w, http.StatusInternalServerError, "Failed to delete wishlist")
		return
	}

	s.logger.LogSuccess(&userID, "delete_wishlist", fmt.Sprintf("deleted wishlist %s", wishlistId.String()))
	w.WriteHeader(http.StatusNoContent)
}

// Add an item to a wishlist (owner only)
func (s *WishlistServer) PostWishlistsWishlistIdItems(w http.ResponseWriter, r *http.Request, wishlistId openapi_types.UUID) {
	s.logger.LogRequest(r, nil, "add_item")

	userID, err := s.extractUserID(r)
	if err != nil {
		s.writeError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	var req wishlistgen.CreateWishlistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.LogBadRequest(&userID, "add_item", fmt.Sprintf("malformed JSON for wishlist %s: %v", wishlistId.String(), err))
		s.writeError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	// Validate request
	if validationErrors := ValidateCreateWishlistItemRequest(req); len(validationErrors) > 0 {
		s.logger.LogValidationError(&userID, "add_item", validationErrors)
		s.writeValidationErrors(w, validationErrors)
		return
	}

	item, err := s.repo.AddItemToWishlist(r.Context(), wishlistId, userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.logger.LogNotFound(&userID, "wishlist", wishlistId.String())
			s.writeError(w, http.StatusNotFound, "Wishlist not found or not owned by user")
			return
		}
		s.logger.LogError(&userID, "add_item", err, fmt.Sprintf("failed to add item to wishlist %s", wishlistId.String()))
		s.writeError(w, http.StatusInternalServerError, "Failed to add item to wishlist")
		return
	}

	itemName := "unknown"
	if nameVal, ok := req.Data["name"]; ok {
		if nameStr, ok := nameVal.(string); ok {
			itemName = nameStr
		}
	}
	s.logger.LogSuccess(&userID, "add_item", fmt.Sprintf("added item '%s' (type: %s) to wishlist %s", itemName, req.Type, wishlistId.String()))
	s.writeJSON(w, http.StatusCreated, item)
}

// Update a wishlist item (owner only)
func (s *WishlistServer) PutWishlistsWishlistIdItemsItemId(w http.ResponseWriter, r *http.Request, wishlistId openapi_types.UUID, itemId openapi_types.UUID) {
	s.logger.LogRequest(r, nil, "update_item")

	userID, err := s.extractUserID(r)
	if err != nil {
		s.writeError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	var req wishlistgen.UpdateWishlistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.LogBadRequest(&userID, "update_item", fmt.Sprintf("malformed JSON for item %s in wishlist %s: %v", itemId.String(), wishlistId.String(), err))
		s.writeError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	// Validate request
	if validationErrors := ValidateUpdateWishlistItemRequest(req); len(validationErrors) > 0 {
		s.logger.LogValidationError(&userID, "update_item", validationErrors)
		s.writeValidationErrors(w, validationErrors)
		return
	}

	item, err := s.repo.UpdateWishlistItem(r.Context(), wishlistId, itemId, userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.logger.LogNotFound(&userID, "item", fmt.Sprintf("%s in wishlist %s", itemId.String(), wishlistId.String()))
			s.writeError(w, http.StatusNotFound, "Wishlist or item not found, or not owned by user")
			return
		}
		s.logger.LogError(&userID, "update_item", err, fmt.Sprintf("failed to update item %s in wishlist %s", itemId.String(), wishlistId.String()))
		s.writeError(w, http.StatusInternalServerError, "Failed to update wishlist item")
		return
	}

	s.logger.LogSuccess(&userID, "update_item", fmt.Sprintf("updated item %s in wishlist %s", itemId.String(), wishlistId.String()))
	s.writeJSON(w, http.StatusOK, item)
}

// Remove an item from a wishlist (owner only)
func (s *WishlistServer) DeleteWishlistsWishlistIdItemsItemId(w http.ResponseWriter, r *http.Request, wishlistId openapi_types.UUID, itemId openapi_types.UUID) {
	s.logger.LogRequest(r, nil, "delete_item")

	userID, err := s.extractUserID(r)
	if err != nil {
		s.writeError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	err = s.repo.DeleteWishlistItem(r.Context(), wishlistId, itemId, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.logger.LogNotFound(&userID, "item", fmt.Sprintf("%s in wishlist %s", itemId.String(), wishlistId.String()))
			s.writeError(w, http.StatusNotFound, "Wishlist or item not found, or not owned by user")
			return
		}
		s.logger.LogError(&userID, "delete_item", err, fmt.Sprintf("failed to delete item %s from wishlist %s", itemId.String(), wishlistId.String()))
		s.writeError(w, http.StatusInternalServerError, "Failed to delete wishlist item")
		return
	}

	s.logger.LogSuccess(&userID, "delete_item", fmt.Sprintf("deleted item %s from wishlist %s", itemId.String(), wishlistId.String()))
	w.WriteHeader(http.StatusNoContent)
}

// Update a wishlist (owner only)
func (s *WishlistServer) PutWishlistsWishlistId(w http.ResponseWriter, r *http.Request, wishlistId openapi_types.UUID) {
	s.logger.LogRequest(r, nil, "update_wishlist")

	userID, err := s.extractUserID(r)
	if err != nil {
		s.writeError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	var req wishlistgen.UpdateWishlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.LogBadRequest(&userID, "update_wishlist", fmt.Sprintf("malformed JSON for wishlist %s: %v", wishlistId.String(), err))
		s.writeError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	// Validate request
	if validationErrors := ValidateUpdateWishlistRequest(req); len(validationErrors) > 0 {
		s.logger.LogValidationError(&userID, "update_wishlist", validationErrors)
		s.writeValidationErrors(w, validationErrors)
		return
	}

	// Get existing wishlist to update
	existing, err := s.repo.GetWishlistByID(r.Context(), wishlistId)
	if err != nil {
		s.logger.LogError(&userID, "update_wishlist", err, fmt.Sprintf("failed to retrieve wishlist %s for update", wishlistId.String()))
		s.writeError(w, http.StatusInternalServerError, "Failed to retrieve wishlist")
		return
	}

	if existing == nil {
		s.logger.LogNotFound(&userID, "wishlist", wishlistId.String())
		s.writeError(w, http.StatusNotFound, "Wishlist not found")
		return
	}

	// Check ownership
	if existing.UserId != userID {
		s.logger.LogUnauthorized(r, "update_wishlist", fmt.Sprintf("user %s attempted to update wishlist %s owned by %s", userID.String(), wishlistId.String(), existing.UserId.String()))
		s.writeError(w, http.StatusUnauthorized, "Not authorized to update this wishlist")
		return
	}

	// Update in database
	title := existing.Title
	if req.Title != nil {
		title = *req.Title
	}

	err = s.repo.UpdateWishlist(r.Context(), wishlistId, userID, title, req.Description)
	if err != nil {
		s.logger.LogError(&userID, "update_wishlist", err, fmt.Sprintf("failed to update wishlist %s in database", wishlistId.String()))
		s.writeError(w, http.StatusInternalServerError, "Failed to update wishlist")
		return
	}

	// Get updated wishlist from database
	updated, err := s.repo.GetWishlistByID(r.Context(), wishlistId)
	if err != nil {
		s.logger.LogError(&userID, "update_wishlist", err, fmt.Sprintf("failed to retrieve updated wishlist %s", wishlistId.String()))
		s.writeError(w, http.StatusInternalServerError, "Failed to retrieve updated wishlist")
		return
	}

	s.logger.LogSuccess(&userID, "update_wishlist", fmt.Sprintf("successfully updated wishlist %s", wishlistId.String()))
	s.writeJSON(w, http.StatusOK, updated)
}
