// Package report là read model của US7.2: Giảng viên/Quản trị viên xem báo
// cáo thống kê học viên/khóa học. Cùng cách tiếp cận read-model chỉ-đọc như
// gradebook/progress, không có invariant/factory function riêng.
package report

import "context"

// CourseStats là số liệu thống kê của 1 khóa học: số học viên đã đăng ký,
// tổng số bài tập, và tỉ lệ hoàn thành trung bình (dựa trên tỉ lệ bài tập
// đã nộp trên tổng số bài tập, tính trung bình qua mọi học viên).
type CourseStats struct {
	CourseID          uint
	CourseTitle       string
	EnrolledStudents  int
	TotalAssignments  int
	AverageCompletion float64
}

type Repository interface {
	// ForTeacher trả về thống kê mọi khóa học do giảng viên đó sở hữu.
	ForTeacher(ctx context.Context, teacherID uint) ([]CourseStats, error)
	// All trả về thống kê mọi khóa học trong hệ thống (dành cho quản trị viên).
	All(ctx context.Context) ([]CourseStats, error)
}
