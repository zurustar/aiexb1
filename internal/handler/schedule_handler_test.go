package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"schedule-app/internal/model"
	"testing"
)

// Helper function to create a user and return their ID
func createUser(t *testing.T, server *testServer, username, email, password string) int64 {
	requestBody := fmt.Sprintf(`{"username": "%s", "email": "%s", "password": "%s"}`, username, email, password)
	req, _ := http.NewRequest("POST", "/api/users/register", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rr := server.executeRequest(req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("Failed to create user %s: %s", username, rr.Body.String())
	}
	var user model.UserResponse
	json.NewDecoder(rr.Body).Decode(&user)
	return user.ID
}

// Helper function to login a user and return their JWT token
func loginUser(t *testing.T, server *testServer, email, password string) string {
	requestBody := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, email, password)
	req, _ := http.NewRequest("POST", "/api/users/login", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rr := server.executeRequest(req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Failed to login user %s: %s", email, rr.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	return resp["token"]
}

func TestScheduleHandlers(t *testing.T) {
	// --- Test Setup ---
	server := newTestServer()
	defer server.db.Close()

	// Create two users, UserA (creator) and UserB (calendar owner)
	userA_ID := createUser(t, server, "usera", "usera@example.com", "password123")
	userB_ID := createUser(t, server, "userb", "userb@example.com", "password456")

	// Get their tokens
	tokenA := loginUser(t, server, "usera@example.com", "password123")
	tokenB := loginUser(t, server, "userb@example.com", "password456")

	var scheduleID int64

	// Create a third user to act as a participant
	userC_ID := createUser(t, server, "userc", "userc@example.com", "password789")

	// --- Test Cases ---
	t.Run("Should create a new schedule with participants", func(t *testing.T) {
		requestBody := fmt.Sprintf(
			`{"title": "Shared Event with Participants", "owner_id": %d, "start_time": "2025-11-01T10:00:00Z", "end_time": "2025-11-01T11:00:00Z", "participant_ids": [%d, %d]}`,
			userB_ID, userA_ID, userC_ID,
		)
		req, _ := http.NewRequest("POST", "/api/schedules", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenA)

		rr := server.executeRequest(req)

		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
		}

		var schedule model.ScheduleResponse
		if err := json.NewDecoder(rr.Body).Decode(&schedule); err != nil {
			t.Fatalf("Could not decode response: %v", err)
		}
		if schedule.Title != "Shared Event with Participants" {
			t.Errorf("Expected title 'Shared Event with Participants', got '%s'", schedule.Title)
		}
		if schedule.CreatorID != userA_ID {
			t.Errorf("Expected creator ID %d, got %d", userA_ID, schedule.CreatorID)
		}
		if schedule.OwnerID != userB_ID {
			t.Errorf("Expected owner ID %d, got %d", userB_ID, schedule.OwnerID)
		}
		if len(schedule.Participants) != 2 {
			t.Fatalf("Expected 2 participants, got %d", len(schedule.Participants))
		}
		// Check if the correct users are participants
		participantIDs := map[int64]bool{schedule.Participants[0].ID: true, schedule.Participants[1].ID: true}
		if !participantIDs[userA_ID] || !participantIDs[userC_ID] {
			t.Errorf("Expected participants to be User A and User C, but they were not")
		}
		scheduleID = schedule.ID
	})

	t.Run("Should allow creator to update the schedule and its participants", func(t *testing.T) {
		// User A (creator) updates the schedule to only include User B as a participant
		requestBody := fmt.Sprintf(`{"title": "Updated Shared Event", "participant_ids": [%d]}`, userB_ID)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/schedules/%d", scheduleID), bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenA)

		rr := server.executeRequest(req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		var schedule model.ScheduleResponse
		if err := json.NewDecoder(rr.Body).Decode(&schedule); err != nil {
			t.Fatalf("Could not decode response: %v", err)
		}
		if schedule.Title != "Updated Shared Event" {
			t.Errorf("Expected title 'Updated Shared Event', got '%s'", schedule.Title)
		}
		if len(schedule.Participants) != 1 {
			t.Fatalf("Expected 1 participant, got %d", len(schedule.Participants))
		}
		if schedule.Participants[0].ID != userB_ID {
			t.Errorf("Expected participant to be User B, got user ID %d", schedule.Participants[0].ID)
		}
	})

	t.Run("Should forbid calendar owner (not creator) from updating", func(t *testing.T) {
		requestBody := `{"title": "Forbidden Update"}`
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/schedules/%d", scheduleID), bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenB)

		rr := server.executeRequest(req)
		if status := rr.Code; status != http.StatusForbidden {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusForbidden)
		}
	})

	t.Run("Should forbid calendar owner (not creator) from deleting", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/schedules/%d", scheduleID), nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)

		rr := server.executeRequest(req)
		if status := rr.Code; status != http.StatusForbidden {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusForbidden)
		}
	})

	t.Run("Should allow creator to delete the schedule", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/schedules/%d", scheduleID), nil)
		req.Header.Set("Authorization", "Bearer "+tokenA)

		rr := server.executeRequest(req)
		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
		}
	})
}