package dto

type ErrorCode string

const (
	ErrorCodeInvalidRequest    ErrorCode = "INVALID_REQUEST"
	ErrorCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrorCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrorCodeRoomNotFound      ErrorCode = "ROOM_NOT_FOUND"
	ErrorCodeSlotNotFound      ErrorCode = "SLOT_NOT_FOUND"
	ErrorCodeSlotAlreadyBooked ErrorCode = "SLOT_ALREADY_BOOKED"
	ErrorCodeBookingNotFound   ErrorCode = "BOOKING_NOT_FOUND"
	ErrorCodeForbidden         ErrorCode = "FORBIDDEN"
	ErrorCodeScheduleExists    ErrorCode = "SCHEDULE_EXISTS"
)

type InternalError string

const (
	ErrorCodeInternalError InternalError = "INTERNAL_ERROR"
)

type ErrorResponse struct {
	Error struct {
		Code    ErrorCode `json:"code"`
		Message string    `json:"message"`
	} `json:"error"`
}
type InternalErrorResponse struct {
	Error struct {
		Code    InternalError `json:"code"`
		Message string        `json:"message"`
	} `json:"error"`
}
