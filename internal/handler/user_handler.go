package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"schedule-app/internal/model"
	"schedule-app/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// UserHandler はユーザー関連のHTTPリクエストを処理します。
type UserHandler struct {
	userRepo  *repository.UserRepository
	jwtSecret []byte
}

// NewUserHandler は UserHandler の新しいインスタンスを生成します。
func NewUserHandler(userRepo *repository.UserRepository, jwtSecret string) *UserHandler {
	return &UserHandler{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

// Register はユーザー登録のためのハンドラです。
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: バリデーションの追加 (username, email, password の形式チェック)

	user, err := h.userRepo.CreateUser(&req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEntry) {
			http.Error(w, "Username or email already exists", http.StatusConflict)
		} else {
			log.Printf("ERROR: Failed to create user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user.ToUserResponse())
}

// Login はユーザーログインのためのハンドラです。
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.FindUserByEmail(req.Email)
	if err != nil {
		// ユーザーが見つからない場合も、パスワードが違う場合も同じエラーを返す
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// パスワードを比較
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		// パスワードが一致しない
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// トークンの有効期限を設定 (例: 24時間)
	expirationTime := time.Now().Add(24 * time.Hour)

	// JWTのクレームを設定
	claims := &model.Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// 新しいトークンを生成し、設定した秘密鍵で署名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	// トークンをJSONレスポンスとして返す
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// writeJSON はGoの構造体をJSONレスポンスとして書き込みます。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// エンコードに失敗した場合はログに出力
		fmt.Printf("Failed to encode response: %v\n", err)
	}
}

// errorJSON はエラーメッセージをJSON形式で返します。
func errorJSON(w http.ResponseWriter, status int, message string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	writeJSON(w, status, errorResponse{Error: message})
}