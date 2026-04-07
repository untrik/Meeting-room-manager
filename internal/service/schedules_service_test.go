package service

import (
	"context"
	"errors"
	"testing"

	"github.com/avito-internships/test-backend-1-untrik/internal/models"
	"github.com/avito-internships/test-backend-1-untrik/internal/repository"
)

type fakeSchedulesRepo struct {
	createScheduleFn func(ctx context.Context, schedule models.Schedule) error
}

func (f *fakeSchedulesRepo) CreateSchedule(ctx context.Context, schedule models.Schedule) error {
	return f.createScheduleFn(ctx, schedule)
}

type fakeSlotsCreater struct {
	slotsBulkCreateFn func(ctx context.Context, slots []models.Slot) error
}

func (f *fakeSlotsCreater) SlotsBulkCreate(ctx context.Context, slots []models.Slot) error {
	return f.slotsBulkCreateFn(ctx, slots)
}

type fakeRoomChecker struct {
	existsRoomFn func(ctx context.Context, roomId string) (bool, error)
}

func (f *fakeRoomChecker) ExistsRoom(ctx context.Context, roomId string) (bool, error) {
	return f.existsRoomFn(ctx, roomId)
}

type fakeTxManager struct {
	withinTransactionFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (f *fakeTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return f.withinTransactionFn(ctx, fn)
}
func TestSchedulesService_CreateSchedule(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid room id", func(t *testing.T) {
		svc := NewSchedulesService(&fakeSchedulesRepo{}, &fakeSlotsCreater{}, &fakeRoomChecker{}, &fakeTxManager{})
		_, err := svc.CreateSchedule(ctx, "bad-uuid", []int{1, 2}, "09:00", "18:00")
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("invalid days of week", func(t *testing.T) {
		svc := NewSchedulesService(&fakeSchedulesRepo{}, &fakeSlotsCreater{}, &fakeRoomChecker{}, &fakeTxManager{})
		_, err := svc.CreateSchedule(ctx, "550e8400-e29b-41d4-a716-446655440000", []int{1, 1}, "09:00", "18:00")
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("invalid time", func(t *testing.T) {
		svc := NewSchedulesService(&fakeSchedulesRepo{}, &fakeSlotsCreater{}, &fakeRoomChecker{}, &fakeTxManager{})

		_, err := svc.CreateSchedule(ctx, "550e8400-e29b-41d4-a716-446655440000", []int{1, 2}, "18:00", "09:00")
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("room checker error", func(t *testing.T) {
		svc := NewSchedulesService(&fakeSchedulesRepo{}, &fakeSlotsCreater{}, &fakeRoomChecker{
			existsRoomFn: func(ctx context.Context, roomId string) (bool, error) {
				return false, errors.New("db error")
			},
		},
			&fakeTxManager{},
		)
		_, err := svc.CreateSchedule(ctx, "550e8400-e29b-41d4-a716-446655440000", []int{1, 2}, "09:00", "18:00")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
	t.Run("room not found", func(t *testing.T) {
		svc := NewSchedulesService(&fakeSchedulesRepo{}, &fakeSlotsCreater{}, &fakeRoomChecker{
			existsRoomFn: func(ctx context.Context, roomId string) (bool, error) {
				return false, nil
			},
		},
			&fakeTxManager{},
		)
		_, err := svc.CreateSchedule(ctx, "550e8400-e29b-41d4-a716-446655440000", []int{1, 2}, "09:00", "18:00")
		if !errors.Is(err, ErrRoomNotFound) {
			t.Fatalf("expected ErrRoomNotFound, got %v", err)
		}
	})

	t.Run("schedule already exists", func(t *testing.T) {
		txCalled := false

		svc := NewSchedulesService(&fakeSchedulesRepo{
			createScheduleFn: func(ctx context.Context, schedule models.Schedule) error {
				return repository.ErrScheduleExists
			},
		},
			&fakeSlotsCreater{
				slotsBulkCreateFn: func(ctx context.Context, slots []models.Slot) error {
					t.Fatal("SlotsBulkCreate must not be called when schedule exists")
					return nil
				},
			},
			&fakeRoomChecker{
				existsRoomFn: func(ctx context.Context, roomId string) (bool, error) {
					return true, nil
				},
			},
			&fakeTxManager{
				withinTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
					txCalled = true
					return fn(ctx)
				},
			},
		)

		_, err := svc.CreateSchedule(ctx, "550e8400-e29b-41d4-a716-446655440000", []int{1, 2}, "09:00", "18:00")
		if !txCalled {
			t.Fatal("expected transaction to be called")
		}
		if !errors.Is(err, ErrScheduleExists) {
			t.Fatalf("expected ErrScheduleExists, got %v", err)
		}
	})

	t.Run("slots create error", func(t *testing.T) {
		var gotSchedule models.Schedule
		txCalled := false

		svc := NewSchedulesService(
			&fakeSchedulesRepo{
				createScheduleFn: func(ctx context.Context, schedule models.Schedule) error {
					gotSchedule = schedule
					return nil
				},
			},
			&fakeSlotsCreater{
				slotsBulkCreateFn: func(ctx context.Context, slots []models.Slot) error {
					if len(slots) == 0 {
						t.Fatal("expected generated slots, got empty slice")
					}
					return errors.New("bulk error")
				},
			},
			&fakeRoomChecker{
				existsRoomFn: func(ctx context.Context, roomId string) (bool, error) {
					return true, nil
				},
			},
			&fakeTxManager{
				withinTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
					txCalled = true
					return fn(ctx)
				},
			},
		)

		roomID := "550e8400-e29b-41d4-a716-446655440000"
		_, err := svc.CreateSchedule(ctx, roomID, []int{1, 2}, "09:00", "18:00")

		if !txCalled {
			t.Fatal("expected transaction to be called")
		}
		if gotSchedule.RoomID != roomID {
			t.Fatalf("expected room id %s, got %s", roomID, gotSchedule.RoomID)
		}
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
	t.Run("success", func(t *testing.T) {
		var gotSchedule models.Schedule
		var gotSlots []models.Slot
		txCalled := false

		svc := NewSchedulesService(
			&fakeSchedulesRepo{
				createScheduleFn: func(ctx context.Context, schedule models.Schedule) error {
					gotSchedule = schedule
					return nil
				},
			},
			&fakeSlotsCreater{
				slotsBulkCreateFn: func(ctx context.Context, slots []models.Slot) error {
					gotSlots = slots
					return nil
				},
			},
			&fakeRoomChecker{
				existsRoomFn: func(ctx context.Context, roomId string) (bool, error) {
					return true, nil
				},
			},
			&fakeTxManager{
				withinTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
					txCalled = true
					return fn(ctx)
				},
			},
		)

		roomID := "550e8400-e29b-41d4-a716-446655440000"
		resp, err := svc.CreateSchedule(ctx, roomID, []int{1, 3, 5}, "09:00", "18:00")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !txCalled {
			t.Fatal("expected transaction to be called")
		}
		if gotSchedule.ID == "" {
			t.Fatal("expected generated schedule ID")
		}
		if gotSchedule.RoomID != roomID {
			t.Fatalf("expected roomID %s, got %s", roomID, gotSchedule.RoomID)
		}
		if gotSchedule.StartTime != "09:00" {
			t.Fatalf("expected start time 09:00, got %s", gotSchedule.StartTime)
		}
		if gotSchedule.EndTime != "18:00" {
			t.Fatalf("expected end time 18:00, got %s", gotSchedule.EndTime)
		}
		if len(gotSlots) == 0 {
			t.Fatal("expected generated slots, got empty slice")
		}

		if resp.Id != gotSchedule.ID {
			t.Fatalf("expected response ID %s, got %s", gotSchedule.ID, resp.Id)
		}
		if resp.RoomId != roomID {
			t.Fatalf("expected response roomID %s, got %s", roomID, resp.RoomId)
		}
		if resp.StartTime != "09:00" || resp.EndTime != "18:00" {
			t.Fatalf("unexpected response times: %+v", resp)
		}
	})
}

func TestValidateTime(t *testing.T) {
	t.Run("invalid start format", func(t *testing.T) {
		_, err := validateTime("abc", "18:00")
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("invalid end format", func(t *testing.T) {
		_, err := validateTime("09:00", "25:00")
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("start not before end", func(t *testing.T) {
		_, err := validateTime("18:00", "09:00")
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("equal times", func(t *testing.T) {
		_, err := validateTime("09:00", "09:00")
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		got, err := validateTime("09:00", "18:30")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2 parsed times, got %d", len(got))
		}
	})
}

func TestValidateDaysOfWeek(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		err := validateDaysOfWeek([]int{})
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("too many days", func(t *testing.T) {
		err := validateDaysOfWeek([]int{1, 2, 3, 4, 5, 6, 7, 1})
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("day less than 1", func(t *testing.T) {
		err := validateDaysOfWeek([]int{0, 2})
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("day greater than 7", func(t *testing.T) {
		err := validateDaysOfWeek([]int{1, 8})
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("duplicate day", func(t *testing.T) {
		err := validateDaysOfWeek([]int{1, 2, 2})
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		err := validateDaysOfWeek([]int{1, 3, 5, 7})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestParseClock(t *testing.T) {
	t.Run("format 15:04", func(t *testing.T) {
		got, err := parseClock("09:30")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Hour() != 9 || got.Minute() != 30 {
			t.Fatalf("expected 09:30, got %02d:%02d", got.Hour(), got.Minute())
		}
	})

	t.Run("format 3:04", func(t *testing.T) {
		got, err := parseClock("9:30")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Hour() != 9 || got.Minute() != 30 {
			t.Fatalf("expected 09:30, got %02d:%02d", got.Hour(), got.Minute())
		}
	})

	t.Run("format 15:04:05", func(t *testing.T) {
		got, err := parseClock("09:30:15")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Hour() != 9 || got.Minute() != 30 || got.Second() != 15 {
			t.Fatalf("expected 09:30:15, got %02d:%02d:%02d", got.Hour(), got.Minute(), got.Second())
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		_, err := parseClock("hello")
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})
}
func TestSchedulesService_CreateSchedule_ResponseShape(t *testing.T) {
	ctx := context.Background()

	svc := NewSchedulesService(
		&fakeSchedulesRepo{createScheduleFn: func(ctx context.Context, schedule models.Schedule) error {
			return nil
		},
		},
		&fakeSlotsCreater{slotsBulkCreateFn: func(ctx context.Context, slots []models.Slot) error {
			return nil
		},
		},
		&fakeRoomChecker{
			existsRoomFn: func(ctx context.Context, roomId string) (bool, error) {
				return true, nil
			},
		},
		&fakeTxManager{
			withinTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		},
	)
	resp, err := svc.CreateSchedule(ctx, "550e8400-e29b-41d4-a716-446655440000", []int{1, 2}, "09:00", "18:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Id == "" {
		t.Fatal("expected generated ID")
	}
	if resp.RoomId != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("unexpected room id: %s", resp.RoomId)
	}
	if len(resp.DaysOfWeek) != 2 {
		t.Fatalf("expected 2 days, got %d", len(resp.DaysOfWeek))
	}
	if resp.DaysOfWeek[0] != 1 || resp.DaysOfWeek[1] != 2 {
		t.Fatalf("unexpected daysOfWeek: %+v", resp.DaysOfWeek)
	}
	if resp.StartTime != "09:00" {
		t.Fatalf("expected start time 09:00, got %s", resp.StartTime)
	}
	if resp.EndTime != "18:00" {
		t.Fatalf("expected end time 18:00, got %s", resp.EndTime)
	}
}
