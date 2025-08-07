package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/theseems/wili/backend/devutil"

	usergen "github.com/theseems/wili/backend/services/user/gen"
)

func main() {
	log.Printf("Starting Wili User Service...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	r := chi.NewRouter()
	devutil.EnableCORS(r)

	// Add health endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "user-service",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	repo, err := newPGRepo()
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	srv := newServer(repo)
	r.Mount("/", usergen.Handler(srv))

	// optional swagger UI (mounted only with `go run -tags=dev`)
	devutil.MountSwagger(r, "Wili User Service API")

	port := getEnv("PORT", "8080")
	addr := ":" + port
	log.Printf("User service listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
