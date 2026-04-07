package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-untrik/internal/app"
	"github.com/avito-internships/test-backend-1-untrik/internal/config"
	auth "github.com/avito-internships/test-backend-1-untrik/internal/jwt"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func setupIntegrationTest(t *testing.T) (*sql.DB, *httptest.Server) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/room_booking_test?sslmode=disable"
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}
	cleanDB(t, db)

	cfg := config.Config{
		JWTSecret: "test-secret",
	}
	jwtService := auth.NewJWTService(cfg)

	handler := app.BuildHandler(db, jwtService)
	server := httptest.NewServer(handler)

	t.Cleanup(func() {
		server.Close()
		db.Close()
	})
	return db, server
}

func cleanDB(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
		TRUNCATE bookings, slots, schedules, rooms
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("clean db: %v", err)
	}
}

func doJSONRequest(t *testing.T, client *http.Client, method, url, token string, body any) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}
func registerUser(t *testing.T, baseURL, email, password, role string) {
	t.Helper()
	client := &http.Client{}
	resp := doJSONRequest(t, client, http.MethodPost, baseURL+"/register", "", map[string]any{
		"email":    email,
		"password": password,
		"role":     role,
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register expected 201, got %d", resp.StatusCode)
	}
}
func dummyLogin(t *testing.T, baseURL, role string) string {
	t.Helper()
	client := &http.Client{}
	resp := doJSONRequest(t, client, http.MethodPost, baseURL+"/dummyLogin", "", map[string]any{
		"role": role,
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("dummyLogin expected 200, got %d", resp.StatusCode)
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode token: %v", err)
	}
	if out.Token == "" {
		t.Fatal("empty token")
	}
	return out.Token
}
func login(t *testing.T, baseURL, email, password string) string {
	t.Helper()
	client := &http.Client{}
	resp := doJSONRequest(t, client, http.MethodPost, baseURL+"/login", "", map[string]any{
		"email":    email,
		"password": password,
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login expected 200, got %d", resp.StatusCode)
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode token: %v", err)
	}
	if out.Token == "" {
		t.Fatal("empty token")
	}
	return out.Token
}
func nextWeekday(target time.Weekday) string {
	now := time.Now().UTC()
	for i := 0; i < 7; i++ {
		d := now.AddDate(0, 0, i)
		if d.Weekday() == target {
			return d.Format("2006-01-02")
		}
	}
	return now.Format("2006-01-02")
}
