module github.com/theseems/wili/backend/services/user

go 1.24.5

replace github.com/theseems/wili/backend/devutil => ../../devutil

require (
	github.com/go-chi/chi/v5 v5.2.2
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.5.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/oapi-codegen/runtime v1.1.2
	github.com/theseems/wili/backend/devutil v0.0.0-00010101000000-000000000000
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/go-chi/cors v1.2.2 // indirect
	github.com/swaggest/swgui v1.8.4 // indirect
	github.com/vearutop/statigz v1.4.0 // indirect
)
