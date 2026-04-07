package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/avito-internships/test-backend-1-untrik/internal/dto"
	"github.com/avito-internships/test-backend-1-untrik/internal/http/helper"
	"github.com/avito-internships/test-backend-1-untrik/internal/service"
)

type AuthServiceInterface interface {
	Login(ctx context.Context, email, password string) (string, error)
	Register(ctx context.Context, email, password, role string) (dto.UserResponse, error)
	DummyLogin(ctx context.Context, role string) (string, error)
}

type authHandler struct {
	authService AuthServiceInterface
}

func NewAuthHandler(authService AuthServiceInterface) *authHandler {
	return &authHandler{authService: authService}
}

func (ah *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid request")
		slog.Info("Error encode request", "error", err)
		return
	}
	token, err := ah.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			helper.WriteError(w, http.StatusUnauthorized, dto.ErrorCodeInvalidRequest, "invalid request")
			slog.Info("invalid request", "reason", err)
			return
		}
		helper.WriteInternalError(w)
		slog.Error("Internal server error", "error", err)
		return
	}
	helper.WriteJSON(w, http.StatusOK, dto.Token{Token: token})
}
func (ah *authHandler) Register(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid request")
		slog.Info("Error encode request", "error", err)
		return
	}
	userResponse, err := ah.authService.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			helper.WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid request")
			slog.Info("invalid request", "reason", err)
			return
		}
		helper.WriteInternalError(w)
		slog.Error("Internal server error", "error", err)
		return
	}
	helper.WriteJSON(w, http.StatusCreated, map[string]any{
		"user": userResponse})
}

func (ah *authHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Role string `json:"role"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid request")
		slog.Info("Error encode request", "error", err)
		return
	}
	token, err := ah.authService.DummyLogin(r.Context(), req.Role)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			helper.WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid request")
			slog.Info("invalid request", "reason", err)
			return
		}
		helper.WriteInternalError(w)
		slog.Error("Internal server error", "error", err)
		return
	}
	helper.WriteJSON(w, http.StatusOK, dto.Token{Token: token})
}
