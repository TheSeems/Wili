# Wili Backend

Go microservices powering authentication and wishlists.

## Services

- user: Yandex ID auth, user profiles, JWT issuance
- wishlist: CRUD for wishlists and items

## Tech

- Go 1.24.x, chi
- Postgres (user), MongoDB (wishlist)
- OpenAPI (oapi-codegen)

## Dev

Prereqs: Go, Postgres, MongoDB, env vars.

Run services:
```
cd backend/services/user && go run .
cd backend/services/wishlist && go run .
```

Spec-first: update `openapi.yaml`, then regenerate clients/servers as needed.
