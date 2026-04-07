package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/avito-internships/test-backend-1-untrik/internal/dto"
	"github.com/avito-internships/test-backend-1-untrik/internal/models"
	"github.com/google/uuid"
)

var ErrInvalidData = errors.New("invalid data")

type RoomsRepositoryInterface interface {
	GetAllRooms(ctx context.Context) ([]models.Room, error)
	CreateRoom(ctx context.Context, room models.Room) (time.Time, error)
}

type roomsService struct {
	roomsRepository RoomsRepositoryInterface
}

func NewRoomsService(roomsRepository RoomsRepositoryInterface) *roomsService {
	return &roomsService{roomsRepository: roomsRepository}
}
func (rs *roomsService) GetListRooms(ctx context.Context) ([]dto.RoomResponse, error) {
	roomsResponse := make([]dto.RoomResponse, 0)
	rooms, err := rs.roomsRepository.GetAllRooms(ctx)
	if err != nil {
		slog.Error("failed to get rooms", "error", err)
		return nil, err
	}
	for _, v := range rooms {
		roomResponse := dto.RoomResponse{Id: v.ID, Name: v.Name,
			Description: v.Description, Capacity: v.Capacity, CreatedAt: &v.CreatedAt}
		roomsResponse = append(roomsResponse, roomResponse)
	}
	return roomsResponse, nil
}
func (rs *roomsService) CreateRoom(ctx context.Context, name string, description *string, capacity *int) (dto.RoomResponse, error) {
	if name == "" {
		slog.Info("name in empthy")
		return dto.RoomResponse{}, ErrInvalidData
	}
	if capacity == nil || *capacity < 0 {
		slog.Info("invalid capacity")
		return dto.RoomResponse{}, ErrInvalidData
	}
	id := uuid.New()
	room := models.Room{ID: id.String(), Name: name, Description: description, Capacity: capacity}
	createdAt, err := rs.roomsRepository.CreateRoom(ctx, room)
	if err != nil {
		slog.Error("failed to create room", "error", err)
		return dto.RoomResponse{}, err
	}
	return dto.RoomResponse{Id: room.ID, Name: room.Name, Description: room.Description, Capacity: room.Capacity, CreatedAt: &createdAt}, nil
}
