package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/avito-internships/test-backend-1-untrik/internal/models"
	"github.com/avito-internships/test-backend-1-untrik/internal/transactions"
	"github.com/jackc/pgx/v5/pgconn"
)

type schedulesRepository struct {
	db *sql.DB
}

func NewSchedulesRepository(db *sql.DB) *schedulesRepository {
	return &schedulesRepository{db: db}
}

var ErrScheduleNotFound = errors.New("schedule not found")
var ErrScheduleExists = errors.New("schedule already exists")

func (sr *schedulesRepository) CreateSchedule(ctx context.Context, schedule models.Schedule) error {
	query := `
		INSERT INTO schedules (id,room_id,days_of_week,start_time,end_time) VALUES($1,$2,$3,$4,$5)
	`
	_, err := sr.dbtx(ctx).ExecContext(ctx, query, schedule.ID, schedule.RoomID, schedule.DaysOfWeek,
		schedule.StartTime, schedule.EndTime)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrScheduleExists
		}
		return err
	}
	return nil
}
func (sr *schedulesRepository) GetByRoomID(ctx context.Context, roomID string) (models.Schedule, error) {
	query := `
		SELECT id,room_id,days_of_week,start_time,end_time FROM schedules
		WHERE room_id = $1
	`
	var schedule models.Schedule
	var daysRaw string
	if err := sr.db.QueryRowContext(ctx, query, roomID).Scan(&schedule.ID, &schedule.RoomID, &daysRaw,
		&schedule.StartTime, &schedule.EndTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Schedule{}, ErrScheduleNotFound
		}
		return models.Schedule{}, err
	}
	days, err := parsePGIntArray(daysRaw)
	if err != nil {
		return models.Schedule{}, err
	}
	schedule.DaysOfWeek = days
	return schedule, nil

}
func (sr *schedulesRepository) dbtx(ctx context.Context) transactions.DBTX {
	if tx, ok := transactions.TxFromContext(ctx); ok {
		return tx
	}
	return sr.db
}
func parsePGIntArray(raw string) ([]int, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "{")
	raw = strings.TrimSuffix(raw, "}")

	if raw == "" {
		return []int{}, nil
	}

	parts := strings.Split(raw, ",")
	result := make([]int, 0, len(parts))

	for _, v := range parts {
		v = strings.TrimSpace(v)

		num, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("parse postgres int array: %w", err)
		}

		result = append(result, num)
	}

	return result, nil
}
