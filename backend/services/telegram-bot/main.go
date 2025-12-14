package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type wishlistResponse struct {
	Title       string         `json:"title"`
	Description *string        `json:"description"`
	Items       []wishlistItem `json:"items"`
}

type wishlistListResponse struct {
	Wishlists []struct {
		ID          string  `json:"id"`
		Title       string  `json:"title"`
		Description *string `json:"description"`
	} `json:"wishlists"`
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
	From      telegramUser     `json:"from"`
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
	ID           int64  `json:"id"`
	LanguageCode string `json:"language_code,omitempty"`
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
	ChatID                int64                 `json:"chat_id"`
	Text                  string                `json:"text"`
	ParseMode             string                `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool                  `json:"disable_web_page_preview,omitempty"`
	ReplyMarkup           *inlineKeyboardMarkup `json:"reply_markup,omitempty"`
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
	userAPIBase  string
	webAppURL    string
	webFallback  string
	frontendURL  string
	miniAppBot   string
	miniAppName  string
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
	webFallback := strings.TrimRight(mustEnv("WISHES_WEB_URL"), "/")
	frontendURL := envOrDefault("FRONTEND_URL", strings.TrimSuffix(webFallback, "/wishlists"))
	return config{
		botToken:     mustEnv("TELEGRAM_BOT_TOKEN"),
		apiBaseURL:   strings.TrimRight(mustEnv("WISHLIST_API_BASE_URL"), "/"),
		userAPIBase:  strings.TrimRight(envOrDefault("USER_API_BASE_URL", "https://api.wili.me"), "/"),
		webAppURL:    strings.TrimRight(mustEnv("TELEGRAM_WEBAPP_URL"), "/"),
		webFallback:  webFallback,
		frontendURL:  frontendURL,
		miniAppBot:   strings.TrimSpace(os.Getenv("TELEGRAM_MINIAPP_BOT_USERNAME")),
		miniAppName:  strings.TrimSpace(os.Getenv("TELEGRAM_MINIAPP_NAME")),
		bindAddr:     envOrDefault("BIND_ADDR", ":8080"),
		webhookPath:  envOrDefault("WEBHOOK_PATH", "webhook"),
		webhookToken: envOrDefault("WEBHOOK_SECRET_TOKEN", ""),
	}
}

func (b *bot) miniAppDeepLink(listID string) string {
	if b.cfg.miniAppBot != "" {
		base := fmt.Sprintf("https://t.me/%s", b.cfg.miniAppBot)
		if b.cfg.miniAppName != "" {
			base = fmt.Sprintf("%s/%s", base, b.cfg.miniAppName)
		}
		return fmt.Sprintf("%s?startapp=list_%s", base, listID)
	}
	return fmt.Sprintf("%s?start=list_%s", b.cfg.webAppURL, listID)
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

func normalizeLang(code string) string {
	c := strings.ToLower(strings.TrimSpace(code))
	if strings.HasPrefix(c, "en") {
		return "en"
	}
	return "ru"
}

const (
	keyMiniAppEntryText    = "miniapp.entry.text"
	keyMiniAppEntryButton  = "miniapp.entry.button"
	keySharePromptText     = "share.prompt.text"
	keySharePromptSendChat = "share.prompt.send_chat"
	keySharePromptOpen     = "share.prompt.open"
	keyPreviewBaseDesc     = "preview.base_desc"
	keyPreviewText         = "preview.text"
	keyPreviewOpen         = "preview.open"
	keyInlineHelpTitle     = "inline.help.title"
	keyInlineHelpDesc      = "inline.help.desc"
	keyInlineHelpMessage   = "inline.help.message"
	keyInlineErrorTitle    = "inline.error.title"
	keyInlineErrorDesc     = "inline.error.desc"
	keyInlineErrorMessage  = "inline.error.message"
	keyInlineBaseDesc      = "inline.base_desc"
	keyInlineMessage       = "inline.message"
	keyInlineOpenDesc      = "inline.open.desc"
	keyInlineOpenButton    = "inline.open.button"
	keyWebAuthText         = "webauth.text"
	keyWebAuthButton       = "webauth.button"
	keyWebAuthNotLinked    = "webauth.not_linked"
	keyWebAuthOpenMiniApp  = "webauth.open_miniapp"
)

var botDict = map[string]map[string]string{
	"ru": {
		keyMiniAppEntryText:    "Откройте Wili в Telegram Mini App.",
		keyMiniAppEntryButton:  "Открыть Wili",
		keySharePromptText:     "<b>«%s»</b>\n\nНажмите кнопку ниже, выберите чат и отправьте сообщение с кнопкой.\n\nЕсли не работает — можно открыть в <a href=\"%s\">web</a>.",
		keySharePromptSendChat: "Отправить в чат…",
		keySharePromptOpen:     "Открыть вишлист",
		keyPreviewBaseDesc:     "Посмотрите список желаний и забронируйте то, что хотите подарить. Чтобы увидеть, что уже забронировано, откройте вишлист.",
		keyPreviewText:         "<b>«%s»</b>\n\n%s\n\nМожно посмотреть по кнопке ниже или в <a href=\"%s\">web</a>",
		keyPreviewOpen:         "Открыть вишлист",
		keyInlineHelpTitle:     "Как поделиться вишлистом",
		keyInlineHelpDesc:      "Формат: wishlist:<uuid>",
		keyInlineHelpMessage:   "Введите запрос в формате: wishlist:<uuid>",
		"inline.my.title":      "Мои вишлисты",
		"inline.my.desc":       "Выберите вишлист и отправьте его в чат",
		"inline.my.notLinked":  "Аккаунт не привязан. Откройте Mini App и войдите через Telegram.",
		"inline.my.empty":      "У вас пока нет вишлистов. Создайте первый в Mini App.",
		keyInlineErrorTitle:    "Не удалось загрузить вишлист",
		keyInlineErrorDesc:     "Проверьте id и попробуйте снова",
		keyInlineErrorMessage:  "Не удалось загрузить вишлист. Проверьте id и попробуйте снова.",
		keyInlineBaseDesc:      "Посмотрите список подарков и забронируйте то, что хотите подарить.",
		keyInlineMessage:       "<b>«%s»</b>\n\n%s\n\nЕсли не работает кнопка, можно открыть в <a href=\"%s\">web</a>",
		keyInlineOpenDesc:      "Открыть вишлист",
		keyInlineOpenButton:    "Открыть вишлист",
		keyWebAuthText:         "Нажмите кнопку ниже, чтобы войти в веб-версию Wili.\n\nКстати, вы можете пользоваться Wili прямо здесь, в Telegram!",
		keyWebAuthButton:       "Войти на сайт",
		keyWebAuthNotLinked:    "Чтобы войти на сайт через Telegram, сначала откройте Mini App и создайте аккаунт.",
		keyWebAuthOpenMiniApp:  "Открыть Wili",
	},
	"en": {
		keyMiniAppEntryText:    "Open Wili in Telegram Mini App.",
		keyMiniAppEntryButton:  "Open Wili",
		keySharePromptText:     "<b>«%s»</b>\n\nTap the button below, choose a chat and send the message with a button.\n\nIf it doesn't work — open in <a href=\"%s\">web</a>.",
		keySharePromptSendChat: "Send to chat…",
		keySharePromptOpen:     "Open wishlist",
		keyPreviewBaseDesc:     "View the wish list and book what you want to give. To see what's already booked, open the wishlist.",
		keyPreviewText:         "<b>«%s»</b>\n\n%s\n\nOpen with the button below or in <a href=\"%s\">web</a>",
		keyPreviewOpen:         "Open wishlist",
		keyInlineHelpTitle:     "How to share a wishlist",
		keyInlineHelpDesc:      "Format: wishlist:<uuid>",
		keyInlineHelpMessage:   "Type a query in the format: wishlist:<uuid>",
		"inline.my.title":      "My wishlists",
		"inline.my.desc":       "Pick a wishlist and send it to chat",
		"inline.my.notLinked":  "Account is not linked. Open the Mini App and log in with Telegram.",
		"inline.my.empty":      "You don’t have any wishlists yet. Create one in the Mini App.",
		keyInlineErrorTitle:    "Couldn't load wishlist",
		keyInlineErrorDesc:     "Check the id and try again",
		keyInlineErrorMessage:  "Couldn't load wishlist. Check the id and try again.",
		keyInlineBaseDesc:      "View the gift list and book what you want to give.",
		keyInlineMessage:       "<b>«%s»</b>\n\n%s\n\nIf the button doesn't work, open in <a href=\"%s\">web</a>",
		keyInlineOpenDesc:      "Open wishlist",
		keyInlineOpenButton:    "Open wishlist",
		keyWebAuthText:         "Tap the button below to log in to Wili web.\n\nBy the way, you can use Wili right here in Telegram!",
		keyWebAuthButton:       "Log in to website",
		keyWebAuthNotLinked:    "To log in via Telegram, first open the Mini App and create an account.",
		keyWebAuthOpenMiniApp:  "Open Wili",
	},
}

func tr(lang, key string) string {
	l := normalizeLang(lang)
	if m, ok := botDict[l]; ok {
		if v, ok := m[key]; ok && strings.TrimSpace(v) != "" {
			return v
		}
	}
	if m, ok := botDict["ru"]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return key
}

func trf(lang, key string, args ...any) string {
	return fmt.Sprintf(tr(lang, key), args...)
}

func (b *bot) miniAppBotLink() string {
	if b.cfg.miniAppBot != "" {
		return fmt.Sprintf("https://t.me/%s", b.cfg.miniAppBot)
	}
	return b.cfg.webAppURL
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
		if strings.HasPrefix(strings.TrimSpace(upd.Message.Text), "/start") {
			log.Printf("webhook start: chat=%d start_param=<empty>", upd.Message.Chat.ID)
			if err := b.sendMiniAppEntry(ctx, upd.Message.Chat.ID, upd.Message.From.LanguageCode); err != nil {
				log.Printf("send miniapp entry failed: chat=%d err=%v", upd.Message.Chat.ID, err)
			}
		} else {
			log.Printf("webhook ignored: no start param (chat=%d)", upd.Message.Chat.ID)
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	shareListID := parseShareListID(startParam)
	if shareListID != "" {
		log.Printf("webhook share: chat=%d start_param=%s list=%s", upd.Message.Chat.ID, startParam, shareListID)
		if err := b.sendSharePrompt(ctx, upd.Message.Chat.ID, shareListID, upd.Message.From.LanguageCode); err != nil {
			log.Printf("send share prompt failed: chat=%d list=%s err=%v", upd.Message.Chat.ID, shareListID, err)
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	if state, ok := parseWebAuth(startParam); ok {
		log.Printf("webhook webauth: chat=%d state=%s", upd.Message.Chat.ID, state)
		if err := b.sendWebAuth(ctx, upd.Message.Chat.ID, upd.Message.From.ID, state, upd.Message.From.LanguageCode); err != nil {
			log.Printf("send webauth failed: chat=%d err=%v", upd.Message.Chat.ID, err)
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

	if err := b.sendWishlistPreview(ctx, upd.Message.Chat.ID, listID, upd.Message.From.LanguageCode); err != nil {
		log.Printf("send preview failed: chat=%d list=%s err=%v", upd.Message.Chat.ID, listID, err)
	}
	w.WriteHeader(http.StatusOK)
}

func (b *bot) sendMiniAppEntry(ctx context.Context, chatID int64, lang string) error {
	text := tr(lang, keyMiniAppEntryText)
	msg := sendMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
		ReplyMarkup: &inlineKeyboardMarkup{
			InlineKeyboard: [][]inlineKeyboardButton{
				{
					{Text: tr(lang, keyMiniAppEntryButton), WebApp: &webAppInfo{URL: b.cfg.webAppURL}},
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

func esc(s string) string {
	return html.EscapeString(s)
}

func (b *bot) fetchTelegramJWT(ctx context.Context, telegramID int64) (string, error) {
	body := fmt.Sprintf(`{"telegramId":%d}`, telegramID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/auth/telegram-bot", b.cfg.userAPIBase), strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Wili-Bot-Token", b.cfg.botToken)

	resp, err := b.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("not linked")
	}
	if resp.StatusCode != http.StatusOK {
		bb, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("user-service auth status %d body=%s", resp.StatusCode, strings.TrimSpace(string(bb)))
	}

	var out struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", fmt.Errorf("empty token")
	}
	return out.AccessToken, nil
}

func (b *bot) fetchMyWishlists(ctx context.Context, jwt string) (*wishlistListResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.cfg.apiBaseURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wishlist api status %d", resp.StatusCode)
	}
	var wl wishlistListResponse
	if err := json.NewDecoder(resp.Body).Decode(&wl); err != nil {
		return nil, err
	}
	return &wl, nil
}

func (b *bot) handleInlineQuery(ctx context.Context, q *telegramInlineQuery) error {
	if q == nil {
		return nil
	}

	log.Printf("inline query: from=%d id=%s q=%q", q.From.ID, q.ID, q.Query)
	lang := q.From.LanguageCode

	queryText := strings.TrimSpace(q.Query)
	listID := parseInlineQueryListID(queryText)
	if listID == "" && queryText == "" {
		jwt, err := b.fetchTelegramJWT(ctx, q.From.ID)
		if err != nil {
			msg := tr(lang, "inline.my.notLinked")
			card := map[string]interface{}{
				"type":        "article",
				"id":          "my_not_linked",
				"title":       tr(lang, "inline.my.title"),
				"description": msg,
				"input_message_content": map[string]interface{}{
					"message_text": msg,
				},
				"reply_markup": inlineKeyboardMarkup{
					InlineKeyboard: [][]inlineKeyboardButton{
						{
							{Text: tr(lang, keyMiniAppEntryButton), URL: b.miniAppBotLink()},
						},
					},
				},
			}
			return b.answerInlineQuery(ctx, q.ID, []interface{}{card})
		}

		lists, err := b.fetchMyWishlists(ctx, jwt)
		if err != nil {
			log.Printf("inline query my wishlists failed: from=%d err=%v", q.From.ID, err)
			return b.answerInlineQuery(ctx, q.ID, []interface{}{})
		}
		if lists == nil || len(lists.Wishlists) == 0 {
			msg := tr(lang, "inline.my.empty")
			card := map[string]interface{}{
				"type":        "article",
				"id":          "my_empty",
				"title":       tr(lang, "inline.my.title"),
				"description": msg,
				"input_message_content": map[string]interface{}{
					"message_text": msg,
				},
				"reply_markup": inlineKeyboardMarkup{
					InlineKeyboard: [][]inlineKeyboardButton{
						{
							{Text: tr(lang, keyMiniAppEntryButton), URL: b.miniAppBotLink()},
						},
					},
				},
			}
			return b.answerInlineQuery(ctx, q.ID, []interface{}{card})
		}

		results := make([]interface{}, 0, len(lists.Wishlists))
		for i, wl := range lists.Wishlists {
			if i >= 10 {
				break
			}
			webAppURL := b.miniAppDeepLink(wl.ID)
			fallbackURL := fmt.Sprintf("%s/wishlists/%s", b.cfg.webFallback, wl.ID)

			desc := tr(lang, keyInlineBaseDesc)
			if wl.Description != nil && strings.TrimSpace(*wl.Description) != "" {
				desc = fmt.Sprintf("%s\n\n%s", strings.TrimSpace(*wl.Description), desc)
			}
			messageText := trf(lang, keyInlineMessage, esc(wl.Title), esc(desc), esc(fallbackURL))
			result := map[string]interface{}{
				"type":        "article",
				"id":          fmt.Sprintf("my_%s", wl.ID),
				"title":       wl.Title,
				"description": tr(lang, "inline.my.desc"),
				"input_message_content": map[string]interface{}{
					"message_text":             messageText,
					"parse_mode":               "HTML",
					"disable_web_page_preview": true,
				},
				"reply_markup": inlineKeyboardMarkup{
					InlineKeyboard: [][]inlineKeyboardButton{
						{
							{Text: tr(lang, keyInlineOpenButton), URL: webAppURL},
						},
					},
				},
			}
			results = append(results, result)
		}
		return b.answerInlineQuery(ctx, q.ID, results)
	}

	if listID == "" {
		help := map[string]interface{}{
			"type":        "article",
			"id":          "help",
			"title":       tr(lang, keyInlineHelpTitle),
			"description": tr(lang, keyInlineHelpDesc),
			"input_message_content": map[string]interface{}{
				"message_text": tr(lang, keyInlineHelpMessage),
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
			"title":       tr(lang, keyInlineErrorTitle),
			"description": tr(lang, keyInlineErrorDesc),
			"input_message_content": map[string]interface{}{
				"message_text": tr(lang, keyInlineErrorMessage),
			},
		}
		return b.answerInlineQuery(ctx, q.ID, []interface{}{errCard})
	}

	webAppURL := b.miniAppDeepLink(listID)
	fallbackURL := fmt.Sprintf("%s/wishlists/%s", b.cfg.webFallback, listID)
	log.Printf("inline query resolved: id=%s list=%s url=%s", q.ID, listID, webAppURL)

	baseDescription := tr(lang, keyInlineBaseDesc)
	wlDesc := strings.TrimSpace(func() string {
		if wl.Description == nil {
			return ""
		}
		return *wl.Description
	}())
	if wlDesc != "" {
		baseDescription = fmt.Sprintf("%s\n\n%s", wlDesc, baseDescription)
	}
	messageText := trf(lang, keyInlineMessage, esc(wl.Title), esc(baseDescription), esc(fallbackURL))

	result := map[string]interface{}{
		"type":        "article",
		"id":          fmt.Sprintf("wishlist_%s", listID),
		"title":       wl.Title,
		"description": tr(lang, keyInlineOpenDesc),
		"input_message_content": map[string]interface{}{
			"message_text":             messageText,
			"parse_mode":               "HTML",
			"disable_web_page_preview": true,
		},
		"reply_markup": inlineKeyboardMarkup{
			InlineKeyboard: [][]inlineKeyboardButton{
				{
					{Text: tr(lang, keyInlineOpenButton), URL: webAppURL},
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
	log.Printf("inline query answered: id=%s results=%d", inlineQueryID, len(results))
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

func parseWebAuth(param string) (state string, ok bool) {
	if !strings.HasPrefix(param, "webauth_") {
		return "", false
	}
	state = strings.TrimPrefix(param, "webauth_")
	if len(state) < 8 || len(state) > 32 {
		return "", false
	}
	return state, true
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

func (b *bot) sendWishlistPreview(ctx context.Context, chatID int64, listID string, lang string) error {
	wl, err := b.fetchWishlist(ctx, listID)
	if err != nil {
		return err
	}

	webAppURL := fmt.Sprintf("%s?start=list_%s", b.cfg.webAppURL, listID)
	fallbackURL := fmt.Sprintf("%s/wishlists/%s", b.cfg.webFallback, listID)

	baseDescription := tr(lang, keyPreviewBaseDesc)
	wlDesc := strings.TrimSpace(func() string {
		if wl.Description == nil {
			return ""
		}
		return *wl.Description
	}())
	if wlDesc != "" {
		baseDescription = fmt.Sprintf("%s\n\n%s", wlDesc, baseDescription)
	}
	text := trf(lang, keyPreviewText, esc(wl.Title), esc(baseDescription), esc(fallbackURL))

	msg := sendMessageRequest{
		ChatID:                chatID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
		ReplyMarkup: &inlineKeyboardMarkup{
			InlineKeyboard: [][]inlineKeyboardButton{
				{
					{Text: tr(lang, keyPreviewOpen), WebApp: &webAppInfo{URL: webAppURL}},
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
	log.Printf("preview sent: chat=%d list=%s status=%d", chatID, listID, resp.StatusCode)
	return nil
}

func (b *bot) sendSharePrompt(ctx context.Context, chatID int64, listID string, lang string) error {
	wl, err := b.fetchWishlist(ctx, listID)
	if err != nil {
		return err
	}

	webAppURL := fmt.Sprintf("%s?start=list_%s", b.cfg.webAppURL, listID)
	fallbackURL := fmt.Sprintf("%s/wishlists/%s", b.cfg.webFallback, listID)

	text := trf(lang, keySharePromptText, esc(wl.Title), esc(fallbackURL))
	msg := sendMessageRequest{
		ChatID:                chatID,
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
		ReplyMarkup: &inlineKeyboardMarkup{
			InlineKeyboard: [][]inlineKeyboardButton{
				{
					{Text: tr(lang, keySharePromptSendChat), SwitchInlineQuery: fmt.Sprintf("wishlist:%s", listID)},
				},
				{
					{Text: tr(lang, keySharePromptOpen), WebApp: &webAppInfo{URL: webAppURL}},
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

func (b *bot) sendWebAuth(ctx context.Context, chatID int64, telegramID int64, state string, lang string) error {
	jwt, err := b.fetchTelegramJWT(ctx, telegramID)
	if err != nil {
		log.Printf("webauth fetch JWT failed: telegram_id=%d err=%v", telegramID, err)
		text := tr(lang, keyWebAuthNotLinked)
		msg := sendMessageRequest{
			ChatID:    chatID,
			Text:      text,
			ParseMode: "HTML",
			ReplyMarkup: &inlineKeyboardMarkup{
				InlineKeyboard: [][]inlineKeyboardButton{
					{
						{Text: tr(lang, keyWebAuthOpenMiniApp), WebApp: &webAppInfo{URL: b.cfg.webAppURL}},
					},
				},
			},
		}
		body, _ := json.Marshal(msg)
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.cfg.botToken), strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := b.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return nil
	}

	loginURL := fmt.Sprintf("%s/auth/telegram-callback?token=%s&state=%s", b.cfg.frontendURL, url.QueryEscape(jwt), url.QueryEscape(state))

	text := tr(lang, keyWebAuthText)
	msg := sendMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
		ReplyMarkup: &inlineKeyboardMarkup{
			InlineKeyboard: [][]inlineKeyboardButton{
				{
					{Text: tr(lang, keyWebAuthButton), URL: loginURL},
				},
				{
					{Text: tr(lang, keyWebAuthOpenMiniApp), WebApp: &webAppInfo{URL: b.cfg.webAppURL}},
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
		bb, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram sendMessage status %d body=%s", resp.StatusCode, strings.TrimSpace(string(bb)))
	}
	log.Printf("webauth sent: chat=%d state=%s", chatID, state)
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
