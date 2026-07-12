// Package handler chuyển đổi request HTTP (chi) sang lời gọi tầng service và
// map lỗi domain/service sang mã HTTP status. Tầng này không biết gì về
// SQL hay bcrypt — chỉ làm việc với service.
package handler

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
