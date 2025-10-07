package middleware

import (
	"context"
	"fmt"
	"net/http"
	"schedule-app/internal/model"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware holds dependencies for authentication middleware.
type AuthMiddleware struct {
	jwtSecret []byte
}

// NewAuthMiddleware creates a new AuthMiddleware.
func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: []byte(jwtSecret)}
}

// userContextKey is a private type to prevent collisions with other context keys.
type userContextKey string

const userIDKey userContextKey = "userID"

// JwtAuthentication is a middleware to protect routes.
func (amw *AuthMiddleware) JwtAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// "Authorization" ヘッダーからトークンを取得
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// "Bearer " プレフィックスを検証・削除
		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}
		tokenString := bearerToken[1]

		// トークンをパース・検証
		claims := &model.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return amw.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// コンテキストにユーザーIDを格納
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		// 次のハンドラにコンテキストを渡す
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext はコンテキストからユーザーIDを取得します。
func GetUserIDFromContext(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value(userIDKey).(int64)
	if !ok {
		return 0, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}