package handler

import (
	"database/sql"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"schedule-app/internal/db"
	"schedule-app/internal/middleware"
	"schedule-app/internal/repository"
	"testing"
)

// testServer holds dependencies for a test server.
type testServer struct {
	router http.Handler
	db     *sql.DB
}

// newTestServer creates a new server for testing, with a fresh in-memory SQLite DB.
func newTestServer() *testServer {
	// Use in-memory SQLite database for testing.
	// Pass ":memory:" to InitDB, which will construct the correct DSN.
	conn, err := db.InitDB(":memory:")
	if err != nil {
		log.Fatalf("Failed to initialize in-memory database: %v", err)
	}

	// Create repositories and handlers
	jwtSecretForTest := "test_secret_key_for_unit_tests"
	userRepo := repository.NewUserRepository(conn)
	userHandler := NewUserHandler(userRepo, jwtSecretForTest)
	scheduleRepo := repository.NewScheduleRepository(conn)
	scheduleHandler := NewScheduleHandler(scheduleRepo)
	authMiddleware := middleware.NewAuthMiddleware(jwtSecretForTest)

	// Set up router
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/users/register", userHandler.Register)
	mux.HandleFunc("POST /api/users/login", userHandler.Login)
	mux.Handle("POST /api/schedules", authMiddleware.JwtAuthentication(http.HandlerFunc(scheduleHandler.CreateSchedule)))
	mux.HandleFunc("GET /api/users/{ownerID}/schedules", scheduleHandler.GetSchedulesByOwner)
	mux.HandleFunc("GET /api/schedules/{scheduleID}", scheduleHandler.GetScheduleByID)
	mux.Handle("PUT /api/schedules/{scheduleID}", authMiddleware.JwtAuthentication(http.HandlerFunc(scheduleHandler.UpdateSchedule)))
	mux.Handle("DELETE /api/schedules/{scheduleID}", authMiddleware.JwtAuthentication(http.HandlerFunc(scheduleHandler.DeleteSchedule)))

	return &testServer{
		router: mux,
		db:     conn,
	}
}

// executeRequest performs a request against the test server's router.
func (ts *testServer) executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)
	return rr
}

// TestMain provides a place for package-level setup/teardown, if needed.
func TestMain(m *testing.M) {
	// Run all tests
	code := m.Run()
	os.Exit(code)
}