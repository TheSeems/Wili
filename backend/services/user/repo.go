package main

import (
	"context"
	"errors"

	"github.com/google/uuid"
	usergen "github.com/theseems/wili/backend/services/user/gen"
)

type User = usergen.User

var ErrNotFound = errors.New("not found")

type UserRepo interface {
	Upsert(ctx context.Context, u *User) error
	UpsertWithEmail(ctx context.Context, u *User, email string) error
	Get(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}
