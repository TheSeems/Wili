module github.com/theseems/wili/backend/services/wishlist

go 1.24.5

replace github.com/theseems/wili/backend/devutil => ../../devutil

replace github.com/theseems/wili/backend/services/user => ../user

require (
	github.com/go-chi/chi/v5 v5.2.2
	github.com/google/uuid v1.5.0
	github.com/joho/godotenv v1.5.1
	github.com/oapi-codegen/runtime v1.1.2
	github.com/theseems/wili/backend/devutil v0.0.0-00010101000000-000000000000
	go.mongodb.org/mongo-driver v1.17.1
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/go-chi/cors v1.2.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/swaggest/swgui v1.8.4 // indirect
	github.com/vearutop/statigz v1.4.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.17.0 // indirect
)
