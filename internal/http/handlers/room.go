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

type roomServiceInterface interface {
	GetListRooms(ctx context.Context) ([]dto.RoomResponse, error)
	CreateRoom(ctx context.Context, name string, description *string, capacity *int) (dto.RoomResponse, error)
}

type roomHandlers struct {
	roomService roomServiceInterface
}

func NewRoomHandler(roomService roomServiceInterface) *roomHandlers {
	return &roomHandlers{roomService: roomService}
}
func (rh *roomHandlers) CreateRoom(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Capacity    int    `json:"capacity"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid request")
		slog.Info("Error encode request", "error", err)
		return
	}
	roomResponse, err := rh.roomService.CreateRoom(r.Context(), req.Name, &req.Description, &req.Capacity)
	if err != nil {
		if errors.Is(err, service.ErrInvalidData) {
			helper.WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid request")
			slog.Info("invalid request", "reason", err)
			return
		}
		helper.WriteInternalError(w)
		slog.Error("Internal server error", "error", err)
		return
	}
	helper.WriteJSON(w, http.StatusCreated, map[string]any{"room": roomResponse})
}
func (rh *roomHandlers) GetListRooms(w http.ResponseWriter, r *http.Request) {
	roomResponse, err := rh.roomService.GetListRooms(r.Context())
	if err != nil {
		helper.WriteInternalError(w)
		slog.Error("Internal server error", "error", err)
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"rooms": roomResponse})
}
