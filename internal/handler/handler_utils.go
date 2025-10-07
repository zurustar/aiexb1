package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// writeJSON はGoの構造体をJSONレスポンスとして書き込みます。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		if err := json.NewEncoder(w).Encode(v); err != nil {
			// エンコードに失敗した場合はログに出力
			fmt.Printf("Failed to encode response: %v\n", err)
		}
	}
}

// errorJSON はエラーメッセージをJSON形式で返します。
func errorJSON(w http.ResponseWriter, status int, message string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	writeJSON(w, status, errorResponse{Error: message})
}