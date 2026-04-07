package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/avito-internships/test-backend-1-untrik/internal/config"
	auth "github.com/avito-internships/test-backend-1-untrik/internal/jwt"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("error open db: %w", err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("error ping db: %w", err)
	}
	jwtService := auth.NewJWTService(*cfg)

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           BuildHandler(db, jwtService),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return server.ListenAndServe()
}
