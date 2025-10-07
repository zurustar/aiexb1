package main

import (
	"log"
	"net/http"
	"schedule-app/internal/config"
	"schedule-app/internal/db"
	"schedule-app/internal/handler"
	"schedule-app/internal/middleware"
	"schedule-app/internal/repository"
)

func main() {
	// 0. 設定を読み込み
	cfg := config.LoadConfig()

	// 1. データベースを初期化
	// サンドボックス環境の書き込み権限問題を回避するため、/tmp にデータベースを作成します。
	conn, err := db.InitDB("/tmp/schedule.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer conn.Close()

	// 2. 依存関係を注入 (DI)
	userRepo := repository.NewUserRepository(conn)
	userHandler := handler.NewUserHandler(userRepo, cfg.JWTSecret)
	scheduleRepo := repository.NewScheduleRepository(conn)
	scheduleHandler := handler.NewScheduleHandler(scheduleRepo)
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)

	// 3. HTTPルーターをセットアップ
	mux := http.NewServeMux()

	// --- ユーザー認証エンドポイント ---
	mux.HandleFunc("POST /api/users/register", userHandler.Register)
	mux.HandleFunc("POST /api/users/login", userHandler.Login)

	// --- スケジュール管理エンドポイント ---
	// 作成 (要認証)
	mux.Handle("POST /api/schedules", authMiddleware.JwtAuthentication(http.HandlerFunc(scheduleHandler.CreateSchedule)))
	// 取得 (公開)
	mux.HandleFunc("GET /api/users/{ownerID}/schedules", scheduleHandler.GetSchedulesByOwner)
	mux.HandleFunc("GET /api/schedules/{scheduleID}", scheduleHandler.GetScheduleByID)
	// 更新 (要認証)
	mux.Handle("PUT /api/schedules/{scheduleID}", authMiddleware.JwtAuthentication(http.HandlerFunc(scheduleHandler.UpdateSchedule)))
	// 削除 (要認証)
	mux.Handle("DELETE /api/schedules/{scheduleID}", authMiddleware.JwtAuthentication(http.HandlerFunc(scheduleHandler.DeleteSchedule)))

	// --- 管理者用エンドポイント ---
	// 全ユーザー取得 (要認証)
	mux.Handle("GET /api/admin/users", authMiddleware.JwtAuthentication(http.HandlerFunc(userHandler.GetAllUsers)))


	// --- 静的ファイル配信 ---
	// API以外のリクエストはwebディレクトリの静적ファイルとして配信
	mux.Handle("/", http.FileServer(http.Dir("web")))


	// 4. HTTPサーバーを起動
	port := "8080"
	log.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}