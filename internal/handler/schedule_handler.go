package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"schedule-app/internal/middleware"
	"schedule-app/internal/model"
	"schedule-app/internal/repository"
	"strconv"
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
	// コンテキストから認証済みユーザーのIDを取得
	creatorID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var req model.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: バリデーション (title, owner_id, times)

	schedule, err := h.scheduleRepo.Create(&req, creatorID)
	if err != nil {
		log.Printf("ERROR: Failed to create schedule: %v", err)
		http.Error(w, "Failed to create schedule", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, schedule.ToScheduleResponse())
}

// GetSchedulesByOwner は特定のユーザーが所有するスケジュール一覧を取得します。
func (h *ScheduleHandler) GetSchedulesByOwner(w http.ResponseWriter, r *http.Request) {
	// URLから所有者IDを取得
	ownerIDStr := r.PathValue("ownerID")
	ownerID, err := strconv.ParseInt(ownerIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid owner ID", http.StatusBadRequest)
		return
	}

	schedules, err := h.scheduleRepo.FindByOwnerID(ownerID)
	if err != nil {
		log.Printf("ERROR: Failed to get schedules for owner %d: %v", ownerID, err)
		http.Error(w, "Failed to retrieve schedules", http.StatusInternalServerError)
		return
	}

	// レスポンス用に変換
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
		http.Error(w, "Invalid schedule ID", http.StatusBadRequest)
		return
	}

	schedule, err := h.scheduleRepo.FindByID(scheduleID)
	if err != nil {
		// TODO: エラーの種類によってステータスコードを分ける (e.g., not found -> 404)
		log.Printf("ERROR: Failed to get schedule %d: %v", scheduleID, err)
		http.Error(w, "Failed to retrieve schedule", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, schedule.ToScheduleResponse())
}

// UpdateSchedule は既存のスケジュールを更新します。
func (h *ScheduleHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	// コンテキストから認証済みユーザーのIDを取得
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	scheduleIDStr := r.PathValue("scheduleID")
	scheduleID, err := strconv.ParseInt(scheduleIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid schedule ID", http.StatusBadRequest)
		return
	}

	// 更新対象のスケジュールが存在するか、また更新権限があるかを確認
	scheduleToUpdate, err := h.scheduleRepo.FindByID(scheduleID)
	if err != nil {
		http.Error(w, "Schedule not found", http.StatusNotFound)
		return
	}

	// 権限チェック: スケジュールを作成したユーザーのみが更新できる
	if scheduleToUpdate.CreatorID != userID {
		http.Error(w, "Forbidden: You are not allowed to update this schedule", http.StatusForbidden)
		return
	}

	var req model.UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: 部分更新のロジックをリポジトリ側で実装する
	updatedSchedule, err := h.scheduleRepo.Update(scheduleID, &req)
	if err != nil {
		log.Printf("ERROR: Failed to update schedule %d: %v", scheduleID, err)
		http.Error(w, "Failed to update schedule", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, updatedSchedule.ToScheduleResponse())
}


// DeleteSchedule はスケジュールを削除します。
func (h *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	scheduleIDStr := r.PathValue("scheduleID")
	scheduleID, err := strconv.ParseInt(scheduleIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid schedule ID", http.StatusBadRequest)
		return
	}

	scheduleToDelete, err := h.scheduleRepo.FindByID(scheduleID)
	if err != nil {
		http.Error(w, "Schedule not found", http.StatusNotFound)
		return
	}

	// 権限チェック: スケジュールを作成したユーザーのみが削除できる
	if scheduleToDelete.CreatorID != userID {
		http.Error(w, "Forbidden: You are not allowed to delete this schedule", http.StatusForbidden)
		return
	}

	err = h.scheduleRepo.Delete(scheduleID)
	if err != nil {
		log.Printf("ERROR: Failed to delete schedule %d: %v", scheduleID, err)
		http.Error(w, "Failed to delete schedule", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}