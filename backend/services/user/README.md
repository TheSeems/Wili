# User Service

Auth and profiles.

## Features

- Yandex ID login (issues JWT)
- User profile CRUD (name, avatar, email)
- Token validation endpoint for internal services

## Run local

Env: Postgres, `JWT_SECRET`, `YANDEX_CLIENT_ID/SECRET`.

```
go run .
```