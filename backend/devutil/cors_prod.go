//go:build !dev
// +build !dev

package devutil

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

// EnableCORS applies production CORS settings for wili.me domain.
func EnableCORS(r chi.Router) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://wili.me", "https://www.wili.me", "https://oauth.yandex.ru"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	}))
}
