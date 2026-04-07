package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestIntegration_CreateRoomScheduleAndBooking(t *testing.T) {
	_, server := setupIntegrationTest(t)

	registerUser(t, server.URL, "user3@example.com", "123456", "user")

	adminToken := dummyLogin(t, server.URL, "admin")
	userToken := login(t, server.URL, "user3@example.com", "123456")

	client := &http.Client{}

	roomResp := doJSONRequest(t, client, http.MethodPost, server.URL+"/rooms/create", adminToken, map[string]any{
		"name":        "Blue room",
		"description": "big room",
		"capacity":    10,
	})
	defer roomResp.Body.Close()
	if roomResp.StatusCode != http.StatusCreated {
		t.Fatalf("create room expected 201, got %d", roomResp.StatusCode)
	}
	var roomOut struct {
		Room struct {
			Id string `json:"id"`
		} `json:"room"`
	}
	if err := json.NewDecoder(roomResp.Body).Decode(&roomOut); err != nil {
		t.Fatalf("decode room: %v", err)
	}
	if roomOut.Room.Id == "" {
		t.Fatal("empty room id")
	}
	scheduleResp := doJSONRequest(
		t,
		client,
		http.MethodPost,
		server.URL+"/rooms/"+roomOut.Room.Id+"/schedule/create",
		adminToken,
		map[string]any{
			"daysOfWeek": []int{1, 2, 3, 4, 5},
			"startTime":  "09:00",
			"endTime":    "10:00",
		},
	)
	defer scheduleResp.Body.Close()

	if scheduleResp.StatusCode != http.StatusCreated {
		t.Fatalf("create schedule expected 201, got %d", scheduleResp.StatusCode)
	}
	date := nextWeekday(time.Monday)
	slotsResp := doJSONRequest(
		t,
		client,
		http.MethodGet,
		server.URL+"/rooms/"+roomOut.Room.Id+"/slots/list?date="+date,
		userToken,
		nil,
	)
	defer slotsResp.Body.Close()
	if slotsResp.StatusCode != http.StatusOK {
		t.Fatalf("get slots expected 200, got %d", slotsResp.StatusCode)
	}
	var slotsOut struct {
		Slots []struct {
			Id string `json:"id"`
		} `json:"slots"`
	}
	if err := json.NewDecoder(slotsResp.Body).Decode(&slotsOut); err != nil {
		t.Fatalf("decode slots: %v", err)
	}
	if len(slotsOut.Slots) == 0 {
		t.Fatal("expected non-empty slots")
	}
	bookingResp := doJSONRequest(
		t,
		client,
		http.MethodPost,
		server.URL+"/bookings/create",
		userToken,
		map[string]any{
			"slotId": slotsOut.Slots[0].Id,
		},
	)
	defer bookingResp.Body.Close()
	if bookingResp.StatusCode != http.StatusCreated {
		t.Fatalf("create booking expected 201, got %d", bookingResp.StatusCode)
	}
	var bookingOut struct {
		Booking struct {
			Id     string `json:"id"`
			SlotId string `json:"slotId"`
			Status string `json:"status"`
		} `json:"booking"`
	}
	if err := json.NewDecoder(bookingResp.Body).Decode(&bookingOut); err != nil {
		t.Fatalf("decode booking: %v", err)
	}
	if bookingOut.Booking.Id == "" {
		t.Fatal("empty booking id")
	}
	if bookingOut.Booking.SlotId != slotsOut.Slots[0].Id {
		t.Fatalf("expected slotId %s, got %s", slotsOut.Slots[0].Id, bookingOut.Booking.SlotId)
	}
}
