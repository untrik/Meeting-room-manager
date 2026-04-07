package service

import (
	"testing"
	"time"
)

func TestToScheduleWeekday(t *testing.T) {
	t.Run("monday becomes 1", func(t *testing.T) {
		d := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
		got := toScheduleWeekday(d)

		if got != 1 {
			t.Fatalf("expected 1, got %d", got)
		}
	})

	t.Run("sunday becomes 7", func(t *testing.T) {
		d := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)
		got := toScheduleWeekday(d)

		if got != 7 {
			t.Fatalf("expected 7, got %d", got)
		}
	})
}

func TestContainsDay(t *testing.T) {
	t.Run("contains target", func(t *testing.T) {
		got := containsDay([]int{1, 3, 5}, 3)

		if !got {
			t.Fatal("expected true, got false")
		}
	})

	t.Run("does not contain target", func(t *testing.T) {
		got := containsDay([]int{1, 3, 5}, 2)

		if got {
			t.Fatal("expected false, got true")
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		got := containsDay([]int{}, 1)

		if got {
			t.Fatal("expected false, got true")
		}
	})
}

func TestGenerateForDate(t *testing.T) {
	t.Run("generates two slots for one hour", func(t *testing.T) {
		roomID := "room-1"
		startTime := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
		endTime := time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)
		date := time.Date(2026, 3, 23, 15, 45, 0, 0, time.UTC)

		got := GenerateForDate(roomID, startTime, endTime, date)

		if len(got) != 2 {
			t.Fatalf("expected 2 slots, got %d", len(got))
		}

		if got[0].RoomID != roomID || got[1].RoomID != roomID {
			t.Fatalf("expected roomID %s, got %+v", roomID, got)
		}

		wantStart1 := time.Date(2026, 3, 23, 9, 0, 0, 0, time.UTC)
		wantEnd1 := time.Date(2026, 3, 23, 9, 30, 0, 0, time.UTC)
		wantStart2 := time.Date(2026, 3, 23, 9, 30, 0, 0, time.UTC)
		wantEnd2 := time.Date(2026, 3, 23, 10, 0, 0, 0, time.UTC)

		if !got[0].StartAt.Equal(wantStart1) || !got[0].EndAt.Equal(wantEnd1) {
			t.Fatalf("unexpected first slot: %+v", got[0])
		}
		if !got[1].StartAt.Equal(wantStart2) || !got[1].EndAt.Equal(wantEnd2) {
			t.Fatalf("unexpected second slot: %+v", got[1])
		}

		if got[0].ID == "" || got[1].ID == "" {
			t.Fatal("expected generated slot IDs")
		}
	})

	t.Run("generates zero slots when interval is less than 30 minutes", func(t *testing.T) {
		startTime := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
		endTime := time.Date(0, 1, 1, 9, 20, 0, 0, time.UTC)
		date := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
		got := GenerateForDate("room-1", startTime, endTime, date)
		if len(got) != 0 {
			t.Fatalf("expected 0 slots, got %d", len(got))
		}
	})

	t.Run("normalizes date to midnight", func(t *testing.T) {
		startTime := time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC)
		endTime := time.Date(0, 1, 1, 8, 30, 0, 0, time.UTC)
		date := time.Date(2026, 3, 23, 22, 10, 55, 0, time.UTC)
		got := GenerateForDate("room-1", startTime, endTime, date)
		if len(got) != 1 {
			t.Fatalf("expected 1 slot, got %d", len(got))
		}
		wantStart := time.Date(2026, 3, 23, 8, 0, 0, 0, time.UTC)
		if !got[0].StartAt.Equal(wantStart) {
			t.Fatalf("expected slot start %v, got %v", wantStart, got[0].StartAt)
		}
	})
}

func TestGenerateSlotsForRange(t *testing.T) {
	t.Run("generates slots only for allowed weekdays", func(t *testing.T) {
		roomID := "room-1"
		daysOfWeek := []int{1, 3}
		startTime := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
		endTime := time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)
		from := time.Date(2026, 3, 23, 12, 0, 0, 0, time.UTC)
		days := 3

		got := GenerateSlotsForRange(roomID, daysOfWeek, startTime, endTime, from, days)
		if len(got) != 4 {
			t.Fatalf("expected 4 slots, got %d", len(got))
		}
		wantStarts := []time.Time{
			time.Date(2026, 3, 23, 9, 0, 0, 0, time.UTC),
			time.Date(2026, 3, 23, 9, 30, 0, 0, time.UTC),
			time.Date(2026, 3, 25, 9, 0, 0, 0, time.UTC),
			time.Date(2026, 3, 25, 9, 30, 0, 0, time.UTC),
		}
		for i, slot := range got {
			if slot.RoomID != roomID {
				t.Fatalf("expected roomID %s, got %s", roomID, slot.RoomID)
			}
			if slot.ID == "" {
				t.Fatal("expected generated slot ID")
			}
			if !slot.StartAt.Equal(wantStarts[i]) {
				t.Fatalf("slot %d: expected start %v, got %v", i, wantStarts[i], slot.StartAt)
			}
		}
	})
	t.Run("returns empty when no allowed days in range", func(t *testing.T) {
		startTime := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
		endTime := time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)
		from := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)

		got := GenerateSlotsForRange("room-1", []int{7}, startTime, endTime, from, 2)

		if len(got) != 0 {
			t.Fatalf("expected 0 slots, got %d", len(got))
		}
	})
	t.Run("returns empty when interval is less than 30 minutes", func(t *testing.T) {
		startTime := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
		endTime := time.Date(0, 1, 1, 9, 20, 0, 0, time.UTC)
		from := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
		got := GenerateSlotsForRange("room-1", []int{1}, startTime, endTime, from, 1)
		if len(got) != 0 {
			t.Fatalf("expected 0 slots, got %d", len(got))
		}
	})
	t.Run("normalizes from date to midnight", func(t *testing.T) {
		startTime := time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC)
		endTime := time.Date(0, 1, 1, 8, 30, 0, 0, time.UTC)
		from := time.Date(2026, 3, 23, 18, 45, 10, 0, time.UTC)
		got := GenerateSlotsForRange("room-1", []int{1}, startTime, endTime, from, 1)

		if len(got) != 1 {
			t.Fatalf("expected 1 slot, got %d", len(got))
		}

		wantStart := time.Date(2026, 3, 23, 8, 0, 0, 0, time.UTC)
		if !got[0].StartAt.Equal(wantStart) {
			t.Fatalf("expected start %v, got %v", wantStart, got[0].StartAt)
		}
	})
}
