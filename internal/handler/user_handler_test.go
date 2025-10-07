package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"schedule-app/internal/model"
	"testing"
)

func TestUserHandlers(t *testing.T) {
	// --- Test Setup ---
	server := newTestServer()
	defer server.db.Close()

	var user model.UserResponse
	var token string

	// --- Test Cases ---
	t.Run("Should register a new user successfully", func(t *testing.T) {
		requestBody := `{"username": "testuser", "email": "test@example.com", "password": "password123"}`
		req, _ := http.NewRequest("POST", "/api/users/register", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		rr := server.executeRequest(req)

		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
			t.Errorf("response body: %s", rr.Body.String())
		}

		// Decode the response to verify its structure and save user for next test
		if err := json.NewDecoder(rr.Body).Decode(&user); err != nil {
			t.Fatalf("Could not decode response: %v", err)
		}
		if user.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got '%s'", user.Username)
		}
	})

	t.Run("Should login the registered user successfully", func(t *testing.T) {
		requestBody := `{"email": "test@example.com", "password": "password123"}`
		req, _ := http.NewRequest("POST", "/api/users/login", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		rr := server.executeRequest(req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		var resp map[string]string
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("Could not decode response: %v", err)
		}

		token = resp["token"]
		if token == "" {
			t.Errorf("Expected a JWT token, but got none")
		}
	})

	t.Run("Should fail to login with wrong password", func(t *testing.T) {
		requestBody := `{"email": "test@example.com", "password": "wrongpassword"}`
		req, _ := http.NewRequest("POST", "/api/users/login", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		rr := server.executeRequest(req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
		}
	})

	t.Run("Should fail to register with a duplicate email", func(t *testing.T) {
		// This user was already created in the first test case
		requestBody := `{"username": "anotheruser", "email": "test@example.com", "password": "password456"}`
		req, _ := http.NewRequest("POST", "/api/users/register", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")

		rr := server.executeRequest(req)

		if status := rr.Code; status != http.StatusConflict {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusConflict)
		}
	})
}