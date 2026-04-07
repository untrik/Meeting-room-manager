package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-untrik/internal/dto"
	"github.com/avito-internships/test-backend-1-untrik/internal/models"
)

type fakeRoomsRepo struct {
	getAllRoomsFn func(ctx context.Context) ([]models.Room, error)
	createRoomFn  func(ctx context.Context, room models.Room) (time.Time, error)
}

func (f *fakeRoomsRepo) GetAllRooms(ctx context.Context) ([]models.Room, error) {
	return f.getAllRoomsFn(ctx)
}

func (f *fakeRoomsRepo) CreateRoom(ctx context.Context, room models.Room) (time.Time, error) {
	return f.createRoomFn(ctx, room)
}

func TestRoomsService_GetListRooms(t *testing.T) {
	ctx := context.Background()
	t.Run("repository error", func(t *testing.T) {
		repo := &fakeRoomsRepo{
			getAllRoomsFn: func(ctx context.Context) ([]models.Room, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewRoomsService(repo)
		_, err := svc.GetListRooms(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
	t.Run("success", func(t *testing.T) {
		now := time.Now()
		description := "big room"
		capacity := 10
		repo := &fakeRoomsRepo{
			getAllRoomsFn: func(ctx context.Context) ([]models.Room, error) {
				return []models.Room{
					{
						ID:          "room-1",
						Name:        "Blue room",
						Description: &description,
						Capacity:    &capacity,
						CreatedAt:   now,
					},
				}, nil
			},
		}
		svc := NewRoomsService(repo)
		got, err := svc.GetListRooms(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []dto.RoomResponse{
			{Id: "room-1", Name: "Blue room", Description: &description, Capacity: &capacity, CreatedAt: &now},
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %+v, got %+v", want, got)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		repo := &fakeRoomsRepo{
			getAllRoomsFn: func(ctx context.Context) ([]models.Room, error) {
				return []models.Room{}, nil
			},
		}
		svc := NewRoomsService(repo)

		got, err := svc.GetListRooms(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("expected empty slice, got %+v", got)
		}
	})
}

func TestRoomsService_CreateRoom(t *testing.T) {
	ctx := context.Background()
	t.Run("empty name", func(t *testing.T) {
		repo := &fakeRoomsRepo{}
		svc := NewRoomsService(repo)
		capacity := 10
		_, err := svc.CreateRoom(ctx, "", nil, &capacity)
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})
	t.Run("nil capacity", func(t *testing.T) {
		repo := &fakeRoomsRepo{}
		svc := NewRoomsService(repo)
		_, err := svc.CreateRoom(ctx, "Room 1", nil, nil)
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})
	t.Run("negative capacity", func(t *testing.T) {
		repo := &fakeRoomsRepo{}
		svc := NewRoomsService(repo)
		capacity := -1
		_, err := svc.CreateRoom(ctx, "Room 1", nil, &capacity)
		if !errors.Is(err, ErrInvalidData) {
			t.Fatalf("expected ErrInvalidData, got %v", err)
		}
	})
	t.Run("repository error", func(t *testing.T) {
		repo := &fakeRoomsRepo{
			createRoomFn: func(ctx context.Context, room models.Room) (time.Time, error) {
				return time.Time{}, errors.New("db error")
			},
		}
		svc := NewRoomsService(repo)
		capacity := 8
		_, err := svc.CreateRoom(ctx, "Room 1", nil, &capacity)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
	t.Run("success", func(t *testing.T) {
		description := "small meeting room"
		capacity := 6
		createdAt := time.Now()

		var savedRoom models.Room

		repo := &fakeRoomsRepo{
			createRoomFn: func(ctx context.Context, room models.Room) (time.Time, error) {
				savedRoom = room
				return createdAt, nil
			},
		}
		svc := NewRoomsService(repo)

		resp, err := svc.CreateRoom(ctx, "Room 1", &description, &capacity)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if savedRoom.ID == "" {
			t.Fatal("expected generated room ID, got empty string")
		}
		if savedRoom.Name != "Room 1" {
			t.Fatalf("expected name Room 1, got %s", savedRoom.Name)
		}
		if savedRoom.Description == nil || *savedRoom.Description != description {
			t.Fatalf("expected description %s, got %+v", description, savedRoom.Description)
		}
		if savedRoom.Capacity == nil || *savedRoom.Capacity != capacity {
			t.Fatalf("expected capacity %d, got %+v", capacity, savedRoom.Capacity)
		}

		if resp.Id != savedRoom.ID {
			t.Fatalf("expected response id %s, got %s", savedRoom.ID, resp.Id)
		}
		if resp.Name != "Room 1" {
			t.Fatalf("expected response name Room 1, got %s", resp.Name)
		}
		if resp.Description == nil || *resp.Description != description {
			t.Fatalf("expected response description %s, got %+v", description, resp.Description)
		}
		if resp.Capacity == nil || *resp.Capacity != capacity {
			t.Fatalf("expected response capacity %d, got %+v", capacity, resp.Capacity)
		}
		if resp.CreatedAt == nil || !resp.CreatedAt.Equal(createdAt) {
			t.Fatalf("expected createdAt %v, got %+v", createdAt, resp.CreatedAt)
		}
	})
}
