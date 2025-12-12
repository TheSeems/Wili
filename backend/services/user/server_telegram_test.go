package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func signTelegramInitData(botToken string, params url.Values) string {
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

	secret := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secret[:])
	mac.Write([]byte(dataCheckString))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestValidateTelegramInitData_OK(t *testing.T) {
	botToken := "123:ABC"
	p := url.Values{}
	p.Set("auth_date", strconv.FormatInt(time.Now().Unix(), 10))
	p.Set("query_id", "AAE")
	p.Set("user", `{"id":123456,"first_name":"Test"}`)
	p.Set("hash", signTelegramInitData(botToken, p))

	_, err := validateTelegramInitData(p.Encode(), botToken, 24*time.Hour)
	if err != nil {
		t.Fatalf("expected ok, got err=%v", err)
	}
}

func TestValidateTelegramInitData_BadSignature(t *testing.T) {
	botToken := "123:ABC"
	p := url.Values{}
	p.Set("auth_date", strconv.FormatInt(time.Now().Unix(), 10))
	p.Set("query_id", "AAE")
	p.Set("user", `{"id":123456,"first_name":"Test"}`)
	p.Set("hash", "deadbeef")

	_, err := validateTelegramInitData(p.Encode(), botToken, 24*time.Hour)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestValidateTelegramInitData_Expired(t *testing.T) {
	botToken := "123:ABC"
	p := url.Values{}
	p.Set("auth_date", strconv.FormatInt(time.Now().Add(-48*time.Hour).Unix(), 10))
	p.Set("query_id", "AAE")
	p.Set("user", `{"id":123456,"first_name":"Test"}`)
	p.Set("hash", signTelegramInitData(botToken, p))

	_, err := validateTelegramInitData(p.Encode(), botToken, 24*time.Hour)
	if err == nil {
		t.Fatalf("expected error")
	}
}
