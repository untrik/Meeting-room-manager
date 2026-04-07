package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestIntegration_CancelBooking(t *testing.T) {
	_, server := setupIntegrationTest(t)

	registerUser(t, server.URL, "user4@example.com", "123456", "user")

	adminToken := dummyLogin(t, server.URL, "admin")
	userToken := login(t, server.URL, "user4@example.com", "123456")

	client := &http.Client{}

	roomResp := doJSONRequest(t, client, http.MethodPost, server.URL+"/rooms/create", adminToken, map[string]any{
		"name":        "Green room",
		"description": "small room",
		"capacity":    6,
	})
	defer roomResp.Body.Close()

	var roomOut struct {
		Room struct {
			Id string `json:"id"`
		} `json:"room"`
	}
	if err := json.NewDecoder(roomResp.Body).Decode(&roomOut); err != nil {
		t.Fatalf("decode room: %v", err)
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
			Id string `json:"id"`
		} `json:"booking"`
	}
	if err := json.NewDecoder(bookingResp.Body).Decode(&bookingOut); err != nil {
		t.Fatalf("decode booking: %v", err)
	}
	if bookingOut.Booking.Id == "" {
		t.Fatal("empty booking id")
	}

	cancelResp := doJSONRequest(
		t,
		client,
		http.MethodPost,
		server.URL+"/bookings/"+bookingOut.Booking.Id+"/cancel",
		userToken,
		nil,
	)
	defer cancelResp.Body.Close()

	if cancelResp.StatusCode != http.StatusOK {
		t.Fatalf("cancel booking expected 200, got %d", cancelResp.StatusCode)
	}

	var cancelOut struct {
		Booking struct {
			Id     string `json:"id"`
			Status string `json:"status"`
		} `json:"booking"`
	}
	if err := json.NewDecoder(cancelResp.Body).Decode(&cancelOut); err != nil {
		t.Fatalf("decode cancelled booking: %v", err)
	}

	if cancelOut.Booking.Id != bookingOut.Booking.Id {
		t.Fatalf("expected booking id %s, got %s", bookingOut.Booking.Id, cancelOut.Booking.Id)
	}
	if cancelOut.Booking.Status == "" {
		t.Fatal("expected non-empty booking status")
	}
}
