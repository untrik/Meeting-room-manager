package dto

type ScheduleResponse struct {
	Id         string `json:"id"`
	RoomId     string `json:"roomId"`
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}
