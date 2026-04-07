package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-untrik/internal/dto"
	"github.com/avito-internships/test-backend-1-untrik/internal/models"
	"github.com/avito-internships/test-backend-1-untrik/internal/repository"
)

type fakeSlotsRepo struct {
	slotsBulkCreateFn           func(ctx context.Context, slots []models.Slot) error
	getAvailableByRoomAndDateFn func(ctx context.Context, roomId string, date time.Time) ([]models.Slot, error)
	existsForRoomAndDateFn      func(ctx context.Context, roomID string, date time.Time) (bool, error)
}

func (f *fakeSlotsRepo) SlotsBulkCreate(ctx context.Context, slots []models.Slot) error {
	return f.slotsBulkCreateFn(ctx, slots)
}
func (f *fakeSlotsRepo) GetAvailableByRoomAndDate(ctx context.Context, roomId string, date time.Time) ([]models.Slot, error) {
	return f.getAvailableByRoomAndDateFn(ctx, roomId, date)
}
func (f *fakeSlotsRepo) ExistsForRoomAndDate(ctx context.Context, roomID string, date time.Time) (bool, error) {
	return f.existsForRoomAndDateFn(ctx, roomID, date)
}

type fakeScheduleGetter struct {
	getByRoomIDFn func(ctx context.Context, roomID string) (models.Schedule, error)
}

func (f *fakeScheduleGetter) GetByRoomID(ctx context.Context, roomID string) (models.Schedule, error) {
	return f.getByRoomIDFn(ctx, roomID)
}

func TestSlotsService_GetAvailableSlots(t *testing.T) {
	ctx := context.Background()
	roomID := "550e8400-e29b-41d4-a716-446655440000"
	t.Run("empty date", func(t *testing.T) {
		svc := NewSlotsService(&fakeScheduleGetter{}, &fakeSlotsRepo{}, &fakeRoomChecker{})
		_, err := svc.GetAvailableSlots(ctx, roomID, "")
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("invalid date format", func(t *testing.T) {
		svc := NewSlotsService(&fakeScheduleGetter{}, &fakeSlotsRepo{}, &fakeRoomChecker{})
		_, err := svc.GetAvailableSlots(ctx, roomID, "20-03-2026")
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("invalid room id", func(t *testing.T) {
		svc := NewSlotsService(&fakeScheduleGetter{}, &fakeSlotsRepo{}, &fakeRoomChecker{})
		_, err := svc.GetAvailableSlots(ctx, "bad-uuid", "2026-03-23")
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("room checker error", func(t *testing.T) {
		svc := NewSlotsService(
			&fakeScheduleGetter{},
			&fakeSlotsRepo{},
			&fakeRoomChecker{existsRoomFn: func(ctx context.Context, roomId string) (bool, error) {
				return false, errors.New("db error")
			},
			},
		)

		_, err := svc.GetAvailableSlots(ctx, roomID, "2026-03-23")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("room not found", func(t *testing.T) {
		svc := NewSlotsService(
			&fakeScheduleGetter{},
			&fakeSlotsRepo{},
			&fakeRoomChecker{existsRoomFn: func(ctx context.Context, roomId string) (bool, error) {
				return false, nil
			},
			},
		)

		_, err := svc.GetAvailableSlots(ctx, roomID, "2026-03-23")
		if !errors.Is(err, ErrRoomNotFound) {
			t.Fatalf("expected ErrRoomNotFound, got %v", err)
		}
	})

	t.Run("schedule not found", func(t *testing.T) {
		svc := NewSlotsService(
			&fakeScheduleGetter{getByRoomIDFn: func(ctx context.Context, roomID string) (models.Schedule, error) {
				return models.Schedule{}, repository.ErrScheduleNotFound
			},
			},
			&fakeSlotsRepo{},
			&fakeRoomChecker{existsRoomFn: func(ctx context.Context, roomId string) (bool, error) {
				return true, nil
			},
			},
		)

		_, err := svc.GetAvailableSlots(ctx, roomID, "2026-03-23")
		if !errors.Is(err, ErrScheduleNotFound) {
			t.Fatalf("expected ErrScheduleNotFound, got %v", err)
		}
	})

	t.Run("schedule getter unexpected error", func(t *testing.T) {
		svc := NewSlotsService(
			&fakeScheduleGetter{getByRoomIDFn: func(ctx context.Context, roomID string) (models.Schedule, error) {
				return models.Schedule{}, errors.New("db error")
			},
			},
			&fakeSlotsRepo{},
			&fakeRoomChecker{existsRoomFn: func(ctx context.Context, roomId string) (bool, error) {
				return true, nil
			},
			},
		)
		_, err := svc.GetAvailableSlots(ctx, roomID, "2026-03-23")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("date is not in schedule days", func(t *testing.T) {
		svc := NewSlotsService(
			&fakeScheduleGetter{
				getByRoomIDFn: func(ctx context.Context, roomID string) (models.Schedule, error) {
					return models.Schedule{RoomID: roomID, DaysOfWeek: []int{2, 4}, StartTime: "09:00", EndTime: "18:00"}, nil
				},
			},
			&fakeSlotsRepo{existsForRoomAndDateFn: func(ctx context.Context, roomID string, date time.Time) (bool, error) {
				t.Fatal("ExistsForRoomAndDate should not be called when day is outside schedule")
				return false, nil
			},
			},
			&fakeRoomChecker{existsRoomFn: func(ctx context.Context, roomId string) (bool, error) {
				return true, nil
			},
			},
		)

		got, err := svc.GetAvailableSlots(ctx, roomID, "2026-03-23")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("expected empty slice, got %+v", got)
		}
	})

	t.Run("exists for room and date error", func(t *testing.T) {
		svc := NewSlotsService(
			&fakeScheduleGetter{
				getByRoomIDFn: func(ctx context.Context, roomID string) (models.Schedule, error) {
					return models.Schedule{RoomID: roomID, DaysOfWeek: []int{1, 3, 5}, StartTime: "09:00", EndTime: "18:00"}, nil
				},
			},
			&fakeSlotsRepo{existsForRoomAndDateFn: func(ctx context.Context, roomID string, date time.Time) (bool, error) {
				return false, errors.New("db error")
			},
			},
			&fakeRoomChecker{existsRoomFn: func(ctx context.Context, roomId string) (bool, error) {
				return true, nil
			},
			},
		)
		_, err := svc.GetAvailableSlots(ctx, roomID, "2026-03-23")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("generate slots when not exist and bulk create fails", func(t *testing.T) {
		var gotGeneratedSlots []models.Slot

		svc := NewSlotsService(
			&fakeScheduleGetter{
				getByRoomIDFn: func(ctx context.Context, roomID string) (models.Schedule, error) {
					return models.Schedule{RoomID: roomID, DaysOfWeek: []int{1, 3, 5}, StartTime: "09:00", EndTime: "10:00"}, nil
				},
			},
			&fakeSlotsRepo{
				existsForRoomAndDateFn: func(ctx context.Context, roomID string, date time.Time) (bool, error) {
					return false, nil
				},
				slotsBulkCreateFn: func(ctx context.Context, slots []models.Slot) error {
					gotGeneratedSlots = slots
					return errors.New("bulk error")
				},
				getAvailableByRoomAndDateFn: func(ctx context.Context, roomId string, date time.Time) ([]models.Slot, error) {
					t.Fatal("GetAvailableByRoomAndDate should not be called when bulk create fails")
					return nil, nil
				},
			},
			&fakeRoomChecker{existsRoomFn: func(ctx context.Context, roomId string) (bool, error) {
				return true, nil
			},
			},
		)

		_, err := svc.GetAvailableSlots(ctx, roomID, "2026-03-23")
		if len(gotGeneratedSlots) == 0 {
			t.Fatal("expected generated slots, got empty slice")
		}
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("slots already exist and get available fails", func(t *testing.T) {
		svc := NewSlotsService(
			&fakeScheduleGetter{
				getByRoomIDFn: func(ctx context.Context, roomID string) (models.Schedule, error) {
					return models.Schedule{RoomID: roomID, DaysOfWeek: []int{1, 3, 5}, StartTime: "09:00", EndTime: "18:00"}, nil
				},
			},
			&fakeSlotsRepo{
				existsForRoomAndDateFn: func(ctx context.Context, roomID string, date time.Time) (bool, error) {
					return true, nil
				},
				slotsBulkCreateFn: func(ctx context.Context, slots []models.Slot) error {
					t.Fatal("SlotsBulkCreate should not be called when slots already exist")
					return nil
				},
				getAvailableByRoomAndDateFn: func(ctx context.Context, roomId string, date time.Time) ([]models.Slot, error) {
					return nil, errors.New("db error")
				},
			},
			&fakeRoomChecker{existsRoomFn: func(ctx context.Context, roomId string) (bool, error) {
				return true, nil
			},
			},
		)
		_, err := svc.GetAvailableSlots(ctx, roomID, "2026-03-23")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("success when slots already exist", func(t *testing.T) {
		targetDate := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
		start1 := time.Date(2026, 3, 23, 9, 0, 0, 0, time.UTC)
		end1 := time.Date(2026, 3, 23, 9, 30, 0, 0, time.UTC)
		start2 := time.Date(2026, 3, 23, 9, 30, 0, 0, time.UTC)
		end2 := time.Date(2026, 3, 23, 10, 0, 0, 0, time.UTC)

		svc := NewSlotsService(
			&fakeScheduleGetter{
				getByRoomIDFn: func(ctx context.Context, roomID string) (models.Schedule, error) {
					return models.Schedule{RoomID: roomID, DaysOfWeek: []int{1, 3, 5}, StartTime: "09:00", EndTime: "10:00"}, nil
				},
			},
			&fakeSlotsRepo{
				existsForRoomAndDateFn: func(ctx context.Context, roomID string, date time.Time) (bool, error) {
					if !date.Equal(targetDate) {
						t.Fatalf("expected target date %v, got %v", targetDate, date)
					}
					return true, nil
				},
				slotsBulkCreateFn: func(ctx context.Context, slots []models.Slot) error {
					t.Fatal("SlotsBulkCreate should not be called when slots already exist")
					return nil
				},
				getAvailableByRoomAndDateFn: func(ctx context.Context, roomId string, date time.Time) ([]models.Slot, error) {
					return []models.Slot{
						{ID: "slot-1", RoomID: roomId, StartAt: start1, EndAt: end1},
						{ID: "slot-2", RoomID: roomId, StartAt: start2, EndAt: end2},
					}, nil
				},
			},
			&fakeRoomChecker{
				existsRoomFn: func(ctx context.Context, roomId string) (bool, error) {
					return true, nil
				},
			},
		)

		got, err := svc.GetAvailableSlots(ctx, roomID, "2026-03-23")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []dto.SlotResponse{
			{Id: "slot-1", RoomID: roomID, Start: start1, End: end1},
			{Id: "slot-2", RoomID: roomID, Start: start2, End: end2},
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %+v, got %+v", want, got)
		}
	})

	t.Run("success when slots do not exist yet", func(t *testing.T) {
		bulkCalled := false
		var gotGeneratedSlots []models.Slot

		svc := NewSlotsService(
			&fakeScheduleGetter{
				getByRoomIDFn: func(ctx context.Context, roomID string) (models.Schedule, error) {
					return models.Schedule{RoomID: roomID, DaysOfWeek: []int{1, 3, 5}, StartTime: "09:00", EndTime: "10:00"}, nil
				},
			},
			&fakeSlotsRepo{
				existsForRoomAndDateFn: func(ctx context.Context, roomID string, date time.Time) (bool, error) {
					return false, nil
				},
				slotsBulkCreateFn: func(ctx context.Context, slots []models.Slot) error {
					bulkCalled = true
					gotGeneratedSlots = slots
					return nil
				},
				getAvailableByRoomAndDateFn: func(ctx context.Context, roomId string, date time.Time) ([]models.Slot, error) {
					return []models.Slot{
						{ID: "slot-1", RoomID: roomId, StartAt: time.Date(2026, 3, 23, 9, 0, 0, 0, time.UTC), EndAt: time.Date(2026, 3, 23, 9, 30, 0, 0, time.UTC)},
					}, nil
				},
			},
			&fakeRoomChecker{
				existsRoomFn: func(ctx context.Context, roomId string) (bool, error) {
					return true, nil
				},
			},
		)
		got, err := svc.GetAvailableSlots(ctx, roomID, "2026-03-23")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bulkCalled {
			t.Fatal("expected SlotsBulkCreate to be called")
		}
		if len(gotGeneratedSlots) == 0 {
			t.Fatal("expected generated slots, got empty slice")
		}
		if len(got) != 1 {
			t.Fatalf("expected 1 slot response, got %d", len(got))
		}
	})
}
