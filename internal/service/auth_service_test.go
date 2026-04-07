package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-untrik/internal/dto"
	"github.com/avito-internships/test-backend-1-untrik/internal/models"
	"github.com/avito-internships/test-backend-1-untrik/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type fakeUserRepo struct {
	createUserFn      func(ctx context.Context, user models.User, passwordHash []byte) (time.Time, error)
	getUserByEmailFn  func(ctx context.Context, email string) (models.User, error)
	getUserByRoleFn   func(ctx context.Context, role string) (models.User, error)
	getUserRoleFn     func(ctx context.Context, userId string) (string, error)
	getRoleIdByCodeFn func(ctx context.Context, roleCode string) (int, error)
}

func (f *fakeUserRepo) CreateUser(ctx context.Context, user models.User, passwordHash []byte) (time.Time, error) {
	return f.createUserFn(ctx, user, passwordHash)
}

func (f *fakeUserRepo) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	return f.getUserByEmailFn(ctx, email)
}

func (f *fakeUserRepo) GetUserByRole(ctx context.Context, role string) (models.User, error) {
	return f.getUserByRoleFn(ctx, role)
}

func (f *fakeUserRepo) GetUserRole(ctx context.Context, userId string) (string, error) {
	return f.getUserRoleFn(ctx, userId)
}

func (f *fakeUserRepo) GetRoleIdByCode(ctx context.Context, roleCode string) (int, error) {
	return f.getRoleIdByCodeFn(ctx, roleCode)
}

type fakeJWT struct {
	generateTokenFn func(duration time.Duration, userId string, role string) (string, error)
}

func (f *fakeJWT) GenerateToken(duration time.Duration, userId string, role string) (string, error) {
	return f.generateTokenFn(duration, userId, role)
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid email", func(t *testing.T) {
		repo := &fakeUserRepo{}
		jwt := &fakeJWT{}
		svc := NewAuthService(repo, jwt)

		_, err := svc.Login(ctx, "bad-email", "123")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("empty password", func(t *testing.T) {
		repo := &fakeUserRepo{}
		jwt := &fakeJWT{}
		svc := NewAuthService(repo, jwt)

		_, err := svc.Login(ctx, "test@example.com", "")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		repo := &fakeUserRepo{
			getUserByEmailFn: func(ctx context.Context, email string) (models.User, error) {
				return models.User{}, repository.ErrUserNotFound
			},
		}
		jwt := &fakeJWT{}
		svc := NewAuthService(repo, jwt)

		_, err := svc.Login(ctx, "test@example.com", "123456")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("repository get user unexpected error", func(t *testing.T) {
		repo := &fakeUserRepo{
			getUserByEmailFn: func(ctx context.Context, email string) (models.User, error) {
				return models.User{}, errors.New("db error")
			},
		}
		jwt := &fakeJWT{}
		svc := NewAuthService(repo, jwt)

		_, err := svc.Login(ctx, "test@example.com", "123456")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)

		repo := &fakeUserRepo{
			getUserByEmailFn: func(ctx context.Context, email string) (models.User, error) {
				hashStr := string(hash)
				return models.User{
					ID:           "user-1",
					Email:        email,
					PasswordHash: &hashStr,
				}, nil
			},
		}
		jwt := &fakeJWT{}
		svc := NewAuthService(repo, jwt)

		_, err := svc.Login(ctx, "test@example.com", "wrong-password")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("role not found", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

		repo := &fakeUserRepo{
			getUserByEmailFn: func(ctx context.Context, email string) (models.User, error) {
				hashStr := string(hash)
				return models.User{
					ID:           "user-1",
					Email:        email,
					PasswordHash: &hashStr,
				}, nil
			},
			getUserRoleFn: func(ctx context.Context, userId string) (string, error) {
				return "", repository.ErrRoleNotFound
			},
		}
		jwt := &fakeJWT{}
		svc := NewAuthService(repo, jwt)

		_, err := svc.Login(ctx, "test@example.com", "123456")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("get role unexpected error", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

		repo := &fakeUserRepo{
			getUserByEmailFn: func(ctx context.Context, email string) (models.User, error) {
				hashStr := string(hash)
				return models.User{
					ID:           "user-1",
					Email:        email,
					PasswordHash: &hashStr,
				}, nil
			},
			getUserRoleFn: func(ctx context.Context, userId string) (string, error) {
				return "", errors.New("db error")
			},
		}
		jwt := &fakeJWT{}
		svc := NewAuthService(repo, jwt)

		_, err := svc.Login(ctx, "test@example.com", "123456")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("generate token error", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

		repo := &fakeUserRepo{
			getUserByEmailFn: func(ctx context.Context, email string) (models.User, error) {
				hashStr := string(hash)
				return models.User{
					ID:           "user-1",
					Email:        email,
					PasswordHash: &hashStr,
				}, nil
			},
			getUserRoleFn: func(ctx context.Context, userId string) (string, error) {
				return "admin", nil
			},
		}
		jwt := &fakeJWT{
			generateTokenFn: func(duration time.Duration, userId string, role string) (string, error) {
				return "", errors.New("jwt error")
			},
		}
		svc := NewAuthService(repo, jwt)

		_, err := svc.Login(ctx, "test@example.com", "123456")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

		repo := &fakeUserRepo{
			getUserByEmailFn: func(ctx context.Context, email string) (models.User, error) {
				hashStr := string(hash)
				return models.User{
					ID:           "user-1",
					Email:        email,
					PasswordHash: &hashStr,
				}, nil
			},
			getUserRoleFn: func(ctx context.Context, userId string) (string, error) {
				if userId != "user-1" {
					t.Fatalf("expected userId user-1, got %s", userId)
				}
				return "admin", nil
			},
		}
		jwt := &fakeJWT{
			generateTokenFn: func(duration time.Duration, userId string, role string) (string, error) {
				if duration != 3*time.Hour {
					t.Fatalf("expected duration 3h, got %v", duration)
				}
				if userId != "user-1" {
					t.Fatalf("expected userId user-1, got %s", userId)
				}
				if role != "admin" {
					t.Fatalf("expected role admin, got %s", role)
				}
				return "token-123", nil
			},
		}
		svc := NewAuthService(repo, jwt)

		token, err := svc.Login(ctx, "test@example.com", "123456")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token != "token-123" {
			t.Fatalf("expected token-123, got %s", token)
		}
	})
}

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid email", func(t *testing.T) {
		repo := &fakeUserRepo{}
		jwt := &fakeJWT{}
		svc := NewAuthService(repo, jwt)

		_, err := svc.Register(ctx, "bad-email", "123456", "admin")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("empty password", func(t *testing.T) {
		repo := &fakeUserRepo{}
		jwt := &fakeJWT{}
		svc := NewAuthService(repo, jwt)

		_, err := svc.Register(ctx, "test@example.com", "", "admin")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("role not found", func(t *testing.T) {
		repo := &fakeUserRepo{
			getRoleIdByCodeFn: func(ctx context.Context, roleCode string) (int, error) {
				return 0, repository.ErrRoleNotFound
			},
		}
		jwt := &fakeJWT{}
		svc := NewAuthService(repo, jwt)

		_, err := svc.Register(ctx, "test@example.com", "123456", "admin")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("get role id unexpected error", func(t *testing.T) {
		repo := &fakeUserRepo{
			getRoleIdByCodeFn: func(ctx context.Context, roleCode string) (int, error) {
				return 0, errors.New("db error")
			},
		}
		jwt := &fakeJWT{}
		svc := NewAuthService(repo, jwt)

		_, err := svc.Register(ctx, "test@example.com", "123456", "admin")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("user already exists", func(t *testing.T) {
		repo := &fakeUserRepo{
			getRoleIdByCodeFn: func(ctx context.Context, roleCode string) (int, error) {
				if roleCode != "admin" {
					t.Fatalf("expected lowercased admin, got %s", roleCode)
				}
				return 1, nil
			},
			createUserFn: func(ctx context.Context, user models.User, passwordHash []byte) (time.Time, error) {
				return time.Time{}, repository.ErrUserExists
			},
		}
		jwt := &fakeJWT{}
		svc := NewAuthService(repo, jwt)

		_, err := svc.Register(ctx, "test@example.com", "123456", "Admin")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("create user unexpected error", func(t *testing.T) {
		repo := &fakeUserRepo{
			getRoleIdByCodeFn: func(ctx context.Context, roleCode string) (int, error) {
				return 1, nil
			},
			createUserFn: func(ctx context.Context, user models.User, passwordHash []byte) (time.Time, error) {
				return time.Time{}, errors.New("db error")
			},
		}
		jwt := &fakeJWT{}
		svc := NewAuthService(repo, jwt)

		_, err := svc.Register(ctx, "test@example.com", "123456", "admin")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		createdAt := time.Now()

		repo := &fakeUserRepo{
			getRoleIdByCodeFn: func(ctx context.Context, roleCode string) (int, error) {
				if roleCode != "admin" {
					t.Fatalf("expected lowercased role admin, got %s", roleCode)
				}
				return 1, nil
			},
			createUserFn: func(ctx context.Context, user models.User, passwordHash []byte) (time.Time, error) {
				if user.Email != "test@example.com" {
					t.Fatalf("expected email test@example.com, got %s", user.Email)
				}
				if user.RoleID != 1 {
					t.Fatalf("expected roleID 1, got %d", user.RoleID)
				}
				if len(passwordHash) == 0 {
					t.Fatal("expected non-empty password hash")
				}
				if err := bcrypt.CompareHashAndPassword(passwordHash, []byte("123456")); err != nil {
					t.Fatalf("password hash does not match source password: %v", err)
				}
				return createdAt, nil
			},
		}
		jwt := &fakeJWT{}
		svc := NewAuthService(repo, jwt)

		resp, err := svc.Register(ctx, "test@example.com", "123456", "Admin")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Email != "test@example.com" {
			t.Fatalf("expected email test@example.com, got %s", resp.Email)
		}
		if resp.Role != "Admin" {
			t.Fatalf("expected role Admin, got %s", resp.Role)
		}
		if resp.CreatedAt == nil || !resp.CreatedAt.Equal(createdAt) {
			t.Fatalf("expected createdAt %v, got %v", createdAt, resp.CreatedAt)
		}
		if resp.ID == "" {
			t.Fatal("expected generated ID, got empty string")
		}
	})
}

func TestAuthService_DummyLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("user not found", func(t *testing.T) {
		repo := &fakeUserRepo{
			getUserByRoleFn: func(ctx context.Context, role string) (models.User, error) {
				return models.User{}, repository.ErrUserNotFound
			},
		}
		jwt := &fakeJWT{}
		svc := NewAuthService(repo, jwt)

		_, err := svc.DummyLogin(ctx, "admin")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("get user unexpected error", func(t *testing.T) {
		repo := &fakeUserRepo{
			getUserByRoleFn: func(ctx context.Context, role string) (models.User, error) {
				return models.User{}, errors.New("db error")
			},
		}
		jwt := &fakeJWT{}
		svc := NewAuthService(repo, jwt)

		_, err := svc.DummyLogin(ctx, "admin")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("generate token error", func(t *testing.T) {
		repo := &fakeUserRepo{
			getUserByRoleFn: func(ctx context.Context, role string) (models.User, error) {
				return models.User{ID: "user-1"}, nil
			},
		}
		jwt := &fakeJWT{
			generateTokenFn: func(duration time.Duration, userId string, role string) (string, error) {
				return "", errors.New("jwt error")
			},
		}
		svc := NewAuthService(repo, jwt)

		_, err := svc.DummyLogin(ctx, "admin")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		repo := &fakeUserRepo{
			getUserByRoleFn: func(ctx context.Context, role string) (models.User, error) {
				if role != "admin" {
					t.Fatalf("expected role admin, got %s", role)
				}
				return models.User{ID: "user-1"}, nil
			},
		}
		jwt := &fakeJWT{
			generateTokenFn: func(duration time.Duration, userId string, role string) (string, error) {
				if duration != 3*time.Hour {
					t.Fatalf("expected duration 3h, got %v", duration)
				}
				if userId != "user-1" {
					t.Fatalf("expected userId user-1, got %s", userId)
				}
				if role != "admin" {
					t.Fatalf("expected role admin, got %s", role)
				}
				return "dummy-token", nil
			},
		}
		svc := NewAuthService(repo, jwt)

		token, err := svc.DummyLogin(ctx, "admin")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token != "dummy-token" {
			t.Fatalf("expected dummy-token, got %s", token)
		}
	})
}

func TestAuthService_Register_ResponseShape(t *testing.T) {
	ctx := context.Background()
	createdAt := time.Now()

	repo := &fakeUserRepo{
		getRoleIdByCodeFn: func(ctx context.Context, roleCode string) (int, error) {
			return 2, nil
		},
		createUserFn: func(ctx context.Context, user models.User, passwordHash []byte) (time.Time, error) {
			return createdAt, nil
		},
	}
	jwt := &fakeJWT{}
	svc := NewAuthService(repo, jwt)

	resp, err := svc.Register(ctx, "user@example.com", "qwerty", "manager")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == (dto.UserResponse{}) {
		t.Fatal("expected non-empty response")
	}
	if resp.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if resp.Email != "user@example.com" {
		t.Fatalf("expected email user@example.com, got %s", resp.Email)
	}
	if resp.Role != "manager" {
		t.Fatalf("expected role manager, got %s", resp.Role)
	}
	if resp.CreatedAt == nil || !resp.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected createdAt %v, got %v", createdAt, resp.CreatedAt)
	}
}
