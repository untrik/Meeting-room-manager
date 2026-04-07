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
	"github.com/google/uuid"
)

type fakeBookingsRepo struct {
	getAllBookingsPaginatedFn func(ctx context.Context, limit, offset int) ([]models.BookingWithStatusCode, error)
	getBookingByUserIdFn      func(ctx context.Context, userId string) ([]models.BookingWithStatusCode, error)
	createBookingFn           func(ctx context.Context, booking models.Booking) (time.Time, error)
	cancelBookingFn           func(ctx context.Context, bookingId, userId string) error
	getBookingByIdFn          func(ctx context.Context, bookingId string) (models.Booking, error)
	getCodeByStatusIdFn       func(ctx context.Context, statusId int) (string, error)
	countBookingsFn           func(ctx context.Context) (int, error)
}

func (f *fakeBookingsRepo) GetAllBookingsPaginated(ctx context.Context, limit, offset int) ([]models.BookingWithStatusCode, error) {
	return f.getAllBookingsPaginatedFn(ctx, limit, offset)
}
func (f *fakeBookingsRepo) GetBookingByUserId(ctx context.Context, userId string) ([]models.BookingWithStatusCode, error) {
	return f.getBookingByUserIdFn(ctx, userId)
}
func (f *fakeBookingsRepo) CreateBooking(ctx context.Context, booking models.Booking) (time.Time, error) {
	return f.createBookingFn(ctx, booking)
}
func (f *fakeBookingsRepo) CancelBooking(ctx context.Context, bookingId, userId string) error {
	return f.cancelBookingFn(ctx, bookingId, userId)
}
func (f *fakeBookingsRepo) GetBookingById(ctx context.Context, bookingId string) (models.Booking, error) {
	return f.getBookingByIdFn(ctx, bookingId)
}
func (f *fakeBookingsRepo) GetCodeByStatusId(ctx context.Context, statusId int) (string, error) {
	return f.getCodeByStatusIdFn(ctx, statusId)
}
func (f *fakeBookingsRepo) CountBookings(ctx context.Context) (int, error) {
	return f.countBookingsFn(ctx)
}
func TestBookingsService_CreateBooking(t *testing.T) {
	ctx := context.Background()
	t.Run("invalid slot id", func(t *testing.T) {
		repo := &fakeBookingsRepo{}
		svc := NewBookingsService(repo)
		_, err := svc.CreateBooking(ctx, "not-uuid", "user-1")
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})
	t.Run("slot not found or in past", func(t *testing.T) {
		repo := &fakeBookingsRepo{
			createBookingFn: func(ctx context.Context, booking models.Booking) (time.Time, error) {
				return time.Time{}, repository.ErrSlotNotFoundOrPast
			},
		}
		svc := NewBookingsService(repo)
		_, err := svc.CreateBooking(ctx, uuid.NewString(), "user-1")
		if !errors.Is(err, ErrSlotNotFoundOrPast) {
			t.Fatalf("expected ErrSlotNotFoundOrPast, got %v", err)
		}
	})
	t.Run("booking already exists", func(t *testing.T) {
		repo := &fakeBookingsRepo{
			createBookingFn: func(ctx context.Context, booking models.Booking) (time.Time, error) {
				return time.Time{}, repository.ErrBookingExists
			},
		}
		svc := NewBookingsService(repo)

		_, err := svc.CreateBooking(ctx, uuid.NewString(), "user-1")
		if !errors.Is(err, ErrBookingExists) {
			t.Fatalf("expected ErrBookingExists, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		slotID := uuid.NewString()
		userID := "user-1"
		createdAt := time.Now()
		var savedBooking models.Booking
		repo := &fakeBookingsRepo{
			createBookingFn: func(ctx context.Context, booking models.Booking) (time.Time, error) {
				savedBooking = booking
				return createdAt, nil
			},
			getCodeByStatusIdFn: func(ctx context.Context, statusId int) (string, error) {
				if statusId != 1 {
					t.Fatalf("expected statusId=1, got %d", statusId)
				}
				return "active", nil
			},
		}
		svc := NewBookingsService(repo)

		resp, err := svc.CreateBooking(ctx, slotID, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if savedBooking.SlotID != slotID {
			t.Fatalf("expected slotID %s, got %s", slotID, savedBooking.SlotID)
		}
		if savedBooking.UserID != userID {
			t.Fatalf("expected userID %s, got %s", userID, savedBooking.UserID)
		}
		if savedBooking.StatusID != 1 {
			t.Fatalf("expected statusID 1, got %d", savedBooking.StatusID)
		}
		if err := uuid.Validate(savedBooking.ID); err != nil {
			t.Fatalf("booking ID must be uuid, got %s", savedBooking.ID)
		}

		if resp.Id != savedBooking.ID {
			t.Fatalf("expected response ID %s, got %s", savedBooking.ID, resp.Id)
		}
		if resp.SlotId != slotID {
			t.Fatalf("expected slotID %s, got %s", slotID, resp.SlotId)
		}
		if resp.UserId != userID {
			t.Fatalf("expected userID %s, got %s", userID, resp.UserId)
		}
		if resp.Status != "active" {
			t.Fatalf("expected status active, got %s", resp.Status)
		}
		if resp.CreatedAt == nil || !resp.CreatedAt.Equal(createdAt) {
			t.Fatalf("expected createdAt %v, got %v", createdAt, resp.CreatedAt)
		}
	})
}

func TestBookingsService_GetMyBooking(t *testing.T) {
	ctx := context.Background()

	t.Run("repository error", func(t *testing.T) {
		repo := &fakeBookingsRepo{
			getBookingByUserIdFn: func(ctx context.Context, userId string) ([]models.BookingWithStatusCode, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewBookingsService(repo)
		_, err := svc.GetMyBooking(ctx, "user-1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		repo := &fakeBookingsRepo{
			getBookingByUserIdFn: func(ctx context.Context, userId string) ([]models.BookingWithStatusCode, error) {
				return []models.BookingWithStatusCode{
					{
						ID:         "b1",
						SlotID:     "s1",
						UserID:     userId,
						StatusCode: "active",
						CreatedAt:  now,
					},
				}, nil
			},
		}
		svc := NewBookingsService(repo)

		got, err := svc.GetMyBooking(ctx, "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []dto.BookingResponse{
			{
				Id:        "b1",
				SlotId:    "s1",
				UserId:    "user-1",
				Status:    "active",
				CreatedAt: &now,
			},
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %+v, got %+v", want, got)
		}
	})
}

func TestBookingsService_CancelBooking(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid booking id", func(t *testing.T) {
		repo := &fakeBookingsRepo{}
		svc := NewBookingsService(repo)

		_, err := svc.CancelBooking(ctx, "bad-id", "user-1")
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("booking not found on first get", func(t *testing.T) {
		repo := &fakeBookingsRepo{
			getBookingByIdFn: func(ctx context.Context, bookingId string) (models.Booking, error) {
				return models.Booking{}, repository.ErrBookingNotFound
			},
		}
		svc := NewBookingsService(repo)

		_, err := svc.CancelBooking(ctx, uuid.NewString(), "user-1")
		if !errors.Is(err, ErrBookingNotFound) {
			t.Fatalf("expected ErrBookingNotFound, got %v", err)
		}
	})

	t.Run("access denied for чужую бронь, классика жанра", func(t *testing.T) {
		repo := &fakeBookingsRepo{
			getBookingByIdFn: func(ctx context.Context, bookingId string) (models.Booking, error) {
				return models.Booking{
					ID:       bookingId,
					SlotID:   uuid.NewString(),
					UserID:   "another-user",
					StatusID: 1,
				}, nil
			},
		}
		svc := NewBookingsService(repo)

		_, err := svc.CancelBooking(ctx, uuid.NewString(), "user-1")
		if !errors.Is(err, ErrAccess) {
			t.Fatalf("expected ErrAccess, got %v", err)
		}
	})

	t.Run("already cancelled", func(t *testing.T) {
		bookingID := uuid.NewString()
		slotID := uuid.NewString()
		createdAt := time.Now()

		repo := &fakeBookingsRepo{
			getBookingByIdFn: func(ctx context.Context, id string) (models.Booking, error) {
				return models.Booking{
					ID:        bookingID,
					SlotID:    slotID,
					UserID:    "user-1",
					StatusID:  2,
					CreatedAt: createdAt,
				}, nil
			},
			getCodeByStatusIdFn: func(ctx context.Context, statusId int) (string, error) {
				if statusId != 2 {
					t.Fatalf("expected statusId=2, got %d", statusId)
				}
				return "cancelled", nil
			},
			cancelBookingFn: func(ctx context.Context, bookingId, userId string) error {
				t.Fatal("cancelBooking must not be called for already cancelled booking")
				return nil
			},
		}
		svc := NewBookingsService(repo)

		resp, err := svc.CancelBooking(ctx, bookingID, "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Status != "cancelled" {
			t.Fatalf("expected cancelled, got %s", resp.Status)
		}
		if resp.Id != bookingID || resp.SlotId != slotID || resp.UserId != "user-1" {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})

	t.Run("cancel success", func(t *testing.T) {
		bookingID := uuid.NewString()
		slotID := uuid.NewString()
		createdAt := time.Now()

		callCountGetByID := 0
		cancelCalled := false

		repo := &fakeBookingsRepo{
			getBookingByIdFn: func(ctx context.Context, id string) (models.Booking, error) {
				callCountGetByID++
				if callCountGetByID == 1 {
					return models.Booking{
						ID:        bookingID,
						SlotID:    slotID,
						UserID:    "user-1",
						StatusID:  1,
						CreatedAt: createdAt,
					}, nil
				}
				return models.Booking{
					ID:        bookingID,
					SlotID:    slotID,
					UserID:    "user-1",
					StatusID:  2,
					CreatedAt: createdAt,
				}, nil
			},
			getCodeByStatusIdFn: func(ctx context.Context, statusId int) (string, error) {
				return "cancelled", nil
			},
			cancelBookingFn: func(ctx context.Context, bookingId, userId string) error {
				cancelCalled = true
				if bookingId != bookingID {
					t.Fatalf("expected bookingId %s, got %s", bookingID, bookingId)
				}
				if userId != "user-1" {
					t.Fatalf("expected userId user-1, got %s", userId)
				}
				return nil
			},
		}
		svc := NewBookingsService(repo)

		resp, err := svc.CancelBooking(ctx, bookingID, "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cancelCalled {
			t.Fatal("expected cancelBooking to be called")
		}
		if callCountGetByID != 2 {
			t.Fatalf("expected GetBookingById to be called 2 times, got %d", callCountGetByID)
		}
		if resp.Status != "cancelled" {
			t.Fatalf("expected cancelled, got %s", resp.Status)
		}
	})

	t.Run("cancel returns booking not found", func(t *testing.T) {
		bookingID := uuid.NewString()

		repo := &fakeBookingsRepo{
			getBookingByIdFn: func(ctx context.Context, id string) (models.Booking, error) {
				return models.Booking{
					ID:       bookingID,
					SlotID:   uuid.NewString(),
					UserID:   "user-1",
					StatusID: 1,
				}, nil
			},
			getCodeByStatusIdFn: func(ctx context.Context, statusId int) (string, error) {
				return "cancelled", nil
			},
			cancelBookingFn: func(ctx context.Context, bookingId, userId string) error {
				return repository.ErrBookingNotFound
			},
		}
		svc := NewBookingsService(repo)

		_, err := svc.CancelBooking(ctx, bookingID, "user-1")
		if !errors.Is(err, ErrBookingNotFound) {
			t.Fatalf("expected ErrBookingNotFound, got %v", err)
		}
	})
}

func TestBookingsService_GetBookingsList(t *testing.T) {
	ctx := context.Background()
	t.Run("invalid page", func(t *testing.T) {
		repo := &fakeBookingsRepo{}
		svc := NewBookingsService(repo)
		_, _, err := svc.GetBookingsList(ctx, "abc", "10")
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("invalid page size", func(t *testing.T) {
		repo := &fakeBookingsRepo{}
		svc := NewBookingsService(repo)
		_, _, err := svc.GetBookingsList(ctx, "1", "abc")
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})
	t.Run("defaults and limits", func(t *testing.T) {
		now := time.Now()
		var gotLimit, gotOffset int
		repo := &fakeBookingsRepo{
			getAllBookingsPaginatedFn: func(ctx context.Context, limit, offset int) ([]models.BookingWithStatusCode, error) {
				gotLimit = limit
				gotOffset = offset
				return []models.BookingWithStatusCode{
					{
						ID: "b1", SlotID: "s1", UserID: "u1", StatusCode: "active", CreatedAt: now},
				}, nil
			},
			countBookingsFn: func(ctx context.Context) (int, error) {
				return 1, nil
			},
		}
		svc := NewBookingsService(repo)
		resp, pagination, err := svc.GetBookingsList(ctx, "0", "200")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotLimit != 100 {
			t.Fatalf("expected limit 100, got %d", gotLimit)
		}
		if gotOffset != 0 {
			t.Fatalf("expected offset 0, got %d", gotOffset)
		}
		wantPagination := dto.Pagination{Page: 1, Size: 100, Total: 1}
		if !reflect.DeepEqual(pagination, wantPagination) {
			t.Fatalf("expected pagination %+v, got %+v", wantPagination, pagination)
		}
		wantResp := []dto.BookingResponse{
			{Id: "b1", SlotId: "s1", UserId: "u1", Status: "active", CreatedAt: &now},
		}
		if !reflect.DeepEqual(resp, wantResp) {
			t.Fatalf("expected %+v, got %+v", wantResp, resp)
		}
	})
	t.Run("page and pageSize less than 1 become defaults", func(t *testing.T) {
		var gotLimit, gotOffset int
		repo := &fakeBookingsRepo{
			getAllBookingsPaginatedFn: func(ctx context.Context, limit, offset int) ([]models.BookingWithStatusCode, error) {
				gotLimit = limit
				gotOffset = offset
				return []models.BookingWithStatusCode{}, nil
			},
			countBookingsFn: func(ctx context.Context) (int, error) {
				return 0, nil
			},
		}
		svc := NewBookingsService(repo)
		_, pagination, err := svc.GetBookingsList(ctx, "-10", "-5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotLimit != 20 {
			t.Fatalf("expected limit 20, got %d", gotLimit)
		}
		if gotOffset != 0 {
			t.Fatalf("expected offset 0, got %d", gotOffset)
		}
		if pagination.Page != 1 || pagination.Size != 20 {
			t.Fatalf("unexpected pagination: %+v", pagination)
		}
	})
	t.Run("repository GetAllBookingsPaginated error", func(t *testing.T) {
		repo := &fakeBookingsRepo{
			getAllBookingsPaginatedFn: func(ctx context.Context, limit, offset int) ([]models.BookingWithStatusCode, error) {
				return nil, errors.New("db error")
			},
			countBookingsFn: func(ctx context.Context) (int, error) {
				return 0, nil
			},
		}
		svc := NewBookingsService(repo)
		_, _, err := svc.GetBookingsList(ctx, "1", "10")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
	t.Run("repository CountBookings error", func(t *testing.T) {
		repo := &fakeBookingsRepo{
			getAllBookingsPaginatedFn: func(ctx context.Context, limit, offset int) ([]models.BookingWithStatusCode, error) {
				return []models.BookingWithStatusCode{}, nil
			},
			countBookingsFn: func(ctx context.Context) (int, error) {
				return 0, errors.New("db error")
			},
		}
		svc := NewBookingsService(repo)
		_, _, err := svc.GetBookingsList(ctx, "1", "10")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
