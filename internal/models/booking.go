package models

import "time"

type Booking struct {
	ID             string
	SlotID         string
	UserID         string
	StatusID       int16
	ConferenceLink *string
	CreatedAt      time.Time
}
