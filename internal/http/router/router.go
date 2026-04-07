package router

import (
	"net/http"

	"github.com/avito-internships/test-backend-1-untrik/internal/http/helper"
	"github.com/avito-internships/test-backend-1-untrik/internal/http/middleware"
	auth "github.com/avito-internships/test-backend-1-untrik/internal/jwt"
)

type authHandlerInterface interface {
	Login(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	DummyLogin(w http.ResponseWriter, r *http.Request)
}

type roomHandlerInterface interface {
	CreateRoom(w http.ResponseWriter, r *http.Request)
	GetListRooms(w http.ResponseWriter, r *http.Request)
}
type jWTServiceInterface interface {
	ParseToken(tokenString string) (*auth.Claims, error)
}
type scheduleHandlerInterface interface {
	CreateSchedule(w http.ResponseWriter, r *http.Request)
}
type slotsHandlerInterface interface {
	GetAvailableSlots(w http.ResponseWriter, r *http.Request)
}

type bookingsHandlerInterface interface {
	CreateBooking(w http.ResponseWriter, r *http.Request)
	GetMyBooking(w http.ResponseWriter, r *http.Request)
	CancelBooking(w http.ResponseWriter, r *http.Request)
	GetBookingsList(w http.ResponseWriter, r *http.Request)
}

func NewRouter(authHandler authHandlerInterface,
	roomHandler roomHandlerInterface,
	scheduleHandler scheduleHandlerInterface,
	slotsHandler slotsHandlerInterface,
	bookingsHandler bookingsHandlerInterface,
	jwtService jWTServiceInterface) http.Handler {
	authMw := middleware.AuthMiddleware(jwtService)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /_info", func(w http.ResponseWriter, r *http.Request) {
		helper.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /dummyLogin", authHandler.DummyLogin)

	mux.Handle("POST /rooms/create", authMw(middleware.IsAdminMiddleware(http.HandlerFunc(roomHandler.CreateRoom))))
	mux.Handle("GET /rooms/list", authMw(http.HandlerFunc(roomHandler.GetListRooms)))

	mux.Handle("POST /rooms/{roomId}/schedule/create", authMw(middleware.IsAdminMiddleware(http.HandlerFunc(scheduleHandler.CreateSchedule))))

	mux.Handle("GET /rooms/{roomId}/slots/list", authMw(http.HandlerFunc(slotsHandler.GetAvailableSlots)))

	mux.Handle("POST /bookings/create", authMw(middleware.IsUserMiddleware(http.HandlerFunc(bookingsHandler.CreateBooking))))
	mux.Handle("GET /bookings/list", authMw(middleware.IsAdminMiddleware(http.HandlerFunc(bookingsHandler.GetBookingsList))))
	mux.Handle("GET /bookings/my", authMw(middleware.IsUserMiddleware(http.HandlerFunc(bookingsHandler.GetMyBooking))))
	mux.Handle("POST /bookings/{bookingId}/cancel", authMw(middleware.IsUserMiddleware(http.HandlerFunc(bookingsHandler.CancelBooking))))

	return mux
}
