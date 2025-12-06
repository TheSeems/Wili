//go:build dev
// +build dev

package devutil

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

var allowedOrigins = []string{"http://localhost:5173", "http://127.0.0.1:5173"}

// EnableCORS applies permissive CORS for local development.
func EnableCORS(r chi.Router) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}))
}
