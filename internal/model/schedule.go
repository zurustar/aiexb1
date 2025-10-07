package model

import "time"

// Schedule represents a schedule event in the database.
type Schedule struct {
	ID          int64
	Title       string
	StartTime   time.Time
	EndTime     time.Time
	Description string
	Location    string
	OwnerID     int64
	CreatorID   int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateScheduleRequest defines the request body for creating a new schedule.
type CreateScheduleRequest struct {
	Title       string    `json:"title"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	OwnerID     int64     `json:"owner_id"` // The ID of the user whose calendar this event belongs to.
	// TODO: Add participants
}

// UpdateScheduleRequest defines the request body for updating an existing schedule.
// Using pointers to distinguish between empty values and omitted fields.
type UpdateScheduleRequest struct {
	Title       *string    `json:"title"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Description *string    `json:"description"`
	Location    *string    `json:"location"`
	// TODO: Add participants update logic
}

// ScheduleResponse defines the structure of a schedule event returned by the API.
type ScheduleResponse struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	OwnerID     int64     `json:"owner_id"`
	CreatorID   int64     `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// TODO: Add participants details
}

// ToScheduleResponse converts a Schedule model to a ScheduleResponse.
func (s *Schedule) ToScheduleResponse() *ScheduleResponse {
	return &ScheduleResponse{
		ID:          s.ID,
		Title:       s.Title,
		StartTime:   s.StartTime,
		EndTime:     s.EndTime,
		Description: s.Description,
		Location:    s.Location,
		OwnerID:     s.OwnerID,
		CreatorID:   s.CreatorID,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}