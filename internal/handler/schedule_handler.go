package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"schedule-app/internal/middleware"
	"schedule-app/internal/model"
	"schedule-app/internal/repository"
	"strconv"
	"strings"
)

// ScheduleHandler はスケジュール関連のHTTPリクエストを処理します。
type ScheduleHandler struct {
	scheduleRepo *repository.ScheduleRepository
}

// NewScheduleHandler は ScheduleHandler の新しいインスタンスを生成します。
func NewScheduleHandler(scheduleRepo *repository.ScheduleRepository) *ScheduleHandler {
	return &ScheduleHandler{scheduleRepo: scheduleRepo}
}

// CreateSchedule は新しいスケジュールを作成するためのハンドラです。
func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	creatorID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	var req model.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 入力値のバリデーション
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		errorJSON(w, http.StatusBadRequest, "Title is required")
		return
	}
	if req.OwnerID == 0 {
		errorJSON(w, http.StatusBadRequest, "OwnerID is required")
		return
	}

	schedule, err := h.scheduleRepo.Create(&req, creatorID)
	if err != nil {
		log.Printf("ERROR: Failed to create schedule: %v", err)
		errorJSON(w, http.StatusInternalServerError, "Failed to create schedule")
		return
	}

	writeJSON(w, http.StatusCreated, schedule.ToScheduleResponse())
}

// GetSchedulesByOwner は特定のユーザーが所有するスケジュール一覧を取得します。
func (h *ScheduleHandler) GetSchedulesByOwner(w http.ResponseWriter, r *http.Request) {
	ownerIDStr := r.PathValue("ownerID")
	ownerID, err := strconv.ParseInt(ownerIDStr, 10, 64)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, "Invalid owner ID")
		return
	}

	schedules, err := h.scheduleRepo.FindByOwnerID(ownerID)
	if err != nil {
		log.Printf("ERROR: Failed to get schedules for owner %d: %v", ownerID, err)
		errorJSON(w, http.StatusInternalServerError, "Failed to retrieve schedules")
		return
	}

	var resp []*model.ScheduleResponse
	for _, s := range schedules {
		resp = append(resp, s.ToScheduleResponse())
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetScheduleByID はIDで特定のスケジュールを取得します。
func (h *ScheduleHandler) GetScheduleByID(w http.ResponseWriter, r *http.Request) {
	scheduleIDStr := r.PathValue("scheduleID")
	scheduleID, err := strconv.ParseInt(scheduleIDStr, 10, 64)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, "Invalid schedule ID")
		return
	}

	schedule, err := h.scheduleRepo.FindByID(scheduleID)
	if err != nil {
		log.Printf("ERROR: Failed to get schedule %d: %v", scheduleID, err)
		if strings.Contains(err.Error(), "not found") {
			errorJSON(w, http.StatusNotFound, "Schedule not found")
		} else {
			errorJSON(w, http.StatusInternalServerError, "Failed to retrieve schedule")
		}
		return
	}

	writeJSON(w, http.StatusOK, schedule.ToScheduleResponse())
}

// UpdateSchedule は既存のスケジュールを更新します。
// 権限チェックはリポジトリ層で行います。
func (h *ScheduleHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	scheduleIDStr := r.PathValue("scheduleID")
	scheduleID, err := strconv.ParseInt(scheduleIDStr, 10, 64)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, "Invalid schedule ID")
		return
	}

	var req model.UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updatedSchedule, err := h.scheduleRepo.Update(scheduleID, &req, userID)
	if err != nil {
		log.Printf("ERROR: Failed to update schedule %d: %v", scheduleID, err)
		if strings.Contains(err.Error(), "not found") {
			errorJSON(w, http.StatusNotFound, "Schedule not found")
		} else if strings.Contains(err.Error(), "not authorized") {
			errorJSON(w, http.StatusForbidden, "Forbidden: You are not authorized to update this schedule")
		} else {
			errorJSON(w, http.StatusInternalServerError, "Failed to update schedule")
		}
		return
	}

	writeJSON(w, http.StatusOK, updatedSchedule.ToScheduleResponse())
}

// DeleteSchedule はスケジュールを削除します。
// 権限チェックはリポジトリ層で行います。
func (h *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	scheduleIDStr := r.PathValue("scheduleID")
	scheduleID, err := strconv.ParseInt(scheduleIDStr, 10, 64)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, "Invalid schedule ID")
		return
	}

	err = h.scheduleRepo.Delete(scheduleID, userID)
	if err != nil {
		log.Printf("ERROR: Failed to delete schedule %d: %v", scheduleID, err)
		if strings.Contains(err.Error(), "not found") {
			errorJSON(w, http.StatusNotFound, "Schedule not found")
		} else if strings.Contains(err.Error(), "not authorized") {
			errorJSON(w, http.StatusForbidden, "Forbidden: You are not authorized to delete this schedule")
		} else {
			errorJSON(w, http.StatusInternalServerError, "Failed to delete schedule")
		}
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}