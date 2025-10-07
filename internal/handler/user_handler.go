package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"schedule-app/internal/model"
	"schedule-app/internal/repository"
	"strings"
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

// emailRegex はEメールアドレスの形式を検証するための正規表現です。
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Register はユーザー登録のためのハンドラです。
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 入力値のバリデーション
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if len(req.Username) < 3 {
		errorJSON(w, http.StatusBadRequest, "Username must be at least 3 characters long")
		return
	}
	if !emailRegex.MatchString(req.Email) {
		errorJSON(w, http.StatusBadRequest, "Invalid email format")
		return
	}
	if len(req.Password) < 8 {
		errorJSON(w, http.StatusBadRequest, "Password must be at least 8 characters long")
		return
	}

	user, err := h.userRepo.CreateUser(&req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEntry) {
			errorJSON(w, http.StatusConflict, "Username or email already exists")
		} else {
			log.Printf("ERROR: Failed to create user: %v", err)
			errorJSON(w, http.StatusInternalServerError, "Failed to create user")
		}
		return
	}

	writeJSON(w, http.StatusCreated, user.ToUserResponse())
}

// Login はユーザーログインのためのハンドラです。
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.userRepo.FindUserByEmail(req.Email)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// パスワードを比較
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, "Invalid email or password")
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
		errorJSON(w, http.StatusInternalServerError, "Failed to create token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": tokenString})
}

// GetAllUsers はすべてのユーザーのリストを取得します。
// 本番環境では、このエンドポイントは管理者のみがアクセスできるように制限する必要があります。
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.FindAll()
	if err != nil {
		log.Printf("ERROR: Failed to get all users: %v", err)
		errorJSON(w, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}

	var resp []*model.UserResponse
	for _, u := range users {
		resp = append(resp, u.ToUserResponse())
	}

	writeJSON(w, http.StatusOK, resp)
}