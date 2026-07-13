// Package gradebook là read model của US5.3 (phần "hệ thống tự tổng hợp
// bảng điểm"): tổng hợp điểm của mọi học viên đã đăng ký 1 khóa học, theo
// từng bài tập/trắc nghiệm. Đây là dữ liệu tổng hợp từ nhiều aggregate
// (Enrollment, Assignment, Submission) nên không có invariant/factory
// function riêng như các domain khác — Repository dựng Entry trực tiếp
// bằng SQL join (không đi qua từng aggregate) vì đây thuần là truy vấn đọc.
package gradebook

import "context"

// Entry là điểm của 1 học viên cho 1 bài tập trong 1 khóa học. Score = nil
// nghĩa là học viên chưa nộp hoặc bài tự luận chưa được giảng viên chấm.
type Entry struct {
	StudentID       uint
	StudentName     string
	AssignmentID    uint
	AssignmentTitle string
	Score           *float64
}

type Repository interface {
	// ForCourse trả về 1 dòng cho mỗi cặp (học viên đã đăng ký, bài tập)
	// thuộc khóa học — kể cả khi học viên chưa nộp bài (Score = nil).
	ForCourse(ctx context.Context, courseID uint) ([]Entry, error)
}
