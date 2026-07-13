// Package progress là read model của US7.1: Học viên xem bảng điều khiển
// (dashboard) tiến độ học tập của mình. Tiến độ mỗi khóa học được tính theo
// tỉ lệ bài tập/trắc nghiệm đã nộp trên tổng số bài tập của khóa học — cùng
// cách tiếp cận read-model như package gradebook, không có invariant/factory
// function riêng vì đây thuần là dữ liệu tổng hợp chỉ-đọc.
package progress

import "context"

// CourseProgress là tiến độ của 1 học viên trong 1 khóa học đã đăng ký.
type CourseProgress struct {
	CourseID         uint
	CourseTitle      string
	TotalAssignments int
	Submitted        int
	// PercentComplete = Submitted/TotalAssignments*100, làm tròn về 0 khi
	// khóa học chưa có bài tập nào (tránh chia cho 0).
	PercentComplete float64
}

type Repository interface {
	// ForStudent trả về tiến độ của mọi khóa học mà học viên đã đăng ký.
	ForStudent(ctx context.Context, studentID uint) ([]CourseProgress, error)
}
