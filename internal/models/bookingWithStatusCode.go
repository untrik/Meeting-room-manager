package models

import (
	"time"
)

type BookingWithStatusCode struct {
	ID             string
	SlotID         string
	StatusCode     string
	UserID         string
	ConferenceLink *string
	CreatedAt      time.Time
}
