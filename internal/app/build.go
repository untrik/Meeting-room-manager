package app

import (
	"database/sql"
	"net/http"

	"github.com/avito-internships/test-backend-1-untrik/internal/http/handlers"
	"github.com/avito-internships/test-backend-1-untrik/internal/http/router"
	auth "github.com/avito-internships/test-backend-1-untrik/internal/jwt"
	"github.com/avito-internships/test-backend-1-untrik/internal/repository"
	"github.com/avito-internships/test-backend-1-untrik/internal/service"
	"github.com/avito-internships/test-backend-1-untrik/internal/transactions"
)

func BuildHandler(db *sql.DB, jwtService *auth.JWTService) http.Handler {
	txManager := transactions.NewSQLTxManager(db)

	bookingsRepo := repository.NewBookingsRepository(db)
	userRepo := repository.NewUserRepository(db)
	roomsRepo := repository.NewRoomsRepository(db)
	schedulesRepo := repository.NewSchedulesRepository(db)
	slotsRepo := repository.NewSlotsRepository(db)

	bookingsService := service.NewBookingsService(bookingsRepo)
	authService := service.NewAuthService(userRepo, jwtService)
	roomsService := service.NewRoomsService(roomsRepo)
	schedulesService := service.NewSchedulesService(schedulesRepo, slotsRepo, roomsRepo, txManager)
	slotsService := service.NewSlotsService(schedulesRepo, slotsRepo, roomsRepo)

	bookingsHandler := handlers.NewBookingHandler(bookingsService)
	roomsHandler := handlers.NewRoomHandler(roomsService)
	authHandler := handlers.NewAuthHandler(authService)
	schedulesHandler := handlers.NewSchedulerHandler(schedulesService)
	slotsHandler := handlers.NewSlotsHandler(slotsService)

	return router.NewRouter(authHandler, roomsHandler, schedulesHandler, slotsHandler,
		bookingsHandler, jwtService)
}
