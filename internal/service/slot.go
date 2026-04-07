package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/avito-internships/test-backend-1-untrik/internal/dto"
	"github.com/avito-internships/test-backend-1-untrik/internal/models"
	"github.com/avito-internships/test-backend-1-untrik/internal/repository"
	"github.com/google/uuid"
)

type slotsRepositoryInterface interface {
	SlotsBulkCreate(ctx context.Context, slots []models.Slot) error
	GetAvailableByRoomAndDate(ctx context.Context, roomId string, date time.Time) ([]models.Slot, error)
	ExistsForRoomAndDate(ctx context.Context, roomID string, date time.Time) (bool, error)
}

type slotsService struct {
	scheduleGetter  scheduleGetter
	slotsRepository slotsRepositoryInterface
	roomChecker     RoomChecker
}

type scheduleGetter interface {
	GetByRoomID(ctx context.Context, roomID string) (models.Schedule, error)
}

func NewSlotsService(scheduleGetter scheduleGetter,
	slotsRepository slotsRepositoryInterface,
	roomChecker RoomChecker) *slotsService {
	return &slotsService{scheduleGetter: scheduleGetter,
		slotsRepository: slotsRepository,
		roomChecker:     roomChecker}
}
func (ss *slotsService) GetAvailableSlots(ctx context.Context, roomId string, date string) ([]dto.SlotResponse, error) {
	if date == "" {
		slog.Info("date in empthy")
		return nil, ErrInvalidData
	}
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		slog.Info("invalid date format")
		return nil, ErrInvalidData
	}
	if err := uuid.Validate(roomId); err != nil {
		slog.Info("invalid roomId")
		return nil, ErrInvalidData
	}
	exist, err := ss.roomChecker.ExistsRoom(ctx, roomId)
	if err != nil {
		slog.Error("error exist room chek", "error", err)
		return nil, err
	}
	if !exist {
		slog.Info("room not found")
		return nil, ErrRoomNotFound
	}
	schedule, err := ss.scheduleGetter.GetByRoomID(ctx, roomId)
	if err != nil {
		if errors.Is(err, repository.ErrScheduleNotFound) {
			slog.Info("schedule not found")
			return nil, ErrScheduleNotFound
		}
		slog.Error("error get schedule", "error", err)
		return nil, err
	}
	targetDate = time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)
	if !containsDay(schedule.DaysOfWeek, toScheduleWeekday(targetDate)) {
		slog.Info("There are no slots on this date")
		return []dto.SlotResponse{}, nil
	}
	exist, err = ss.slotsRepository.ExistsForRoomAndDate(ctx, roomId, targetDate)
	if err != nil {
		slog.Error("error exist slots chek", "error", err)
		return nil, err
	}
	var slotsDto []dto.SlotResponse
	if !exist {
		startTime, err := parseClock(schedule.StartTime)
		if err != nil {
			slog.Error("failed to parse schedule start time", "error", err)
			return nil, err
		}

		endTime, err := parseClock(schedule.EndTime)
		if err != nil {
			slog.Error("failed to parse schedule end time", "error", err)
			return nil, err
		}
		slots := GenerateForDate(roomId, startTime, endTime, targetDate)
		err = ss.slotsRepository.SlotsBulkCreate(ctx, slots)
		if err != nil {
			slog.Error("error bulk create", "error", err)
			return nil, err
		}
	}
	slots, err := ss.slotsRepository.GetAvailableByRoomAndDate(ctx, roomId, targetDate)
	if err != nil {
		slog.Error("error get slots", "error", err)
		return nil, err
	}
	for _, v := range slots {
		slotsDto = append(slotsDto, dto.SlotResponse{Id: v.ID, RoomID: v.RoomID, Start: v.StartAt, End: v.EndAt})
	}
	return slotsDto, nil
}
