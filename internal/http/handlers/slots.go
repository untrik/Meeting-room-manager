package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/avito-internships/test-backend-1-untrik/internal/dto"
	"github.com/avito-internships/test-backend-1-untrik/internal/http/helper"
	"github.com/avito-internships/test-backend-1-untrik/internal/service"
)

type slotsServiceInterface interface {
	GetAvailableSlots(ctx context.Context, roomId string, date string) ([]dto.SlotResponse, error)
}

type slotsHandlers struct {
	slotsService slotsServiceInterface
}

func NewSlotsHandler(slotsService slotsServiceInterface) *slotsHandlers {
	return &slotsHandlers{slotsService: slotsService}
}
func (sh *slotsHandlers) GetAvailableSlots(w http.ResponseWriter, r *http.Request) {
	roomId := r.PathValue("roomId")
	date := r.URL.Query().Get("date")
	slotsResponse, err := sh.slotsService.GetAvailableSlots(r.Context(), roomId, date)
	if err != nil {
		if errors.Is(err, service.ErrInvalidData) {
			helper.WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid request")
			return
		}
		if errors.Is(err, service.ErrRoomNotFound) {
			helper.WriteError(w, http.StatusNotFound, dto.ErrorCodeRoomNotFound, "room not found")
			return
		}
		if errors.Is(err, service.ErrScheduleNotFound) {
			helper.WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid request")
			return
		}
		helper.WriteInternalError(w)
		return
	}
	helper.WriteJSON(w, http.StatusOK, map[string]any{"slots": slotsResponse})
}
