package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"user-service/internal/domain"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Insert(ctx context.Context, u *domain.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, email, password_hash, full_name, verified, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		u.ID, u.Email, u.PasswordHash, u.FullName, u.Verified, u.CreatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "users_email_key") {
			return domain.ErrEmailTaken
		}
	}
	return err
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return r.scan(r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, full_name, verified, created_at FROM users WHERE id = $1`, id))
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.scan(r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, full_name, verified, created_at FROM users WHERE email = $1`, email))
}

func (r *UserRepo) UpdateName(ctx context.Context, id, fullName string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE users SET full_name = $2 WHERE id = $1`, id, fullName)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id, hash string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE users SET password_hash = $2 WHERE id = $1`, id, hash)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) MarkVerified(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE users SET verified = TRUE WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) Exists(ctx context.Context, email string) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE email = $1`, email).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *UserRepo) RevokeToken(ctx context.Context, jti string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO revoked_tokens (jti, expires_at) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		jti, expiresAt)
	return err
}

func (r *UserRepo) IsTokenRevoked(ctx context.Context, jti string) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM revoked_tokens WHERE jti = $1`, jti).Scan(&n)
	return n > 0, err
}

func (r *UserRepo) scan(row *sql.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Verified, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
