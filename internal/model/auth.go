package model

import "github.com/golang-jwt/jwt/v5"

// Claims represents the JWT claims used in the application.
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}