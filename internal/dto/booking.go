package dto

import "time"

type BookingResponse struct {
	Id             string     `json:"id"`
	SlotId         string     `json:"slotId"`
	UserId         string     `json:"userId"`
	Status         string     `json:"status"`
	ConferenceLink *string    `json:"conferenceLink,omitempty"`
	CreatedAt      *time.Time `json:"createdAt,omitempty"`
}
