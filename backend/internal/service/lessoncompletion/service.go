// Package lessoncompletion là service layer của US4.10: học viên đánh dấu
// hoàn thành bài học; bài học sau bị khóa cho tới khi bài trước hoàn thành.
package lessoncompletion

import (
	"context"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/lessoncompletion"
	"educonnect-lms/backend/internal/domain/user"
)

// CourseGetter là tập con method của course.Repository (cần để tìm khóa học
// chứa bài học, phục vụ kiểm tra đã đăng ký học hay chưa).
type CourseGetter interface {
	FindByID(ctx context.Context, id uint) (*course.Course, error)
}

// EnrollmentChecker là tập con method của enrollment.Repository.
type EnrollmentChecker interface {
	IsEnrolled(ctx context.Context, studentID, courseID uint) (bool, error)
}

type Service struct {
	completions lessoncompletion.Repository
	lessons     curriculum.LessonRepository
	chapters    curriculum.ChapterRepository
	courses     CourseGetter
	enrollments EnrollmentChecker
}

func NewService(
	completions lessoncompletion.Repository,
	lessons curriculum.LessonRepository,
	chapters curriculum.ChapterRepository,
	courses CourseGetter,
	enrollments EnrollmentChecker,
) *Service {
	return &Service{
		completions: completions,
		lessons:     lessons,
		chapters:    chapters,
		courses:     courses,
		enrollments: enrollments,
	}
}

// courseOrderedLessons trả về toàn bộ bài học của 1 khóa học theo đúng thứ
// tự hiển thị: chương theo Position (US4.7), trong mỗi chương là bài học
// theo Position — đây là thứ tự "tuần tự" mà US4.10 dùng để tính khóa/mở.
func (s *Service) courseOrderedLessons(ctx context.Context, courseID uint) ([]*curriculum.Lesson, error) {
	chapters, err := s.chapters.ListByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	var lessons []*curriculum.Lesson
	for _, ch := range chapters {
		ls, err := s.lessons.ListByChapter(ctx, ch.ID())
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, ls...)
	}
	return lessons, nil
}

func (s *Service) courseForLesson(ctx context.Context, lessonID uint) (*curriculum.Lesson, *course.Course, error) {
	l, err := s.lessons.FindByID(ctx, lessonID)
	if err != nil {
		return nil, nil, err
	}
	ch, err := s.chapters.FindByID(ctx, l.ChapterID())
	if err != nil {
		return nil, nil, err
	}
	c, err := s.courses.FindByID(ctx, ch.CourseID())
	if err != nil {
		return nil, nil, err
	}
	return l, c, nil
}

// ListForStudent trả trạng thái hoàn thành + khóa của mọi bài học trong 1
// khóa học, theo đúng thứ tự — dùng để vẽ sidebar course player (US4.9).
// Giảng viên/Admin (role != student) luôn thấy mọi bài mở khóa, không có
// khái niệm "hoàn thành" (chỉ áp dụng cho học viên).
func (s *Service) ListForStudent(ctx context.Context, courseID, userID uint, role user.Role) ([]lessoncompletion.LessonState, error) {
	lessons, err := s.courseOrderedLessons(ctx, courseID)
	if err != nil {
		return nil, err
	}

	if role != user.RoleStudent {
		out := make([]lessoncompletion.LessonState, 0, len(lessons))
		for _, l := range lessons {
			out = append(out, lessoncompletion.LessonState{LessonID: l.ID()})
		}
		return out, nil
	}

	completedSet, err := s.completions.ListCompletedByStudent(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]lessoncompletion.LessonState, 0, len(lessons))
	allPrevCompleted := true
	for _, l := range lessons {
		completed := completedSet[l.ID()]
		locked := !allPrevCompleted
		out = append(out, lessoncompletion.LessonState{LessonID: l.ID(), Completed: completed, Locked: locked})
		if !completed {
			allPrevCompleted = false
		}
	}
	return out, nil
}

// MarkComplete hiện thực US4.10: học viên đánh dấu đã học xong 1 bài học.
// Idempotent — gọi lại trên bài đã hoàn thành không lỗi. Chặn
// (ErrLessonLocked) nếu bài đang bị khóa, đề phòng client bỏ qua việc ẩn
// nút và gọi thẳng API.
func (s *Service) MarkComplete(ctx context.Context, lessonID, studentID uint) error {
	l, c, err := s.courseForLesson(ctx, lessonID)
	if err != nil {
		return err
	}

	enrolled, err := s.enrollments.IsEnrolled(ctx, studentID, c.ID())
	if err != nil {
		return err
	}
	if !enrolled {
		return curriculum.ErrLessonNotFound // không tiết lộ tồn tại, giống pattern US4.3
	}

	already, err := s.completions.IsCompleted(ctx, studentID, lessonID)
	if err != nil {
		return err
	}
	if already {
		return nil
	}

	states, err := s.ListForStudent(ctx, c.ID(), studentID, user.RoleStudent)
	if err != nil {
		return err
	}
	for _, st := range states {
		if st.LessonID == l.ID() {
			if st.Locked {
				return lessoncompletion.ErrLessonLocked
			}
			break
		}
	}

	return s.completions.Create(ctx, lessoncompletion.New(studentID, lessonID))
}
