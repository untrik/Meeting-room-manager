package models

import "time"

type Slot struct {
	ID      string
	RoomID  string
	StartAt time.Time
	EndAt   time.Time
}
