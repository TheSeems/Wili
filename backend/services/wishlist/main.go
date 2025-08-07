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
	// Initialize logger
	logger := NewLogger("WISHLIST")

	log.Printf("Starting Wili Wishlist Service...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get environment variables with defaults
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

	// Initialize MongoDB repository
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

	// Initialize user service client
	userClient := NewUserClient(userServiceURL)

	// Initialize server
	server := NewWishlistServer(repo, userClient)

	// Setup router
	r := chi.NewRouter()
	devutil.EnableCORS(r)

	// Add health endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "wishlist-service",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Mount API routes
	r.Mount("/", wishlistgen.Handler(server))

	// Mount Swagger UI for development
	devutil.MountSwagger(r, "Wili Wishlist Service API")

	logger.LogStartup(addr, mongoURI+"/"+dbName)
	log.Printf("User Service URL: %s", userServiceURL)
	log.Fatal(http.ListenAndServe(addr, r))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
