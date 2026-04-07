package service

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"time"

	"github.com/avito-internships/test-backend-1-untrik/internal/dto"
	"github.com/avito-internships/test-backend-1-untrik/internal/models"
	"github.com/avito-internships/test-backend-1-untrik/internal/repository"
	"github.com/google/uuid"
)

var ErrScheduleNotFound = errors.New("schedule not found")
var ErrScheduleExists = errors.New("schedule already exists")
var ErrRoomNotFound = errors.New("room not found")
var timePattern = regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]$`)

type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type schedulesRepositoryInterface interface {
	CreateSchedule(ctx context.Context, schedule models.Schedule) error
}
type RoomChecker interface {
	ExistsRoom(ctx context.Context, roomId string) (bool, error)
}
type slotsCreaterInterface interface {
	SlotsBulkCreate(ctx context.Context, slots []models.Slot) error
}
type schedulesService struct {
	txManager           TxManager
	schedulesRepository schedulesRepositoryInterface
	slotsCreater        slotsCreaterInterface
	roomChecker         RoomChecker
}

func NewSchedulesService(schedulesRepository schedulesRepositoryInterface,
	slotsCreater slotsCreaterInterface,
	roomChecker RoomChecker,
	txManager TxManager) *schedulesService {
	return &schedulesService{schedulesRepository: schedulesRepository,
		slotsCreater: slotsCreater,
		roomChecker:  roomChecker,
		txManager:    txManager}
}

const DAYSWINDOW int = 7

func (ss *schedulesService) CreateSchedule(ctx context.Context, roomId string, daysOfWeek []int,
	startTime, endTime string) (dto.ScheduleResponse, error) {
	if err := uuid.Validate(roomId); err != nil {
		slog.Info("invalid roomId")
		return dto.ScheduleResponse{}, ErrInvalidData
	}
	if err := validateDaysOfWeek(daysOfWeek); err != nil {
		return dto.ScheduleResponse{}, err
	}
	timeSaver, err := validateTime(startTime, endTime)
	if err != nil {
		return dto.ScheduleResponse{}, ErrInvalidData
	}
	exists, err := ss.roomChecker.ExistsRoom(ctx, roomId)
	if err != nil {
		slog.Error("failed to check exists room", "error", err)
		return dto.ScheduleResponse{}, err
	}
	if !exists {
		slog.Info("room not found", "roomId", roomId)
		return dto.ScheduleResponse{}, ErrRoomNotFound
	}
	id := uuid.New()
	slots := GenerateSlotsForRange(roomId, daysOfWeek, timeSaver[0], timeSaver[1], time.Now().UTC(), DAYSWINDOW)
	schedule := models.Schedule{ID: id.String(), RoomID: roomId,
		DaysOfWeek: daysOfWeek, StartTime: startTime, EndTime: endTime}
	err = ss.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := ss.schedulesRepository.CreateSchedule(ctx, schedule); err != nil {
			if errors.Is(err, repository.ErrScheduleExists) {
				slog.Info("schedule already exists")
				return ErrScheduleExists
			}
			slog.Error("error get schedules", "error", err)
			return err
		}
		if err := ss.slotsCreater.SlotsBulkCreate(ctx, slots); err != nil {
			slog.Error("error bulk create", "error", err)
			return err
		}
		return nil
	})
	if err != nil {
		return dto.ScheduleResponse{}, err
	}
	return dto.ScheduleResponse{Id: id.String(), RoomId: roomId, DaysOfWeek: daysOfWeek,
		StartTime: startTime, EndTime: endTime}, nil

}

func validateTime(startTime, endTime string) ([]time.Time, error) {
	if !timePattern.Match([]byte(startTime)) || !timePattern.Match([]byte(endTime)) {
		slog.Info("invalid time format")
		return nil, ErrInvalidData
	}
	startParsed, err := parseClock(startTime)
	if err != nil {
		slog.Info("unsupported time format for parsing startTime")
		return nil, err
	}

	endParsed, err := parseClock(endTime)
	if err != nil {
		slog.Info("unsupported time format for parsing endTime")
		return nil, err
	}
	if !startParsed.Before(endParsed) {
		slog.Info("start time after end time")
		return nil, ErrInvalidData
	}
	timeSaver := make([]time.Time, 0)
	timeSaver = append(timeSaver, startParsed, endParsed)
	return timeSaver, nil
}

func validateDaysOfWeek(daysOfWeek []int) error {
	if len(daysOfWeek) > 7 || len(daysOfWeek) <= 0 {
		slog.Info("invalid length daysOfWeek")
		return ErrInvalidData
	}
	seenDays := make(map[int]struct{}, len(daysOfWeek))
	for _, v := range daysOfWeek {
		if v <= 0 || v > 7 {
			slog.Info("invalid values daysOfWeek")
			return ErrInvalidData
		}
		if _, ok := seenDays[v]; ok {
			slog.Info("duplicate day in daysOfWeek")
			return ErrInvalidData
		}
		seenDays[v] = struct{}{}
	}
	return nil
}

func parseClock(value string) (time.Time, error) {
	layouts := []string{"15:04:05", "15:04", "3:04"}

	for _, v := range layouts {
		t, err := time.Parse(v, value)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, ErrInvalidData
}
