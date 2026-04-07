package service

import (
	"context"
	"errors"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"github.com/avito-internships/test-backend-1-untrik/internal/dto"
	"github.com/avito-internships/test-backend-1-untrik/internal/models"
	"github.com/avito-internships/test-backend-1-untrik/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type jwtInterface interface {
	GenerateToken(duration time.Duration, userId string, role string) (string, error)
}
type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, user models.User, passwordHash []byte) (time.Time, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetUserByRole(ctx context.Context, role string) (models.User, error)
	GetUserRole(ctx context.Context, userId string) (string, error)
	GetRoleIdByCode(ctx context.Context, roleCode string) (int, error)
}

type authService struct {
	userRepository UserRepositoryInterface
	jwtService     jwtInterface
}

func NewAuthService(userRepository UserRepositoryInterface, jwtService jwtInterface) *authService {
	return &authService{userRepository: userRepository, jwtService: jwtService}
}
func (us *authService) Login(ctx context.Context, email, password string) (string, error) {
	_, err := mail.ParseAddress(email)
	if err != nil {
		slog.Info("Invalid credentials", "reason", err)
		return "", ErrInvalidCredentials
	}
	if password == "" {
		slog.Info("password in empthy")
		return "", ErrInvalidCredentials
	}
	user, err := us.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			slog.Info("Invalid credentials", "reason", err)
			return "", ErrInvalidCredentials
		}
		slog.Error("failed to get user", "error", err)
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		slog.Info("Invalid credentials", "reason", err)
		return "", ErrInvalidCredentials
	}
	role, err := us.userRepository.GetUserRole(ctx, user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			slog.Info("Invalid credentials", "reason", err)
			return "", ErrInvalidCredentials
		}
		slog.Error("failed to get user role", "error", err)
		return "", err
	}
	slog.Info("user logged successfully")
	token, err := us.jwtService.GenerateToken(3*time.Hour, user.ID, role)
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		return "", err
	}
	return token, nil
}
func (us *authService) Register(ctx context.Context, email, password, role string) (dto.UserResponse, error) {
	_, err := mail.ParseAddress(email)
	if err != nil {
		slog.Info("Invalid credentials", "reason", err)
		return dto.UserResponse{}, ErrInvalidCredentials
	}
	if password == "" {
		slog.Info("password is empthy")
		return dto.UserResponse{}, ErrInvalidCredentials
	}
	roleId, err := us.userRepository.GetRoleIdByCode(ctx, strings.ToLower(role))
	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			slog.Info("Invalid credentials", "reason", err)
			return dto.UserResponse{}, ErrInvalidCredentials
		}
		slog.Error("failed to get roleId", "error", err)
		return dto.UserResponse{}, err
	}
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("failed to generate password hash", "error", err)
		return dto.UserResponse{}, err
	}
	userId := uuid.New()
	user := models.User{ID: userId.String(), Email: email, RoleID: int16(roleId)}
	createdAt, err := us.userRepository.CreateUser(ctx, user, passHash)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			slog.Info("Invalid credentials", "reason", err)
			return dto.UserResponse{}, ErrInvalidCredentials
		}
		slog.Error("failed to create user", "error", err)
		return dto.UserResponse{}, err
	}
	return dto.UserResponse{ID: userId.String(), Email: email, Role: role, CreatedAt: &createdAt}, nil
}
func (us *authService) DummyLogin(ctx context.Context, role string) (string, error) {
	user, err := us.userRepository.GetUserByRole(ctx, role)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			slog.Info("Invalid credentials", "reason", err)
			return "", ErrInvalidCredentials
		}
		slog.Error("failed to get user", "error", err)
		return "", err
	}
	token, err := us.jwtService.GenerateToken(3*time.Hour, user.ID, role)
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		return "", err
	}
	return token, nil
}
