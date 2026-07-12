// Package handler chuyển đổi request HTTP (chi) sang lời gọi tầng service và
// map lỗi domain/service sang mã HTTP status. Tầng này không biết gì về
// SQL hay bcrypt — chỉ làm việc với service.
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
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

// parseIDString parse 1 URL param dạng chuỗi số thành uint, dùng chung bởi
// mọi handler cần đọc {id}/{courseId}/{chapterId}/... từ route.
func parseIDString(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
