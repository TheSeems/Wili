package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	usergen "github.com/theseems/wili/backend/services/user/gen"
)

type server struct {
	repo UserRepo
}

// Yandex OAuth response structures
type yandexTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type yandexUserInfo struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	DefaultEmail    string `json:"default_email"`
	IsAvatarEmpty   bool   `json:"is_avatar_empty"`
	DefaultAvatarID string `json:"default_avatar_id"`
}

func newServer(r UserRepo) *server { return &server{repo: r} }

// --- helpers
var jwtKey = []byte(os.Getenv("JWT_SIGNING_KEY"))

func sign(userID string) (string, int64) {
	exp := time.Hour * 24
	claims := jwt.MapClaims{"sub": userID, "exp": time.Now().Add(exp).Unix()}
	t, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtKey)
	return t, int64(exp.Seconds())
}

func parseBearer(r *http.Request) (uuid.UUID, bool) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return uuid.UUID{}, false
	}
	tok := strings.TrimPrefix(h, "Bearer ")
	t, err := jwt.Parse(tok, func(t *jwt.Token) (interface{}, error) { return jwtKey, nil })
	if err != nil || !t.Valid {
		return uuid.UUID{}, false
	}
	idStr, ok := t.Claims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		return uuid.UUID{}, false
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.UUID{}, false
	}
	return id, true
}

// ---- handlers -----

func (s *server) PostAuthYandex(w http.ResponseWriter, r *http.Request) {
	log.Printf("[AUTH] Yandex auth request from %s", r.RemoteAddr)

	var req usergen.YandexAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[AUTH] Failed to decode request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Exchange authorization code for access token
	accessToken, err := s.exchangeYandexCode(req.Code)
	if err != nil {
		log.Printf("[AUTH] Failed to exchange Yandex code: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Get user info from Yandex
	yandexUser, err := s.getYandexUserInfo(accessToken)
	if err != nil {
		log.Printf("[AUTH] Failed to get Yandex user info: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if user already exists by email
	var u *usergen.User
	log.Printf("[AUTH] Checking for existing user with email: %s", yandexUser.DefaultEmail)
	existingUser, err := s.repo.GetByEmail(r.Context(), yandexUser.DefaultEmail)
	if err != nil {
		log.Printf("[AUTH] Error checking for existing user with email %s: %v", yandexUser.DefaultEmail, err)
	} else {
		log.Printf("[AUTH] GetByEmail successful, existingUser is nil: %v", existingUser == nil)
	}
	if err == nil && existingUser != nil {
		// User exists, update their info
		log.Printf("[AUTH] Existing user found with email %s, ID: %s", yandexUser.DefaultEmail, existingUser.Id.String())
		u = existingUser

		// Update display name if it changed
		if yandexUser.DisplayName != "" && yandexUser.DisplayName != u.DisplayName {
			u.DisplayName = yandexUser.DisplayName
		}

		// Update email if not set
		if u.Email == nil || *u.Email != types.Email(yandexUser.DefaultEmail) {
			newEmail := types.Email(yandexUser.DefaultEmail)
			u.Email = &newEmail
		}

		// Update avatar URL if available
		if !yandexUser.IsAvatarEmpty && yandexUser.DefaultAvatarID != "" {
			avatarURL := fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", yandexUser.DefaultAvatarID)
			u.AvatarUrl = &avatarURL
		}
	} else {
		// Create new user from Yandex data
		log.Printf("[AUTH] Creating new user with email %s", yandexUser.DefaultEmail)
		newEmail := types.Email(yandexUser.DefaultEmail)
		newUser := usergen.User{
			Id:          uuid.New(),
			DisplayName: yandexUser.DisplayName,
			Email:       &newEmail,
		}

		// Set avatar URL if available
		if !yandexUser.IsAvatarEmpty && yandexUser.DefaultAvatarID != "" {
			avatarURL := fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", yandexUser.DefaultAvatarID)
			newUser.AvatarUrl = &avatarURL
		}

		// Use email as display name if display name is empty
		if newUser.DisplayName == "" {
			if yandexUser.DefaultEmail != "" {
				newUser.DisplayName = yandexUser.DefaultEmail
			} else if yandexUser.Login != "" {
				newUser.DisplayName = yandexUser.Login
			} else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("User info does not contain display name"))
				return
			}
		}

		u = &newUser
	}

	// Store the user with email
	err = s.repo.UpsertWithEmail(r.Context(), u, yandexUser.DefaultEmail)
	if err != nil {
		log.Printf("[AUTH] Failed to save user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to save user"))
		return
	}

	tok, exp := sign(u.Id.String())
	resp := usergen.AuthResponse{AccessToken: tok, ExpiresIn: exp, User: *u}
	log.Printf("[AUTH] Successfully authenticated user: %s (%s) with email: %s", u.DisplayName, u.Id.String(), yandexUser.DefaultEmail)
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *server) GetUsersMe(w http.ResponseWriter, r *http.Request) {
	id, ok := parseBearer(r)
	if !ok {
		log.Printf("[USER] Unauthorized access to /users/me from %s", r.RemoteAddr)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	u, err := s.repo.Get(r.Context(), id)
	if err != nil {
		log.Printf("[USER] Failed to get user %s: %v", id.String(), err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Printf("[USER] Retrieved profile for user: %s", u.DisplayName)
	_ = json.NewEncoder(w).Encode(u)
}

func (s *server) PutUsersMe(w http.ResponseWriter, r *http.Request) {
	id, ok := parseBearer(r)
	if !ok {
		log.Printf("[USER] Unauthorized profile update attempt from %s", r.RemoteAddr)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var upd usergen.UpdateUserRequest
	_ = json.NewDecoder(r.Body).Decode(&upd)
	u, err := s.repo.Get(r.Context(), id)
	if err != nil {
		log.Printf("[USER] Failed to get user %s for update: %v", id.String(), err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if upd.DisplayName != nil {
		u.DisplayName = *upd.DisplayName
	}
	if upd.AvatarUrl != nil {
		u.AvatarUrl = upd.AvatarUrl
	}
	err = s.repo.Upsert(r.Context(), u)
	if err != nil {
		log.Printf("[USER] Failed to update user profile: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to update profile"))
		return
	}
	log.Printf("[USER] Updated profile for user: %s", u.DisplayName)
	_ = json.NewEncoder(w).Encode(u)
}

// PostAuthValidate validates a JWT token and returns user information (internal endpoint)
func (s *server) PostAuthValidate(w http.ResponseWriter, r *http.Request) {
	log.Printf("[VALIDATE] Token validation request from %s", r.RemoteAddr)

	// For now, use a simple struct until we regenerate the types
	type validateReq struct {
		Token string `json:"token"`
	}
	type validateResp struct {
		Valid bool         `json:"valid"`
		User  usergen.User `json:"user"`
	}

	var req validateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[VALIDATE] Failed to decode request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Parse and validate the JWT token
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		log.Printf("[VALIDATE] Invalid token: %v", err)
		json.NewEncoder(w).Encode(validateResp{
			Valid: false,
			User:  usergen.User{}, // Empty user for invalid tokens
		})
		return
	}

	// Extract user ID from token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("[VALIDATE] Invalid token claims")
		json.NewEncoder(w).Encode(validateResp{
			Valid: false,
			User:  usergen.User{},
		})
		return
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		log.Printf("[VALIDATE] Missing user ID in token")
		json.NewEncoder(w).Encode(validateResp{
			Valid: false,
			User:  usergen.User{},
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("[VALIDATE] Invalid user ID format: %v", err)
		json.NewEncoder(w).Encode(validateResp{
			Valid: false,
			User:  usergen.User{},
		})
		return
	}

	// Get user from repository
	user, err := s.repo.Get(r.Context(), userID)
	if err != nil {
		log.Printf("[VALIDATE] Failed to get user %s: %v", userID.String(), err)
		json.NewEncoder(w).Encode(validateResp{
			Valid: false,
			User:  usergen.User{},
		})
		return
	}

	log.Printf("[VALIDATE] Successfully validated token for user: %s", user.DisplayName)
	json.NewEncoder(w).Encode(validateResp{
		Valid: true,
		User:  *user,
	})
}

func (s *server) GetUsersUserId(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	u, err := s.repo.Get(r.Context(), uuid.UUID(userId))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(u)
}

// exchangeYandexCode exchanges authorization code for access token
func (s *server) exchangeYandexCode(code string) (string, error) {
	clientID := os.Getenv("YANDEX_CLIENT_ID")
	clientSecret := os.Getenv("YANDEX_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("missing Yandex OAuth credentials")
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	resp, err := http.PostForm("https://oauth.yandex.ru/token", data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}

	var tokenResp yandexTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

// getYandexUserInfo retrieves user information using access token
func (s *server) getYandexUserInfo(accessToken string) (*yandexUserInfo, error) {
	req, err := http.NewRequest("GET", "https://login.yandex.ru/info", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "OAuth "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	var userInfo yandexUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
