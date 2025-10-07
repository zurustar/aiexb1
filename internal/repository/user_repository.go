package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"schedule-app/internal/model"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// ErrDuplicateEntry is returned when a database insert fails due to a UNIQUE constraint.
var ErrDuplicateEntry = errors.New("duplicate entry")

// UserRepository はユーザー関連のデータベース操作を扱います。
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository は UserRepository の新しいインスタンスを生成します。
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser は新しいユーザーを作成し、データベースに保存します。
func (r *UserRepository) CreateUser(req *model.RegisterUserRequest) (*model.User, error) {
	// パスワードをハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// ユーザーをデータベースに挿入
	result, err := r.db.Exec("INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?);", req.Username, req.Email, string(hashedPassword))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrDuplicateEntry
		}
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	// 挿入されたユーザーのIDを取得
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	// 作成したユーザー情報を取得して返す
	return r.FindUserByID(id)
}

// FindUserByID はIDでユーザーを検索します。
func (r *UserRepository) FindUserByID(id int64) (*model.User, error) {
	var user model.User
	query := "SELECT id, username, email, password_hash, created_at FROM users WHERE id = ? LIMIT 1;"
	row := r.db.QueryRow(query, id)

	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("query for user by id failed: %w", err)
	}
	return &user, nil
}


// FindUserByEmail はEmailでユーザーを検索します。
func (r *UserRepository) FindUserByEmail(email string) (*model.User, error) {
	var user model.User
	query := "SELECT id, username, email, password_hash, created_at FROM users WHERE email = ? LIMIT 1;"
	row := r.db.QueryRow(query, email)

	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// 認証失敗時はエラーメッセージを曖昧にするため、ハンドラ側で「ユーザーが見つからない」ことを直接返さないようにする
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("query for user by email failed: %w", err)
	}
	return &user, nil
}