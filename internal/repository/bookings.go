package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/avito-internships/test-backend-1-untrik/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrBookingNotFound = errors.New("booking not found")
var ErrBookingExists = errors.New("booking exists")
var ErrSlotNotFoundOrPast = errors.New("slots not found or in past")

type bookingRepository struct {
	db *sql.DB
}

func NewBookingsRepository(db *sql.DB) *bookingRepository {
	return &bookingRepository{db: db}
}
func (br *bookingRepository) GetAllBookingsPaginated(ctx context.Context, limit, offset int) ([]models.BookingWithStatusCode, error) {
	query := `
		SELECT bookings.id,bookings.slot_id,booking_statuses.code,bookings.user_id,
		bookings.conference_link,bookings.created_at FROM bookings
		JOIN booking_statuses ON  booking_statuses.id = bookings.status_id
		ORDER BY bookings.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := br.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bookings []models.BookingWithStatusCode
	for rows.Next() {
		var booking models.BookingWithStatusCode
		if err := rows.Scan(&booking.ID, &booking.SlotID, &booking.StatusCode, &booking.UserID,
			&booking.ConferenceLink, &booking.CreatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bookings, nil
}
func (br *bookingRepository) GetBookingByUserId(ctx context.Context, userId string) ([]models.BookingWithStatusCode, error) {
	query := `
	SELECT bookings.id,bookings.slot_id,booking_statuses.code,bookings.user_id,
	bookings.conference_link,bookings.created_at FROM bookings
	JOIN booking_statuses ON  booking_statuses.id = bookings.status_id
	JOIN slots ON slots.id = bookings.slot_id
	WHERE bookings.user_id = $1 AND slots.start_at >= NOW()
	ORDER BY slots.start_at
 	`
	rows, err := br.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bookings []models.BookingWithStatusCode
	for rows.Next() {
		var booking models.BookingWithStatusCode
		if err := rows.Scan(&booking.ID, &booking.SlotID, &booking.StatusCode, &booking.UserID,
			&booking.ConferenceLink, &booking.CreatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bookings, nil
}
func (br *bookingRepository) CreateBooking(ctx context.Context, booking models.Booking) (time.Time, error) {
	query := `
	INSERT INTO bookings (id,slot_id,status_id,user_id,
	conference_link,created_at) 
	SELECT $1,slots.id,1,$2,NULL,NOW() 
	FROM slots
	WHERE slots.id = $3 AND slots.start_at >= NOW()
	RETURNING created_at
	`
	var createdAt time.Time
	err := br.db.QueryRowContext(ctx, query, booking.ID, booking.UserID, booking.SlotID).Scan(&createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, ErrSlotNotFoundOrPast
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return createdAt, ErrBookingExists
		}
		return createdAt, err
	}
	return createdAt, nil
}
func (br *bookingRepository) CancelBooking(ctx context.Context, bookingId, userId string) error {
	query := `
	UPDATE bookings SET status_id = 2
	WHERE id = $1 AND user_id = $2
	`
	res, err := br.db.ExecContext(ctx, query, bookingId, userId)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrBookingNotFound
	}
	return nil
}

func (br *bookingRepository) GetBookingById(ctx context.Context, bookingId string) (models.Booking, error) {
	query := `
		SELECT id,slot_id,status_id,user_id,conference_link,created_at FROM bookings
		WHERE id = $1
	`
	var booking models.Booking
	if err := br.db.QueryRowContext(ctx, query, bookingId).Scan(&booking.ID, &booking.SlotID, &booking.StatusID,
		&booking.UserID, &booking.ConferenceLink, &booking.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Booking{}, ErrBookingNotFound
		}
		return models.Booking{}, err
	}
	return booking, nil

}
func (br *bookingRepository) GetCodeByStatusId(ctx context.Context, statusId int) (string, error) {
	query := `
		SELECT code FROM booking_statuses
		WHERE id = $1
	`
	var code string
	if err := br.db.QueryRowContext(ctx, query, statusId).Scan(&code); err != nil {
		return "", err
	}
	return code, nil
}
func (br *bookingRepository) CountBookings(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM bookings`

	var total int
	err := br.db.QueryRowContext(ctx, query).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}
