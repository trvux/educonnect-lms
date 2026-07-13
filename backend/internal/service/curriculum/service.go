// Package curriculum là service layer của US2.2: Giảng viên tạo/sắp xếp
// chương và bài học trong khóa học; US4.6: sửa/xóa chương và bài học, chỉ
// giảng viên sở hữu khóa học (hoặc admin) mới được phép.
package curriculum

import (
	"context"

	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/user"
)

// CourseGetter là tập con method của course.Repository (US4.6: xác nhận
// giảng viên sở hữu khóa học chứa chương/bài học muốn sửa/xóa).
type CourseGetter interface {
	FindByID(ctx context.Context, id uint) (*course.Course, error)
}

type Service struct {
	chapters curriculum.ChapterRepository
	lessons  curriculum.LessonRepository
	courses  CourseGetter
}

func NewService(chapters curriculum.ChapterRepository, lessons curriculum.LessonRepository, courses CourseGetter) *Service {
	return &Service{chapters: chapters, lessons: lessons, courses: courses}
}

// CreateChapter tạo 1 chương mới, tự động xếp vào cuối danh sách chương
// hiện có của khóa học (US2.2).
func (s *Service) CreateChapter(ctx context.Context, courseID uint, title string) (*curriculum.Chapter, error) {
	count, err := s.chapters.CountByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	ch, err := curriculum.NewChapter(courseID, title, count)
	if err != nil {
		return nil, err
	}
	if err := s.chapters.Create(ctx, ch); err != nil {
		return nil, err
	}
	return ch, nil
}

func (s *Service) ListChapters(ctx context.Context, courseID uint) ([]*curriculum.Chapter, error) {
	return s.chapters.ListByCourse(ctx, courseID)
}

// RenameChapter hiện thực US4.6.
func (s *Service) RenameChapter(ctx context.Context, chapterID uint, title string, userID uint, role user.Role) (*curriculum.Chapter, error) {
	ch, err := s.chapters.FindByID(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	owns, err := s.ownsCourse(ctx, ch.CourseID(), userID, role)
	if err != nil {
		return nil, err
	}
	if !owns {
		return nil, curriculum.ErrChapterNotFound // không tiết lộ khóa học người khác tồn tại
	}
	if err := ch.Rename(title); err != nil {
		return nil, err
	}
	if err := s.chapters.Update(ctx, ch); err != nil {
		return nil, err
	}
	return ch, nil
}

// DeleteChapter hiện thực US4.6. Xóa thất bại (ErrChapterNotEmpty) nếu
// chương còn bài học bên trong — chặn ở tầng DB (khóa ngoại), dịch sang
// lỗi domain có ý nghĩa ở tầng repository.
func (s *Service) DeleteChapter(ctx context.Context, chapterID, userID uint, role user.Role) error {
	ch, err := s.chapters.FindByID(ctx, chapterID)
	if err != nil {
		return err
	}
	owns, err := s.ownsCourse(ctx, ch.CourseID(), userID, role)
	if err != nil {
		return err
	}
	if !owns {
		return curriculum.ErrChapterNotFound
	}
	return s.chapters.Delete(ctx, chapterID)
}

// CreateLesson tạo 1 bài học mới trong 1 chương, tự xếp cuối danh sách bài
// học hiện có của chương đó (US2.2).
func (s *Service) CreateLesson(ctx context.Context, chapterID uint, title string) (*curriculum.Lesson, error) {
	// Đảm bảo chapter tồn tại trước khi tạo lesson con.
	if _, err := s.chapters.FindByID(ctx, chapterID); err != nil {
		return nil, err
	}
	count, err := s.lessons.CountByChapter(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	l, err := curriculum.NewLesson(chapterID, title, count)
	if err != nil {
		return nil, err
	}
	if err := s.lessons.Create(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *Service) ListLessons(ctx context.Context, chapterID uint) ([]*curriculum.Lesson, error) {
	return s.lessons.ListByChapter(ctx, chapterID)
}

// RenameLesson hiện thực US4.6.
func (s *Service) RenameLesson(ctx context.Context, lessonID uint, title string, userID uint, role user.Role) (*curriculum.Lesson, error) {
	l, err := s.lessons.FindByID(ctx, lessonID)
	if err != nil {
		return nil, err
	}
	owns, err := s.ownsLesson(ctx, l, userID, role)
	if err != nil {
		return nil, err
	}
	if !owns {
		return nil, curriculum.ErrLessonNotFound
	}
	if err := l.Rename(title); err != nil {
		return nil, err
	}
	if err := s.lessons.Update(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

// DeleteLesson hiện thực US4.6. Xóa thất bại (ErrLessonNotEmpty) nếu bài
// học còn tài liệu/bài tập bên trong.
func (s *Service) DeleteLesson(ctx context.Context, lessonID, userID uint, role user.Role) error {
	l, err := s.lessons.FindByID(ctx, lessonID)
	if err != nil {
		return err
	}
	owns, err := s.ownsLesson(ctx, l, userID, role)
	if err != nil {
		return err
	}
	if !owns {
		return curriculum.ErrLessonNotFound
	}
	return s.lessons.Delete(ctx, lessonID)
}

// ReorderChapters hiện thực US4.7: giảng viên kéo-thả sắp xếp lại thứ tự
// chương trong khóa học. ids phải là hoán vị đầy đủ của toàn bộ ID chương
// hiện có của khóa học (không thiếu/thừa/lặp) — thứ tự xuất hiện trong ids
// quyết định Position mới (0, 1, 2, ...).
func (s *Service) ReorderChapters(ctx context.Context, courseID uint, ids []uint, userID uint, role user.Role) ([]*curriculum.Chapter, error) {
	owns, err := s.ownsCourse(ctx, courseID, userID, role)
	if err != nil {
		return nil, err
	}
	if !owns {
		return nil, curriculum.ErrChapterNotFound
	}
	existing, err := s.chapters.ListByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	if len(ids) != len(existing) {
		return nil, curriculum.ErrInvalidChapterOrder
	}
	byID := make(map[uint]*curriculum.Chapter, len(existing))
	for _, ch := range existing {
		byID[ch.ID()] = ch
	}
	ordered := make([]*curriculum.Chapter, 0, len(ids))
	seen := make(map[uint]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return nil, curriculum.ErrInvalidChapterOrder
		}
		ch, ok := byID[id]
		if !ok {
			return nil, curriculum.ErrInvalidChapterOrder
		}
		seen[id] = true
		ordered = append(ordered, ch)
	}
	for i, ch := range ordered {
		ch.Reorder(i)
		if err := s.chapters.Update(ctx, ch); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

// ReorderLessons hiện thực US4.7, cùng logic như ReorderChapters nhưng cho
// bài học trong 1 chương.
func (s *Service) ReorderLessons(ctx context.Context, chapterID uint, ids []uint, userID uint, role user.Role) ([]*curriculum.Lesson, error) {
	ch, err := s.chapters.FindByID(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	owns, err := s.ownsCourse(ctx, ch.CourseID(), userID, role)
	if err != nil {
		return nil, err
	}
	if !owns {
		return nil, curriculum.ErrLessonNotFound
	}
	existing, err := s.lessons.ListByChapter(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	if len(ids) != len(existing) {
		return nil, curriculum.ErrInvalidLessonOrder
	}
	byID := make(map[uint]*curriculum.Lesson, len(existing))
	for _, l := range existing {
		byID[l.ID()] = l
	}
	ordered := make([]*curriculum.Lesson, 0, len(ids))
	seen := make(map[uint]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return nil, curriculum.ErrInvalidLessonOrder
		}
		l, ok := byID[id]
		if !ok {
			return nil, curriculum.ErrInvalidLessonOrder
		}
		seen[id] = true
		ordered = append(ordered, l)
	}
	for i, l := range ordered {
		l.Reorder(i)
		if err := s.lessons.Update(ctx, l); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

func (s *Service) ownsLesson(ctx context.Context, l *curriculum.Lesson, userID uint, role user.Role) (bool, error) {
	ch, err := s.chapters.FindByID(ctx, l.ChapterID())
	if err != nil {
		return false, err
	}
	return s.ownsCourse(ctx, ch.CourseID(), userID, role)
}

// ownsCourse: admin luôn được phép; giảng viên chỉ được phép nếu sở hữu
// khóa học chứa chương/bài học.
func (s *Service) ownsCourse(ctx context.Context, courseID, userID uint, role user.Role) (bool, error) {
	if role == user.RoleAdmin {
		return true, nil
	}
	c, err := s.courses.FindByID(ctx, courseID)
	if err != nil {
		return false, err
	}
	return c.TeacherID() == userID, nil
}
