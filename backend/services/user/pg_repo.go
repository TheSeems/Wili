package main

import (
	"context"
	"os"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/oapi-codegen/runtime/types"
	usergen "github.com/theseems/wili/backend/services/user/gen"
)

type pgRepo struct {
	db *sqlx.DB
}

// dbUser is a struct that matches the database schema exactly
type dbUser struct {
	ID          uuid.UUID    `db:"id"`
	DisplayName string       `db:"display_name"`
	AvatarUrl   *string      `db:"avatar_url"`
	Email       *types.Email `db:"email"`
	TelegramID  *int64       `db:"telegram_id"`
}

// toUsergen converts dbUser to usergen.User
func (du *dbUser) toUsergen() *usergen.User {
	return &usergen.User{
		Id:          du.ID,
		DisplayName: du.DisplayName,
		AvatarUrl:   du.AvatarUrl,
		Email:       du.Email,
	}
}

func newPGRepo() (*pgRepo, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("DATABASE_URL is not set")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	repo := &pgRepo{db: db}
	if err := repo.migrate(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (p *pgRepo) migrate() error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		display_name TEXT NOT NULL,
		avatar_url TEXT,
		email TEXT UNIQUE,
		telegram_id BIGINT
	);`
	if _, err := p.db.Exec(query); err != nil {
		return err
	}
	if _, err := p.db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS telegram_id BIGINT`); err != nil {
		return err
	}
	_, err := p.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS users_telegram_id_uq ON users (telegram_id) WHERE telegram_id IS NOT NULL`)
	return err
}

func (p *pgRepo) Upsert(ctx context.Context, u *usergen.User) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO users (id, display_name, avatar_url)
		VALUES ($1,$2,$3)
		ON CONFLICT (id) DO UPDATE SET display_name=EXCLUDED.display_name, avatar_url=EXCLUDED.avatar_url`,
		u.Id, u.DisplayName, u.AvatarUrl)
	return err
}

func (p *pgRepo) UpsertWithEmail(ctx context.Context, u *usergen.User, email string) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO users (id, display_name, avatar_url, email)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (email) DO UPDATE SET 
			display_name=EXCLUDED.display_name, avatar_url=EXCLUDED.avatar_url`,
		u.Id, u.DisplayName, u.AvatarUrl, email)
	return err
}

func (p *pgRepo) UpsertWithTelegramID(ctx context.Context, u *usergen.User, telegramID int64) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO users (id, display_name, avatar_url, telegram_id)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (telegram_id) WHERE telegram_id IS NOT NULL DO UPDATE SET
			display_name=EXCLUDED.display_name, avatar_url=EXCLUDED.avatar_url`,
		u.Id, u.DisplayName, u.AvatarUrl, telegramID)
	return err
}

func (p *pgRepo) Get(ctx context.Context, id uuid.UUID) (*usergen.User, error) {
	var u dbUser
	err := p.db.GetContext(ctx, &u, `SELECT id, display_name, avatar_url, email, telegram_id FROM users WHERE id=$1`, id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u.toUsergen(), nil
}

func (p *pgRepo) GetByEmail(ctx context.Context, email string) (*usergen.User, error) {
	var u dbUser
	err := p.db.GetContext(ctx, &u, `SELECT id, display_name, avatar_url, email, telegram_id FROM users WHERE email=$1`, email)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u.toUsergen(), nil
}

func (p *pgRepo) GetByTelegramID(ctx context.Context, telegramID int64) (*usergen.User, error) {
	var u dbUser
	err := p.db.GetContext(ctx, &u, `SELECT id, display_name, avatar_url, email, telegram_id FROM users WHERE telegram_id=$1`, telegramID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u.toUsergen(), nil
}
