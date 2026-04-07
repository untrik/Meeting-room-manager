package dto

import "time"

type SlotResponse struct {
	Id     string    `json:"id"`
	RoomID string    `json:"roomId"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}
