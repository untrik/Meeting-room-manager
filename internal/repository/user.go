package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/avito-internships/test-backend-1-untrik/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrUserExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")
var ErrRoleNotFound = errors.New("role not found")

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *userRepository {
	return &userRepository{db: db}
}
func (u *userRepository) CreateUser(ctx context.Context, user models.User, passwordHash []byte) (time.Time, error) {
	query := `
		INSERT INTO users (id,email,role_id,password_hash,created_at) VALUES ($1,$2,$3,$4,NOW())
		RETURNING created_at
	`
	var createdAt time.Time
	err := u.db.QueryRowContext(ctx, query, user.ID, strings.ToLower(user.Email), user.RoleID, string(passwordHash)).Scan(&createdAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return createdAt, ErrUserExists
		}
		return createdAt, err
	}
	return createdAt, nil
}
func (u *userRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	query := `
		SELECT id, email, role_id, password_hash, created_at FROM users WHERE email = $1
	`
	var user models.User
	if err := u.db.QueryRowContext(ctx, query, strings.ToLower(email)).Scan(&user.ID,
		&user.Email, &user.RoleID, &user.PasswordHash, &user.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	return user, nil
}
func (u *userRepository) GetUserByRole(ctx context.Context, role string) (models.User, error) {
	query := `
		SELECT users.id, users.email, users.role_id, users.password_hash, users.created_at FROM users 
		JOIN roles ON users.role_id = roles.id
		WHERE users.id IN (
		'11111111-1111-1111-1111-111111111111',
		'22222222-2222-2222-2222-222222222222')
		AND (roles.code = $1)
	`
	var user models.User
	if err := u.db.QueryRowContext(ctx, query, strings.ToLower(role)).Scan(&user.ID,
		&user.Email, &user.RoleID, &user.PasswordHash, &user.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	return user, nil
}
func (u *userRepository) GetUserRole(ctx context.Context, userId string) (string, error) {
	query := `
		SELECT roles.code FROM roles
		JOIN users ON users.role_id = roles.id
		WHERE users.id = $1
	`
	var role string
	if err := u.db.QueryRowContext(ctx, query, userId).Scan(&role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrRoleNotFound
		}
		return "", err
	}
	return role, nil
}
func (u *userRepository) GetRoleIdByCode(ctx context.Context, roleCode string) (int, error) {
	query := `
		SELECT id FROM roles WHERE code = $1
	`
	var roleId int
	if err := u.db.QueryRowContext(ctx, query, roleCode).Scan(&roleId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrRoleNotFound
		}
		return 0, err
	}
	return roleId, nil
}
