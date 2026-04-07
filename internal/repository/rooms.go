package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/avito-internships/test-backend-1-untrik/internal/models"
)

var ErrRoomExists = errors.New("room already exists")
var ErrRoomNotFound = errors.New("room not found")

type roomsRepository struct {
	db *sql.DB
}

func NewRoomsRepository(db *sql.DB) *roomsRepository {
	return &roomsRepository{db: db}
}
func (rr *roomsRepository) CreateRoom(ctx context.Context, room models.Room) (time.Time, error) {
	query := `
		INSERT INTO rooms (id,name,description,capacity,created_at) VALUES($1,$2,$3,$4,NOW())
		RETURNING created_at
`
	var createdAt time.Time
	err := rr.db.QueryRowContext(ctx, query, room.ID, room.Name, room.Description, room.Capacity).Scan(&createdAt)
	if err != nil {
		return createdAt, err
	}
	return createdAt, nil
}
func (rr *roomsRepository) GetAllRooms(ctx context.Context) ([]models.Room, error) {
	query := `
		SELECT id,name,description,capacity,created_at FROM rooms
`
	var rooms []models.Room
	rows, err := rr.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var room models.Room
		if err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rooms, nil
}
func (rr *roomsRepository) ExistsRoom(ctx context.Context, roomId string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM rooms WHERE id = $1
		)
`
	var exists bool
	err := rr.db.QueryRowContext(ctx, query, roomId).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
