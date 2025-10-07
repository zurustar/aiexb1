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
// スケジュール作成と参加者追加を単一トランザクションで実行します。
func (r *ScheduleRepository) Create(req *model.CreateScheduleRequest, creatorID int64) (*model.Schedule, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // エラー発生時にロールバック

	// スケジュールを挿入
	query := `
		INSERT INTO schedules (title, start_time, end_time, description, location, owner_id, creator_id)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	result, err := tx.Exec(query, req.Title, req.StartTime, req.EndTime, req.Description, req.Location, req.OwnerID, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert schedule: %w", err)
	}
	scheduleID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	// 参加者を `schedule_participants` テーブルに追加
	if len(req.ParticipantIDs) > 0 {
		stmt, err := tx.Prepare("INSERT INTO schedule_participants (schedule_id, user_id) VALUES (?, ?)")
		if err != nil {
			return nil, fmt.Errorf("failed to prepare participant statement: %w", err)
		}
		defer stmt.Close()
		for _, userID := range req.ParticipantIDs {
			if _, err := stmt.Exec(scheduleID, userID); err != nil {
				return nil, fmt.Errorf("failed to insert participant %d: %w", userID, err)
			}
		}
	}

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.FindByID(scheduleID)
}

// FindByID はIDでスケジュールを検索し、参加者情報も取得します。
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

	// 参加者情報を取得
	participants, err := r.findParticipantsByScheduleID(s.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find participants for schedule %d: %w", s.ID, err)
	}
	s.Participants = participants

	return &s, nil
}

// findParticipantsByScheduleID は指定されたスケジュールIDの参加者リストを取得します。
func (r *ScheduleRepository) findParticipantsByScheduleID(scheduleID int64) ([]*model.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.created_at
		FROM users u
		JOIN schedule_participants sp ON u.id = sp.user_id
		WHERE sp.schedule_id = ?;
	`
	rows, err := r.db.Query(query, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("query for participants failed: %w", err)
	}
	defer rows.Close()

	var participants []*model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan participant row: %w", err)
		}
		participants = append(participants, &u)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during participant rows iteration: %w", err)
	}

	return participants, nil
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
		// 各スケジュールの参加者情報を取得
		participants, err := r.findParticipantsByScheduleID(s.ID)
		if err != nil {
			// 1つのスケジュールの参加者取得に失敗しても、全体を失敗させない（エラーログは出すべき）
			fmt.Printf("could not fetch participants for schedule %d: %v\n", s.ID, err)
		}
		s.Participants = participants
		schedules = append(schedules, &s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return schedules, nil
}

// Update は既存のスケジュール情報を更新します。
// リクエストで指定されたnilでないフィールドのみを動的に更新します。
func (r *ScheduleRepository) Update(id int64, req *model.UpdateScheduleRequest, userID int64) (*model.Schedule, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 更新権限をチェック (作成者のみが更新可能)
	var creatorID int64
	err = tx.QueryRow("SELECT creator_id FROM schedules WHERE id = ?", id).Scan(&creatorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schedule with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to query creator_id: %w", err)
	}
	if creatorID != userID {
		return nil, fmt.Errorf("user %d is not authorized to update schedule %d", userID, id)
	}

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

	// スケジュール本体の更新
	if len(setClauses) > 0 {
		setClauses = append(setClauses, "updated_at = ?")
		args = append(args, time.Now())
		query := fmt.Sprintf("UPDATE schedules SET %s WHERE id = ?", strings.Join(setClauses, ", "))
		args = append(args, id)
		if _, err := tx.Exec(query, args...); err != nil {
			return nil, fmt.Errorf("failed to update schedule: %w", err)
		}
	}

	// 参加者の更新
	if req.ParticipantIDs != nil {
		// 既存の参加者を削除
		if _, err := tx.Exec("DELETE FROM schedule_participants WHERE schedule_id = ?", id); err != nil {
			return nil, fmt.Errorf("failed to delete existing participants: %w", err)
		}
		// 新しい参加者を追加
		if len(*req.ParticipantIDs) > 0 {
			stmt, err := tx.Prepare("INSERT INTO schedule_participants (schedule_id, user_id) VALUES (?, ?)")
			if err != nil {
				return nil, fmt.Errorf("failed to prepare participant statement for update: %w", err)
			}
			defer stmt.Close()
			for _, userID := range *req.ParticipantIDs {
				if _, err := stmt.Exec(id, userID); err != nil {
					return nil, fmt.Errorf("failed to insert participant %d for update: %w", userID, err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.FindByID(id)
}

// Delete はIDでスケジュールを削除します。作成者のみが削除可能です。
func (r *ScheduleRepository) Delete(id int64, userID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 削除権限をチェック
	var creatorID int64
	err = tx.QueryRow("SELECT creator_id FROM schedules WHERE id = ?", id).Scan(&creatorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("schedule with id %d not found", id)
		}
		return fmt.Errorf("failed to query creator_id for deletion: %w", err)
	}
	if creatorID != userID {
		return fmt.Errorf("user %d is not authorized to delete schedule %d", userID, id)
	}

	// 関連する参加者情報を削除 (CASCADE DELETEが設定されているが、念のため)
	if _, err := tx.Exec("DELETE FROM schedule_participants WHERE schedule_id = ?;", id); err != nil {
		return fmt.Errorf("failed to delete participants for schedule %d: %w", id, err)
	}

	// スケジュールを削除
	result, err := tx.Exec("DELETE FROM schedules WHERE id = ?;", id)
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

	return tx.Commit()
}