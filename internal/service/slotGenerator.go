package service

import (
	"time"

	"github.com/avito-internships/test-backend-1-untrik/internal/models"
	"github.com/google/uuid"
)

func toScheduleWeekday(t time.Time) int {
	if t.Weekday() == time.Sunday {
		return 7
	}
	return int(t.Weekday())
}
func containsDay(days []int, target int) bool {
	for _, v := range days {
		if v == target {
			return true
		}
	}
	return false
}
func GenerateSlotsForRange(roomId string, daysOfWeek []int, startTime, endTime time.Time, from time.Time, days int) []models.Slot {
	startDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, days)

	allowedDays := make(map[int]struct{}, len(daysOfWeek))
	for _, day := range daysOfWeek {
		allowedDays[day] = struct{}{}
	}
	startMinutes := startTime.Hour()*60 + startTime.Minute()
	endMinutes := endTime.Hour()*60 + endTime.Minute()
	var slots []models.Slot

	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		weekday := toScheduleWeekday(d)
		if _, ok := allowedDays[weekday]; !ok {
			continue
		}
		for m := startMinutes; m+30 <= endMinutes; m += 30 {
			slotStart := d.Add(time.Duration(m) * time.Minute)
			slotEnd := slotStart.Add(30 * time.Minute)
			slot := models.Slot{ID: uuid.NewString(), RoomID: roomId, StartAt: slotStart, EndAt: slotEnd}
			slots = append(slots, slot)
		}
	}
	return slots
}
func GenerateForDate(roomId string, startTime, endTime time.Time, date time.Time) []models.Slot {
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	var slots []models.Slot
	startMinutes := startTime.Hour()*60 + startTime.Minute()
	endMinutes := endTime.Hour()*60 + endTime.Minute()
	for m := startMinutes; m+30 <= endMinutes; m += 30 {
		slotStart := startDate.Add(time.Duration(m) * time.Minute)
		slotEnd := slotStart.Add(30 * time.Minute)
		slot := models.Slot{ID: uuid.NewString(), RoomID: roomId, StartAt: slotStart, EndAt: slotEnd}
		slots = append(slots, slot)
	}

	return slots
}
