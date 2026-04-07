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

type schedulerServiceInterface interface {
	CreateSchedule(ctx context.Context, roomId string, daysOfWeek []int,
		startTime, endTime string) (dto.ScheduleResponse, error)
}

type schedulerHandlers struct {
	schedulerService schedulerServiceInterface
}

func NewSchedulerHandler(schedulerService schedulerServiceInterface) *schedulerHandlers {
	return &schedulerHandlers{schedulerService: schedulerService}
}

func (sh *schedulerHandlers) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	roomId := r.PathValue("roomId")
	type request struct {
		DaysOfWeek []int  `json:"daysOfWeek"`
		StartTime  string `json:"startTime"`
		EndTime    string `json:"endTime"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid request")
		slog.Info("Error encode request", "error", err)
		return
	}
	scheduleResponse, err := sh.schedulerService.CreateSchedule(r.Context(), roomId, req.DaysOfWeek, req.StartTime, req.EndTime)
	if err != nil {
		if errors.Is(err, service.ErrInvalidData) {
			helper.WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid request")
			return
		}
		if errors.Is(err, service.ErrRoomNotFound) {
			helper.WriteError(w, http.StatusNotFound, dto.ErrorCodeRoomNotFound, "room not found")
			return
		}
		if errors.Is(err, service.ErrScheduleExists) {
			helper.WriteError(w, http.StatusConflict, dto.ErrorCodeScheduleExists, "The schedule cannot be edited")
			return
		}
		helper.WriteInternalError(w)
		return
	}
	helper.WriteJSON(w, http.StatusCreated, map[string]any{"schedule": scheduleResponse})
}
