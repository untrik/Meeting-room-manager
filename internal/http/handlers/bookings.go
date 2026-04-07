package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/avito-internships/test-backend-1-untrik/internal/dto"
	"github.com/avito-internships/test-backend-1-untrik/internal/http/helper"
	"github.com/avito-internships/test-backend-1-untrik/internal/http/middleware"
	auth "github.com/avito-internships/test-backend-1-untrik/internal/jwt"
	"github.com/avito-internships/test-backend-1-untrik/internal/service"
)

type bookingServiceInterface interface {
	CreateBooking(ctx context.Context, slotId, userId string) (dto.BookingResponse, error)
	GetMyBooking(ctx context.Context, userId string) ([]dto.BookingResponse, error)
	CancelBooking(ctx context.Context, bookingId, userId string) (dto.BookingResponse, error)
	GetBookingsList(ctx context.Context, pageStr, pageSizeStr string) ([]dto.BookingResponse, dto.Pagination, error)
}

type bokingHandlers struct {
	bookingService bookingServiceInterface
}

func NewBookingHandler(bookingService bookingServiceInterface) *bokingHandlers {
	return &bokingHandlers{bookingService: bookingService}
}
func (bk *bokingHandlers) CreateBooking(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsContextKey).(*auth.Claims)
	if !ok {
		helper.WriteError(w, http.StatusUnauthorized, dto.ErrorCodeUnauthorized, "unauthorized")
		return
	}
	type request struct {
		SlotId string `json:"slotId"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid request")
		slog.Info("Error encode request", "error", err)
		return
	}
	bookingResponse, err := bk.bookingService.CreateBooking(r.Context(), req.SlotId, claims.UserId)
	if err != nil {
		if errors.Is(err, service.ErrInvalidData) {
			slog.Info("Error invalid data", "error", err)
			helper.WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid request")
			return
		}
		if errors.Is(err, service.ErrSlotNotFoundOrPast) {
			slog.Info("slot not found or in past", "error", err)
			helper.WriteError(w, http.StatusNotFound, dto.ErrorCodeSlotNotFound, "slot not found")
			return
		}
		if errors.Is(err, service.ErrBookingExists) {
			slog.Info("slot already booked", "error", err)
			helper.WriteError(w, http.StatusNotFound, dto.ErrorCodeSlotAlreadyBooked, "slot already booked")
			return
		}
		helper.WriteInternalError(w)
		slog.Error("Internal server error", "error", err)
		return
	}
	helper.WriteJSON(w, http.StatusCreated, map[string]any{"booking": bookingResponse})
}
func (bk *bokingHandlers) GetMyBooking(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsContextKey).(*auth.Claims)
	if !ok {
		helper.WriteError(w, http.StatusUnauthorized, dto.ErrorCodeUnauthorized, "unauthorized")
		return
	}
	bookingsResponse, err := bk.bookingService.GetMyBooking(r.Context(), claims.UserId)
	if err != nil {
		helper.WriteInternalError(w)
		slog.Error("Internal server error", "error", err)
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"bookings": bookingsResponse})
}
func (bk *bokingHandlers) CancelBooking(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsContextKey).(*auth.Claims)
	if !ok {
		helper.WriteError(w, http.StatusUnauthorized, dto.ErrorCodeUnauthorized, "unauthorized")
		return
	}
	bookingId := r.PathValue("bookingId")
	bookingResponse, err := bk.bookingService.CancelBooking(r.Context(), bookingId, claims.UserId)
	if err != nil {
		if errors.Is(err, service.ErrInvalidData) {
			slog.Info("Error invalid data", "error", err)
			helper.WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid request")
			return
		}
		if errors.Is(err, service.ErrBookingNotFound) {
			slog.Info("booking not found", "error", err)
			helper.WriteError(w, http.StatusNotFound, dto.ErrorCodeBookingNotFound, "booking not found")
			return
		}
		if errors.Is(err, service.ErrAccess) {
			slog.Info("cancel other booking", "error", err)
			helper.WriteError(w, http.StatusForbidden, dto.ErrorCodeForbidden, "cannot cancel another user's booking")
			return
		}
		helper.WriteInternalError(w)
		slog.Error("Internal server error", "error", err)
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"booking": bookingResponse})
}
func (bk *bokingHandlers) GetBookingsList(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")
	bookingsResponse, pagination, err := bk.bookingService.GetBookingsList(r.Context(), pageStr, pageSizeStr)
	if err != nil {
		if errors.Is(err, service.ErrInvalidData) {
			slog.Info("Error invalid data", "error", err)
			helper.WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid request")
		}
		helper.WriteInternalError(w)
		slog.Error("Internal server error", "error", err)
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"bookings": bookingsResponse, "pagination": pagination})
}
