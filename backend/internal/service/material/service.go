// Package material là service layer của US4.1 (upload) / US4.2 (xem) và
// US4.3 (bảo vệ truy cập): chỉ học viên đã đăng ký khóa học, giảng viên sở
// hữu khóa học, hoặc quản trị viên mới được xem/tải tài liệu. US4.8: giảng
// viên sở hữu khóa học (hoặc admin) xóa được tài liệu đã upload nhầm.
package material

import (
	"context"
	"io"
	"strconv"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/material"
	"educonnect-lms/backend/internal/domain/user"
)

// FileStorage là port lưu trữ file vật lý, implement bởi
// internal/platform/storage (local disk cho bản demo).
type FileStorage interface {
	Save(ctx context.Context, subdir, fileName string, content io.Reader) (path string, err error)
	Delete(ctx context.Context, path string) error
}

// CourseGetter là tập con method của course.Repository (US4.3: xác nhận
// giảng viên sở hữu khóa học chứa tài liệu).
type CourseGetter interface {
	FindByID(ctx context.Context, id uint) (*course.Course, error)
}

// EnrollmentChecker là tập con method của enrollment.Repository (US4.3:
// xác nhận học viên đã đăng ký khóa học chứa tài liệu).
type EnrollmentChecker interface {
	IsEnrolled(ctx context.Context, studentID, courseID uint) (bool, error)
}

type Service struct {
	materials   material.Repository
	lessons     curriculum.LessonRepository
	chapters    curriculum.ChapterRepository
	courses     CourseGetter
	enrollments EnrollmentChecker
	storage     FileStorage
}

func NewService(
	materials material.Repository,
	lessons curriculum.LessonRepository,
	chapters curriculum.ChapterRepository,
	courses CourseGetter,
	enrollments EnrollmentChecker,
	storage FileStorage,
) *Service {
	return &Service{
		materials:   materials,
		lessons:     lessons,
		chapters:    chapters,
		courses:     courses,
		enrollments: enrollments,
		storage:     storage,
	}
}

// Upload hiện thực US4.1: xác nhận Lesson tồn tại, lưu file vật lý qua
// FileStorage, rồi ghi metadata vào bảng materials.
func (s *Service) Upload(ctx context.Context, lessonID uint, fileName string, content io.Reader) (*material.Material, error) {
	if _, err := s.lessons.FindByID(ctx, lessonID); err != nil {
		return nil, err
	}

	subdir := lessonSubdir(lessonID)
	path, err := s.storage.Save(ctx, subdir, fileName, content)
	if err != nil {
		return nil, err
	}

	m, err := material.NewMaterial(lessonID, fileName, path)
	if err != nil {
		return nil, err
	}
	if err := s.materials.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// ListByLesson hiện thực US4.2 (Học viên xem/tải tài liệu). Chỉ trả danh
// sách nếu userID có quyền truy cập lesson này (US4.3).
func (s *Service) ListByLesson(ctx context.Context, lessonID, userID uint, role user.Role) ([]*material.Material, error) {
	allowed, err := s.canAccessLesson(ctx, lessonID, userID, role)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, material.ErrNotFound // không tiết lộ nội dung khóa học người dùng chưa có quyền
	}
	return s.materials.ListByLesson(ctx, lessonID)
}

// Get lấy metadata 1 material để phục vụ download (US4.3), có kiểm tra
// quyền truy cập giống ListByLesson.
func (s *Service) Get(ctx context.Context, materialID, userID uint, role user.Role) (*material.Material, error) {
	m, err := s.materials.FindByID(ctx, materialID)
	if err != nil {
		return nil, err
	}
	allowed, err := s.canAccessLesson(ctx, m.LessonID(), userID, role)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, material.ErrNotFound
	}
	return m, nil
}

// Delete hiện thực US4.8: xóa cả file vật lý lẫn metadata. Khác với
// ListByLesson/Get (học viên đã đăng ký cũng xem/tải được), xóa chỉ dành
// cho giảng viên sở hữu khóa học hoặc admin — học viên không bao giờ được
// xóa dù có đăng ký hay không.
func (s *Service) Delete(ctx context.Context, materialID, userID uint, role user.Role) error {
	m, err := s.materials.FindByID(ctx, materialID)
	if err != nil {
		return err
	}
	allowed, err := s.isManagerOfLesson(ctx, m.LessonID(), userID, role)
	if err != nil {
		return err
	}
	if !allowed {
		return material.ErrNotFound // không tiết lộ tài liệu của khóa học người khác tồn tại
	}
	if err := s.storage.Delete(ctx, m.FilePath()); err != nil {
		return err
	}
	return s.materials.Delete(ctx, materialID)
}

// canAccessLesson hiện thực US4.3: admin luôn được phép; giảng viên chỉ
// được phép nếu sở hữu khóa học chứa lesson; học viên chỉ được phép nếu đã
// đăng ký khóa học đó.
func (s *Service) canAccessLesson(ctx context.Context, lessonID, userID uint, role user.Role) (bool, error) {
	if role == user.RoleAdmin {
		return true, nil
	}
	c, err := s.courseForLesson(ctx, lessonID)
	if err != nil {
		return false, err
	}
	if role == user.RoleTeacher {
		return c.TeacherID() == userID, nil
	}
	return s.enrollments.IsEnrolled(ctx, userID, c.ID())
}

// isManagerOfLesson: admin hoặc giảng viên sở hữu khóa học — KHÔNG bao gồm
// học viên dù đã đăng ký, dùng riêng cho các thao tác quản trị như Delete
// (US4.8) mà học viên không bao giờ được phép thực hiện.
func (s *Service) isManagerOfLesson(ctx context.Context, lessonID, userID uint, role user.Role) (bool, error) {
	if role == user.RoleAdmin {
		return true, nil
	}
	if role != user.RoleTeacher {
		return false, nil
	}
	c, err := s.courseForLesson(ctx, lessonID)
	if err != nil {
		return false, err
	}
	return c.TeacherID() == userID, nil
}

func (s *Service) courseForLesson(ctx context.Context, lessonID uint) (*course.Course, error) {
	l, err := s.lessons.FindByID(ctx, lessonID)
	if err != nil {
		return nil, err
	}
	ch, err := s.chapters.FindByID(ctx, l.ChapterID())
	if err != nil {
		return nil, err
	}
	return s.courses.FindByID(ctx, ch.CourseID())
}

func lessonSubdir(lessonID uint) string {
	return "lesson-" + strconv.FormatUint(uint64(lessonID), 10)
}
