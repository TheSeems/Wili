package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
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

type telegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
}

func newServer(r UserRepo) *server { return &server{repo: r} }

var jwtKey = []byte(strings.TrimSpace(os.Getenv("JWT_SECRET")))

func sign(userID string) (string, int64) {
	if len(jwtKey) == 0 {
		log.Fatalf("Missing required env: JWT_SECRET")
	}
	exp := time.Hour * 24 * 365 * 10
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

func (s *server) PostAuthYandex(w http.ResponseWriter, r *http.Request) {
	log.Printf("[AUTH] Yandex auth request from %s", r.RemoteAddr)

	type yandexAuthRequest struct {
		Code        string `json:"code"`
		RedirectURI string `json:"redirectUri"`
	}

	var req yandexAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[AUTH] Failed to decode request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Code == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accessToken, err := s.exchangeYandexCode(req.Code, req.RedirectURI)
	if err != nil {
		log.Printf("[AUTH] Failed to exchange Yandex code: %v", err)
		if strings.Contains(err.Error(), "missing Yandex OAuth credentials") {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
		return
	}

	yandexUser, err := s.getYandexUserInfo(accessToken)
	if err != nil {
		log.Printf("[AUTH] Failed to get Yandex user info: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var u *usergen.User
	log.Printf("[AUTH] Checking for existing user with email: %s", yandexUser.DefaultEmail)
	existingUser, err := s.repo.GetByEmail(r.Context(), yandexUser.DefaultEmail)
	if err != nil {
		log.Printf("[AUTH] Error checking for existing user with email %s: %v", yandexUser.DefaultEmail, err)
	} else {
		log.Printf("[AUTH] GetByEmail successful, existingUser is nil: %v", existingUser == nil)
	}
	if err == nil && existingUser != nil {
		log.Printf("[AUTH] Existing user found with email %s, ID: %s", yandexUser.DefaultEmail, existingUser.Id.String())
		u = existingUser

		if yandexUser.DisplayName != "" && yandexUser.DisplayName != u.DisplayName {
			u.DisplayName = yandexUser.DisplayName
		}

		if u.Email == nil || *u.Email != types.Email(yandexUser.DefaultEmail) {
			newEmail := types.Email(yandexUser.DefaultEmail)
			u.Email = &newEmail
		}

		if !yandexUser.IsAvatarEmpty && yandexUser.DefaultAvatarID != "" {
			avatarURL := fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", yandexUser.DefaultAvatarID)
			u.AvatarUrl = &avatarURL
		}
	} else {
		log.Printf("[AUTH] Creating new user with email %s", yandexUser.DefaultEmail)
		newEmail := types.Email(yandexUser.DefaultEmail)
		newUser := usergen.User{
			Id:          uuid.New(),
			DisplayName: yandexUser.DisplayName,
			Email:       &newEmail,
		}

		if !yandexUser.IsAvatarEmpty && yandexUser.DefaultAvatarID != "" {
			avatarURL := fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", yandexUser.DefaultAvatarID)
			newUser.AvatarUrl = &avatarURL
		}

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

func (s *server) PostAuthTelegram(w http.ResponseWriter, r *http.Request) {
	log.Printf("[AUTH] Telegram auth request from %s", r.RemoteAddr)

	type telegramAuthRequest struct {
		InitData string `json:"initData"`
	}

	var req telegramAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[AUTH] Failed to decode telegram auth request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.InitData) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	botToken := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if botToken == "" {
		log.Printf("[AUTH] Telegram auth failed: TELEGRAM_BOT_TOKEN is missing")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	params, err := validateTelegramInitData(req.InitData, botToken, 24*time.Hour)
	if err != nil {
		log.Printf("[AUTH] Telegram initData validation failed: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userRaw := params.Get("user")
	if userRaw == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var tu telegramUser
	if err := json.Unmarshal([]byte(userRaw), &tu); err != nil || tu.ID == 0 {
		log.Printf("[AUTH] Telegram user parse failed: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	displayName := strings.TrimSpace(strings.Join([]string{tu.FirstName, tu.LastName}, " "))
	if displayName == "" && tu.Username != "" {
		displayName = "@" + tu.Username
	}
	if displayName == "" {
		displayName = fmt.Sprintf("tg:%d", tu.ID)
	}

	var avatarURL *string
	if strings.TrimSpace(tu.PhotoURL) != "" {
		u := strings.TrimSpace(tu.PhotoURL)
		avatarURL = &u
	}

	existingUser, err := s.repo.GetByTelegramID(r.Context(), tu.ID)
	if err != nil && err != ErrNotFound {
		log.Printf("[AUTH] Telegram GetByTelegramID error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var u *usergen.User
	if existingUser != nil {
		u = existingUser
		u.DisplayName = displayName
		u.AvatarUrl = avatarURL
	} else {
		u = &usergen.User{
			Id:          uuid.New(),
			DisplayName: displayName,
			AvatarUrl:   avatarURL,
		}
	}

	if err := s.repo.UpsertWithTelegramID(r.Context(), u, tu.ID); err != nil {
		log.Printf("[AUTH] Telegram UpsertWithTelegramID failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tok, exp := sign(u.Id.String())
	resp := usergen.AuthResponse{AccessToken: tok, ExpiresIn: exp, User: *u}
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

func validateTelegramInitData(initData string, botToken string, maxAge time.Duration) (url.Values, error) {
	params, err := url.ParseQuery(initData)
	if err != nil {
		return nil, err
	}

	hash := params.Get("hash")
	if hash == "" {
		return nil, fmt.Errorf("missing hash")
	}

	if maxAge > 0 {
		authDateStr := params.Get("auth_date")
		if authDateStr == "" {
			return nil, fmt.Errorf("missing auth_date")
		}
		sec, err := strconv.ParseInt(authDateStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad auth_date")
		}
		authTime := time.Unix(sec, 0)
		if time.Since(authTime) > maxAge {
			return nil, fmt.Errorf("auth_date expired")
		}
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "hash" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(params.Get(k))
	}
	dataCheckString := b.String()

	secretMac := hmac.New(sha256.New, []byte("WebAppData"))
	secretMac.Write([]byte(botToken))
	secretKey := secretMac.Sum(nil)

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))
	sum := mac.Sum(nil)

	expected := hex.EncodeToString(sum)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(hash)) != 1 {
		return nil, fmt.Errorf("bad signature")
	}
	return params, nil
}

// PostAuthValidate validates a JWT token and returns user information (internal endpoint)
func (s *server) PostAuthValidate(w http.ResponseWriter, r *http.Request) {
	log.Printf("[VALIDATE] Token validation request from %s", r.RemoteAddr)

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
func (s *server) exchangeYandexCode(code string, redirectURIFromClient string) (string, error) {
	clientID := os.Getenv("YANDEX_CLIENT_ID")
	clientSecret := os.Getenv("YANDEX_CLIENT_SECRET")
	redirectURI := redirectURIFromClient
	if redirectURI == "" {
		redirectURI = os.Getenv("YANDEX_REDIRECT_URI")
	}

	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("missing Yandex OAuth credentials (YANDEX_CLIENT_ID / YANDEX_CLIENT_SECRET)")
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	if redirectURI != "" {
		data.Set("redirect_uri", redirectURI)
	}

	resp, err := http.PostForm("https://oauth.yandex.ru/token", data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("[AUTH] Token exchange failed: %d - %s", resp.StatusCode, string(bodyBytes))
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
