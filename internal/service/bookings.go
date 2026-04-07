package service

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/avito-internships/test-backend-1-untrik/internal/dto"
	"github.com/avito-internships/test-backend-1-untrik/internal/models"
	"github.com/avito-internships/test-backend-1-untrik/internal/repository"
	"github.com/google/uuid"
)

var ErrAccess = errors.New("cancel other booking")
var ErrSlotNotFoundOrPast = errors.New("slots not found or in past")
var ErrBookingExists = errors.New("booking exists")
var ErrBookingNotFound = errors.New("booking not found")

type BookingsRepositoryInterface interface {
	GetAllBookingsPaginated(ctx context.Context, limit, offset int) ([]models.BookingWithStatusCode, error)
	GetBookingByUserId(ctx context.Context, userId string) ([]models.BookingWithStatusCode, error)
	CreateBooking(ctx context.Context, booking models.Booking) (time.Time, error)
	CancelBooking(ctx context.Context, bookingId, userId string) error
	GetBookingById(ctx context.Context, bookingId string) (models.Booking, error)
	GetCodeByStatusId(ctx context.Context, statusId int) (string, error)
	CountBookings(ctx context.Context) (int, error)
}

type bookingsService struct {
	bookingsRepository BookingsRepositoryInterface
}

func NewBookingsService(bookingsRepository BookingsRepositoryInterface) *bookingsService {
	return &bookingsService{bookingsRepository: bookingsRepository}
}
func (bs *bookingsService) CreateBooking(ctx context.Context, slotId, userId string) (dto.BookingResponse, error) {
	if err := uuid.Validate(slotId); err != nil {
		slog.Info("invalid slotId")
		return dto.BookingResponse{}, ErrInvalidData
	}
	booking := models.Booking{ID: uuid.NewString(), SlotID: slotId, UserID: userId, StatusID: 1}

	createdAt, err := bs.bookingsRepository.CreateBooking(ctx, booking)
	if err != nil {
		if errors.Is(err, repository.ErrSlotNotFoundOrPast) {
			slog.Info("slots not found or in past", "error", err)
			return dto.BookingResponse{}, ErrSlotNotFoundOrPast
		}
		if errors.Is(err, repository.ErrBookingExists) {
			slog.Info("booking exists", "error", err)
			return dto.BookingResponse{}, ErrBookingExists
		}
		slog.Error("error create booking", "error", err)
		return dto.BookingResponse{}, err
	}
	status, err := bs.bookingsRepository.GetCodeByStatusId(ctx, 1)
	if err != nil {
		slog.Error("error get role code", "error", err)
		return dto.BookingResponse{}, err
	}
	return dto.BookingResponse{Id: booking.ID, SlotId: booking.SlotID, UserId: booking.UserID, Status: status, CreatedAt: &createdAt}, nil
}
func (bs *bookingsService) GetMyBooking(ctx context.Context, userId string) ([]dto.BookingResponse, error) {
	bookings, err := bs.bookingsRepository.GetBookingByUserId(ctx, userId)
	if err != nil {
		slog.Error("error get booking by user id", "error", err)
		return nil, err
	}
	var bookingsResponse []dto.BookingResponse
	for _, v := range bookings {
		bookingsResponse = append(bookingsResponse, dto.BookingResponse{Id: v.ID, SlotId: v.SlotID, UserId: v.UserID,
			Status: v.StatusCode, CreatedAt: &v.CreatedAt})
	}
	return bookingsResponse, nil
}
func (bs *bookingsService) CancelBooking(ctx context.Context, bookingId, userId string) (dto.BookingResponse, error) {
	if err := uuid.Validate(bookingId); err != nil {
		slog.Info("invalid bookingId")
		return dto.BookingResponse{}, ErrInvalidData
	}
	booking, err := bs.bookingsRepository.GetBookingById(ctx, bookingId)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			slog.Info("booking not found")
			return dto.BookingResponse{}, ErrBookingNotFound
		}
		slog.Error("error get booking", "error", err)
		return dto.BookingResponse{}, err
	}
	if booking.UserID != userId {
		slog.Info("cancel other booking")
		return dto.BookingResponse{}, ErrAccess
	}
	status, err := bs.bookingsRepository.GetCodeByStatusId(ctx, 2)
	if err != nil {
		slog.Error("error get role code", "error", err)
		return dto.BookingResponse{}, err
	}
	if booking.StatusID == 2 {
		return dto.BookingResponse{Id: booking.ID, SlotId: booking.SlotID, UserId: booking.UserID,
			Status: status, CreatedAt: &booking.CreatedAt}, nil
	}
	if err := bs.bookingsRepository.CancelBooking(ctx, bookingId, userId); err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			slog.Info("booking not found")
			return dto.BookingResponse{}, ErrBookingNotFound
		}
		slog.Error("error cancel booking", "error", err)
		return dto.BookingResponse{}, err
	}
	booking, err = bs.bookingsRepository.GetBookingById(ctx, bookingId)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			slog.Info("booking not found")
			return dto.BookingResponse{}, ErrBookingNotFound
		}
		slog.Error("error get booking", "error", err)
		return dto.BookingResponse{}, err
	}
	return dto.BookingResponse{Id: booking.ID, SlotId: booking.SlotID, UserId: booking.UserID,
		Status: status, CreatedAt: &booking.CreatedAt}, nil
}
func (bs *bookingsService) GetBookingsList(ctx context.Context, pageStr, pageSizeStr string) ([]dto.BookingResponse, dto.Pagination, error) {
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		slog.Info("invalid page")
		return nil, dto.Pagination{}, ErrInvalidData
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		slog.Info("invalid pageSize")
		return nil, dto.Pagination{}, ErrInvalidData
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	bookings, err := bs.bookingsRepository.GetAllBookingsPaginated(ctx, pageSize, offset)
	if err != nil {
		slog.Error("error get paginated bookings", "error", err)
		return nil, dto.Pagination{}, err
	}
	total, err := bs.bookingsRepository.CountBookings(ctx)
	if err != nil {
		slog.Error("error get total bookings", "error", err)
		return nil, dto.Pagination{}, err
	}
	var bookingsResponse []dto.BookingResponse
	for _, v := range bookings {
		bookingsResponse = append(bookingsResponse, dto.BookingResponse{Id: v.ID, SlotId: v.SlotID, UserId: v.UserID,
			Status: v.StatusCode, CreatedAt: &v.CreatedAt})
	}
	return bookingsResponse, dto.Pagination{Page: page, Size: pageSize, Total: total}, nil
}
