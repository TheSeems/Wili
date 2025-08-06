# Wili Backend

## Stack

- Golang 1.24.5 for modern web development
- Postgres for efficiency in user service
- MongoDB for extensibility in wishlist service as we're about to support super-custom wishlists with products from marketplaces and multiple other reward types
- OpenAPI
- Kubernetes

## Architecture

Microservice architecture

## Approach

- Spec-first development. First, create a spec, then generate (or regenerate) client or server endpoints and use it. Same spec, client for frontend, server endpoints for backend
