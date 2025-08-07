package main

import (
	"log"
	"net/http"
	"os"

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
