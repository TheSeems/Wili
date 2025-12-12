package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
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
	Message     *telegramMessage     `json:"message,omitempty"`
	InlineQuery *telegramInlineQuery `json:"inline_query,omitempty"`
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

type telegramInlineQuery struct {
	ID    string       `json:"id"`
	From  telegramUser `json:"from"`
	Query string       `json:"query"`
}

type telegramUser struct {
	ID int64 `json:"id"`
}

type inlineKeyboardButton struct {
	Text                         string      `json:"text"`
	WebApp                       *webAppInfo `json:"web_app,omitempty"`
	URL                          string      `json:"url,omitempty"`
	SwitchInlineQuery            string      `json:"switch_inline_query,omitempty"`
	SwitchInlineQueryCurrentChat string      `json:"switch_inline_query_current_chat,omitempty"`
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
	ParseMode   string                `json:"parse_mode,omitempty"`
	ReplyMarkup *inlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type answerInlineQueryRequest struct {
	InlineQueryID string        `json:"inline_query_id"`
	Results       []interface{} `json:"results"`
	CacheTime     int           `json:"cache_time,omitempty"`
	IsPersonal    bool          `json:"is_personal,omitempty"`
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
		bindAddr:     envOrDefault("BIND_ADDR", ":8080"),
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

	if upd.InlineQuery != nil {
		if err := b.handleInlineQuery(ctx, upd.InlineQuery); err != nil {
			log.Printf("inline query failed: id=%s err=%v", upd.InlineQuery.ID, err)
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	if upd.Message == nil {
		log.Printf("webhook ignored: no message or inline query")
		w.WriteHeader(http.StatusOK)
		return
	}

	startParam := extractStartParam(upd.Message)
	if startParam == "" {
		log.Printf("webhook ignored: no start param (chat=%d)", upd.Message.Chat.ID)
		w.WriteHeader(http.StatusOK)
		return
	}

	shareListID := parseShareListID(startParam)
	if shareListID != "" {
		log.Printf("webhook share: chat=%d start_param=%s list=%s", upd.Message.Chat.ID, startParam, shareListID)
		if err := b.sendSharePrompt(ctx, upd.Message.Chat.ID, shareListID); err != nil {
			log.Printf("send share prompt failed: chat=%d list=%s err=%v", upd.Message.Chat.ID, shareListID, err)
		}
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

func (b *bot) handleInlineQuery(ctx context.Context, q *telegramInlineQuery) error {
	if q == nil {
		return nil
	}

	listID := parseInlineQueryListID(q.Query)
	if listID == "" {
		help := map[string]interface{}{
			"type":        "article",
			"id":          "help",
			"title":       "Как поделиться вишлистом",
			"description": "Формат: wishlist:<uuid>",
			"input_message_content": map[string]interface{}{
				"message_text": "Введите запрос в формате: wishlist:<uuid>",
			},
		}
		return b.answerInlineQuery(ctx, q.ID, []interface{}{help})
	}

	wl, err := b.fetchWishlist(ctx, listID)
	if err != nil {
		log.Printf("inline query wishlist fetch failed: list=%s err=%v", listID, err)
		errCard := map[string]interface{}{
			"type":        "article",
			"id":          "error",
			"title":       "Не удалось загрузить вишлист",
			"description": "Проверьте id и попробуйте снова",
			"input_message_content": map[string]interface{}{
				"message_text": "Не удалось загрузить вишлист. Проверьте id и попробуйте снова.",
			},
		}
		return b.answerInlineQuery(ctx, q.ID, []interface{}{errCard})
	}

	totalItems := len(wl.Items)
	status := fmt.Sprintf("Всего предметов: %d", totalItems)

	webAppURL := fmt.Sprintf("%s?start=list_%s", b.cfg.webAppURL, listID)
	fallbackURL := fmt.Sprintf("%s/wishlists/%s", b.cfg.webFallback, listID)

	description := "Посмотрите список подарков и забронируйте то, что хотите подарить."
	messageText := fmt.Sprintf("*«%s»*\n%s\n\n%s\n\nМожно открыть по кнопке ниже или в [web](%s)", wl.Title, status, description, fallbackURL)

	result := map[string]interface{}{
		"type":        "article",
		"id":          fmt.Sprintf("wishlist_%s", listID),
		"title":       wl.Title,
		"description": status,
		"input_message_content": map[string]interface{}{
			"message_text": messageText,
			"parse_mode":   "Markdown",
		},
		"reply_markup": inlineKeyboardMarkup{
			InlineKeyboard: [][]inlineKeyboardButton{
				{
					{Text: "Открыть вишлист", URL: webAppURL},
				},
			},
		},
	}

	return b.answerInlineQuery(ctx, q.ID, []interface{}{result})
}

func (b *bot) answerInlineQuery(ctx context.Context, inlineQueryID string, results []interface{}) error {
	if results == nil {
		results = make([]interface{}, 0)
	}
	reqBody := answerInlineQueryRequest{
		InlineQueryID: inlineQueryID,
		Results:       results,
		CacheTime:     1,
		IsPersonal:    true,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://api.telegram.org/bot%s/answerInlineQuery", b.cfg.botToken), strings.NewReader(string(body)))
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
		bb, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram answerInlineQuery status %d body=%s", resp.StatusCode, strings.TrimSpace(string(bb)))
	}
	return nil
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
	uuidRe := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !uuidRe.MatchString(id) {
		return ""
	}
	return id
}

func parseShareListID(param string) string {
	if !strings.HasPrefix(param, "share_") {
		return ""
	}
	id := strings.TrimPrefix(param, "share_")
	return parseListID("list_" + id)
}

func parseInlineQueryListID(query string) string {
	q := strings.TrimSpace(query)
	if q == "" {
		return ""
	}
	if strings.HasPrefix(q, "wishlist:") {
		return parseListID("list_" + strings.TrimPrefix(q, "wishlist:"))
	}
	if strings.HasPrefix(q, "list_") {
		return parseListID(q)
	}
	return ""
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

	description := "Посмотрите список подарков и забронируйте то, что хотите подарить. Чтобы увидеть, что уже забронировано, откройте вишлист."
	text := fmt.Sprintf("*«%s»*\n%s\n\n%s\n\nМожно посмотреть по кнопке ниже или в [web](%s)", wl.Title, status, description, fallbackURL)

	msg := sendMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
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

func (b *bot) sendSharePrompt(ctx context.Context, chatID int64, listID string) error {
	wl, err := b.fetchWishlist(ctx, listID)
	if err != nil {
		return err
	}

	webAppURL := fmt.Sprintf("%s?start=list_%s", b.cfg.webAppURL, listID)
	fallbackURL := fmt.Sprintf("%s/wishlists/%s", b.cfg.webFallback, listID)

	text := fmt.Sprintf("*«%s»*\n\nНажмите кнопку ниже, выберите чат и отправьте сообщение с кнопкой.\n\nЕсли не работает — можно открыть в [web](%s).", wl.Title, fallbackURL)
	msg := sendMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
		ReplyMarkup: &inlineKeyboardMarkup{
			InlineKeyboard: [][]inlineKeyboardButton{
				{
					{Text: "Отправить в чат…", SwitchInlineQuery: fmt.Sprintf("wishlist:%s", listID)},
				},
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
