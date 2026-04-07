package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/avito-internships/test-backend-1-untrik/internal/models"
	"github.com/avito-internships/test-backend-1-untrik/internal/transactions"
)

type slotsRepository struct {
	db *sql.DB
}

func NewSlotsRepository(db *sql.DB) *slotsRepository {
	return &slotsRepository{db: db}
}

func (sr *slotsRepository) GetAvailableByRoomAndDate(ctx context.Context, roomId string, date time.Time) ([]models.Slot, error) {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, 0, 1)
	query := `
	SELECT slots.id,slots.room_id,slots.start_at,slots.end_at FROM slots
	LEFT JOIN bookings ON bookings.slot_id = slots.id AND bookings.status_id = 1
	WHERE slots.room_id = $1 AND slots.start_at >= $2 AND slots.end_at < $3 
	AND bookings.id is NULL 
	ORDER BY slots.start_at
`
	rows, err := sr.db.QueryContext(ctx, query, roomId, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var slots []models.Slot
	for rows.Next() {
		var slot models.Slot
		err := rows.Scan(&slot.ID, &slot.RoomID, &slot.StartAt, &slot.EndAt)
		if err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return slots, nil
}
func (sr *slotsRepository) ExistsForRoomAndDate(ctx context.Context, roomID string, date time.Time) (bool, error) {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, 0, 1)

	query := `
		SELECT EXISTS (
			SELECT 1 FROM slots
			WHERE room_id = $1 AND start_at >= $2 AND start_at < $3)
	`
	var exists bool
	err := sr.db.QueryRowContext(ctx, query, roomID, dayStart, dayEnd).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
func (sr *slotsRepository) SlotsBulkCreate(ctx context.Context, slots []models.Slot) error {
	if len(slots) == 0 {
		return nil
	}
	var queryBuilder strings.Builder
	var args []any
	queryBuilder.WriteString(`INSERT INTO slots (id, room_id, start_at, end_at) VALUES`)
	for i, v := range slots {
		if i > 0 {
			queryBuilder.WriteString(",")
		}
		queryBuilder.WriteString(fmt.Sprintf(" ($%d,$%d,$%d,$%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		args = append(args, v.ID, v.RoomID, v.StartAt, v.EndAt)
	}
	queryBuilder.WriteString(`ON CONFLICT (room_id, start_at, end_at) DO NOTHING`)
	_, err := sr.dbtx(ctx).ExecContext(ctx, queryBuilder.String(), args...)
	return err
}
func (sr *slotsRepository) dbtx(ctx context.Context) transactions.DBTX {
	if tx, ok := transactions.TxFromContext(ctx); ok {
		return tx
	}
	return sr.db
}
