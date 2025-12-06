package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type wishlistResponse struct {
	Title string         `json:"title"`
	Items []wishlistItem `json:"items"`
}

type wishlistItem struct {
	Booking *itemBooking `json:"booking"`
}

type itemBooking struct {
	BookerName *string `json:"bookerName"`
}

type telegramUpdate struct {
	Message *telegramMessage `json:"message"`
}

type telegramMessage struct {
	MessageID int64            `json:"message_id"`
	Chat      telegramChat     `json:"chat"`
	Text      string           `json:"text"`
	Entities  []telegramEntity `json:"entities"`
}

type telegramChat struct {
	ID int64 `json:"id"`
}

type telegramEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

type inlineKeyboardButton struct {
	Text   string      `json:"text"`
	WebApp *webAppInfo `json:"web_app,omitempty"`
	URL    string      `json:"url,omitempty"`
}

type webAppInfo struct {
	URL string `json:"url"`
}

type inlineKeyboardMarkup struct {
	InlineKeyboard [][]inlineKeyboardButton `json:"inline_keyboard"`
}

type sendMessageRequest struct {
	ChatID      int64                 `json:"chat_id"`
	Text        string                `json:"text"`
	ReplyMarkup *inlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type config struct {
	botToken     string
	apiBaseURL   string
	webAppURL    string
	webFallback  string
	bindAddr     string
	webhookPath  string
	webhookToken string
}

func mustEnv(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		log.Fatalf("Missing required env: %s", key)
	}
	return v
}

func loadConfig() config {
	return config{
		botToken:     mustEnv("TELEGRAM_BOT_TOKEN"),
		apiBaseURL:   strings.TrimRight(mustEnv("WISHLIST_API_BASE_URL"), "/"),
		webAppURL:    strings.TrimRight(mustEnv("TELEGRAM_WEBAPP_URL"), "/"),
		webFallback:  strings.TrimRight(mustEnv("WISHES_WEB_URL"), "/"),
		bindAddr:     envOrDefault("BIND_ADDR", ":8090"),
		webhookPath:  envOrDefault("WEBHOOK_PATH", "webhook"),
		webhookToken: envOrDefault("WEBHOOK_SECRET_TOKEN", ""),
	}
}

func envOrDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

type bot struct {
	cfg    config
	client *http.Client
}

func newBot(cfg config) *bot {
	return &bot{
		cfg: cfg,
		client: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

func (b *bot) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if b.cfg.webhookToken != "" {
		if got := r.Header.Get("X-Telegram-Bot-Api-Secret-Token"); got != b.cfg.webhookToken {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	ctx := r.Context()
	var upd telegramUpdate
	if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		log.Printf("webhook decode failed: %v", err)
		return
	}

	if upd.Message == nil {
		log.Printf("webhook ignored: no message")
		w.WriteHeader(http.StatusOK)
		return
	}

	startParam := extractStartParam(upd.Message)
	if startParam == "" {
		log.Printf("webhook ignored: no start param (chat=%d)", upd.Message.Chat.ID)
		w.WriteHeader(http.StatusOK)
		return
	}

	listID := parseListID(startParam)
	if listID == "" {
		log.Printf("webhook ignored: bad start param=%s (chat=%d)", startParam, upd.Message.Chat.ID)
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("webhook start: chat=%d start_param=%s list=%s", upd.Message.Chat.ID, startParam, listID)

	if err := b.sendWishlistPreview(ctx, upd.Message.Chat.ID, listID); err != nil {
		log.Printf("send preview failed: chat=%d list=%s err=%v", upd.Message.Chat.ID, listID, err)
	}
	w.WriteHeader(http.StatusOK)
}

func extractStartParam(msg *telegramMessage) string {
	if msg == nil || msg.Text == "" {
		return ""
	}

	if !strings.HasPrefix(msg.Text, "/start") {
		return ""
	}

	parts := strings.SplitN(msg.Text, " ", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func parseListID(param string) string {
	if !strings.HasPrefix(param, "list_") {
		return ""
	}
	id := strings.TrimPrefix(param, "list_")
	if len(id) < 10 {
		return ""
	}
	return id
}

func (b *bot) fetchWishlist(ctx context.Context, listID string) (*wishlistResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s", b.cfg.apiBaseURL, listID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wishlist api status %d", resp.StatusCode)
	}

	var wl wishlistResponse
	if err := json.NewDecoder(resp.Body).Decode(&wl); err != nil {
		return nil, err
	}
	return &wl, nil
}

func (b *bot) sendWishlistPreview(ctx context.Context, chatID int64, listID string) error {
	wl, err := b.fetchWishlist(ctx, listID)
	if err != nil {
		return err
	}

	totalItems := len(wl.Items)
	status := fmt.Sprintf("Всего предметов: %d", totalItems)

	webAppURL := fmt.Sprintf("%s?start=list_%s", b.cfg.webAppURL, listID)
	fallbackURL := fmt.Sprintf("%s/wishlists/%s", b.cfg.webFallback, listID)

	description := "Это вишлист Wili. Смотрите позиции и бронируйте подарки. Чтобы увидеть актуальные брони и отметить предмет, откройте мини‑приложение."
	text := fmt.Sprintf("«%s»\n%s\n\n%s\n\nОткрыть в браузере: %s", wl.Title, status, description, fallbackURL)

	msg := sendMessageRequest{
		ChatID: chatID,
		Text:   text,
		ReplyMarkup: &inlineKeyboardMarkup{
			InlineKeyboard: [][]inlineKeyboardButton{
				{
					{Text: "Открыть вишлист", WebApp: &webAppInfo{URL: webAppURL}},
				},
			},
		},
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.cfg.botToken), strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram sendMessage status %d", resp.StatusCode)
	}
	log.Printf("preview sent: chat=%d list=%s status=%d items=%d", chatID, listID, resp.StatusCode, totalItems)
	return nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env not found: %v", err)
	}

	cfg := loadConfig()
	b := newBot(cfg)
	log.Printf("telegram-bot starting bind=%s webhook=/%s api=%s webapp=%s fallback=%s", cfg.bindAddr, cfg.webhookPath, cfg.apiBaseURL, cfg.webAppURL, cfg.webFallback)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/"+strings.TrimPrefix(cfg.webhookPath, "/"), b.handleWebhook)

	log.Printf("telegram-bot listening on %s", cfg.bindAddr)
	if err := http.ListenAndServe(cfg.bindAddr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
