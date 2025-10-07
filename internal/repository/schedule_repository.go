package repository

import (
	"database/sql"
	"fmt"
	"schedule-app/internal/model"
	"strings"
	"time"
)

// ScheduleRepository はスケジュール関連のデータベース操作を扱います。
type ScheduleRepository struct {
	db *sql.DB
}

// NewScheduleRepository は ScheduleRepository の新しいインスタンスを生成します。
func NewScheduleRepository(db *sql.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

// Create は新しいスケジュールを作成し、データベースに保存します。
func (r *ScheduleRepository) Create(req *model.CreateScheduleRequest, creatorID int64) (*model.Schedule, error) {
	query := `
		INSERT INTO schedules (title, start_time, end_time, description, location, owner_id, creator_id)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	result, err := r.db.Exec(query, req.Title, req.StartTime, req.EndTime, req.Description, req.Location, req.OwnerID, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert schedule: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return r.FindByID(id)
}

// FindByID はIDでスケジュールを検索します。
func (r *ScheduleRepository) FindByID(id int64) (*model.Schedule, error) {
	var s model.Schedule
	query := `
		SELECT id, title, start_time, end_time, description, location, owner_id, creator_id, created_at, updated_at
		FROM schedules WHERE id = ?;
	`
	row := r.db.QueryRow(query, id)
	err := row.Scan(&s.ID, &s.Title, &s.StartTime, &s.EndTime, &s.Description, &s.Location, &s.OwnerID, &s.CreatorID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schedule with id %d not found", id)
		}
		return nil, fmt.Errorf("query for schedule by id failed: %w", err)
	}
	return &s, nil
}

// FindByOwnerID は指定された所有者のスケジュールをすべて取得します。
func (r *ScheduleRepository) FindByOwnerID(ownerID int64) ([]*model.Schedule, error) {
	query := `
		SELECT id, title, start_time, end_time, description, location, owner_id, creator_id, created_at, updated_at
		FROM schedules WHERE owner_id = ? ORDER BY start_time ASC;
	`
	rows, err := r.db.Query(query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("query for schedules by owner id failed: %w", err)
	}
	defer rows.Close()

	var schedules []*model.Schedule
	for rows.Next() {
		var s model.Schedule
		err := rows.Scan(&s.ID, &s.Title, &s.StartTime, &s.EndTime, &s.Description, &s.Location, &s.OwnerID, &s.CreatorID, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule row: %w", err)
		}
		schedules = append(schedules, &s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return schedules, nil
}


// Update は既存のスケジュール情報を更新します。
// リクエストで指定されたnilでないフィールドのみを動的に更新します。
func (r *ScheduleRepository) Update(id int64, req *model.UpdateScheduleRequest) (*model.Schedule, error) {
	var setClauses []string
	var args []interface{}

	if req.Title != nil {
		setClauses = append(setClauses, "title = ?")
		args = append(args, *req.Title)
	}
	if req.StartTime != nil {
		setClauses = append(setClauses, "start_time = ?")
		args = append(args, *req.StartTime)
	}
	if req.EndTime != nil {
		setClauses = append(setClauses, "end_time = ?")
		args = append(args, *req.EndTime)
	}
	if req.Description != nil {
		setClauses = append(setClauses, "description = ?")
		args = append(args, *req.Description)
	}
	if req.Location != nil {
		setClauses = append(setClauses, "location = ?")
		args = append(args, *req.Location)
	}

	// 更新するフィールドがない場合は、何もせずに現在のスケジュールを返す
	if len(setClauses) == 0 {
		return r.FindByID(id)
	}

	// updated_at は常に更新する
	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, time.Now())

	// 動的にクエリを構築
	query := fmt.Sprintf("UPDATE schedules SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	args = append(args, id)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update schedule: %w", err)
	}

	return r.FindByID(id)
}

// Delete はIDでスケジュールを削除します。
func (r *ScheduleRepository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM schedules WHERE id = ?;", id)
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("schedule with id %d not found for deletion", id)
	}
	return nil
}