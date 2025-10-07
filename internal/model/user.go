package model

import "time"

// User はデータベースの users テーブルに対応する構造体です。
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // パスワードハッシュはJSONに含めない
	CreatedAt    time.Time `json:"created_at"`
}

// RegisterUserRequest はユーザー登録APIのリクエストボディを表します。
type RegisterUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginUserRequest はログインAPIのリクエストボディを表します。
type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserResponse はAPIから返すユーザー情報の構造体です。
type UserResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// ToUserResponse は User モデルを UserResponse に変換します。
func (u *User) ToUserResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}