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
	log.Printf("Starting Wili User Service (build tag dev=%v)...", isDevBuild())

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	r := chi.NewRouter()
	devutil.EnableCORS(r)

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

	devutil.MountSwagger(r, "Wili User Service API")

	port := getEnv("PORT", "8080")
	addr := ":" + port
	log.Printf("User service listening on %s (cors=%s)", addr, corsProfile())
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

func isDevBuild() bool {
	// build tags are not directly detectable; rely on presence of dev CORS origins
	return len(devutil.AllowedOrigins()) > 0 && devutil.AllowedOrigins()[0] == "http://localhost:5173"
}

func corsProfile() string {
	if isDevBuild() {
		return "dev"
	}
	return "prod"
}
