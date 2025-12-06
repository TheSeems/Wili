package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/theseems/wili/backend/devutil"
	wishlistgen "github.com/theseems/wili/backend/services/wishlist/gen"
)

func main() {
	logger := NewLogger("WISHLIST")

	log.Printf("Starting Wili Wishlist Service...")

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	mongoURI := getEnv("MONGODB_URI", "")
	if mongoURI == "" {
		panic("MONGODB_URI is not set")
	}
	dbName := getEnv("DATABASE_NAME", "")
	if dbName == "" {
		panic("DATABASE_NAME is not set")
	}
	userServiceURL := getEnv("USER_SERVICE_URL", "")
	if userServiceURL == "" {
		panic("USER_SERVICE_URL is not set")
	}

	port := getEnv("PORT", "8081")
	addr := ":" + port

	repo, err := NewMongoRepo(mongoURI, dbName)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB repository: %v", err)
	}
	defer func() {
		if err := repo.Close(context.Background()); err != nil {
			logger.LogShutdown("Error closing MongoDB connection: " + err.Error())
		} else {
			logger.LogShutdown("MongoDB connection closed successfully")
		}
	}()

	userClient := NewUserClient(userServiceURL)
	server := NewWishlistServer(repo, userClient)

	r := chi.NewRouter()
	devutil.EnableCORS(r)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "wishlist-service",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	r.Mount("/", wishlistgen.Handler(server))
	devutil.MountSwagger(r, "Wili Wishlist Service API")

	logger.LogStartup(addr, mongoURI+"/"+dbName)
	log.Printf("User Service URL: %s (cors=%s)", userServiceURL, corsProfile())
	log.Fatal(http.ListenAndServe(addr, r))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func corsProfile() string {
	if len(devutil.AllowedOrigins()) > 0 && devutil.AllowedOrigins()[0] == "http://localhost:5173" {
		return "dev"
	}
	return "prod"
}
