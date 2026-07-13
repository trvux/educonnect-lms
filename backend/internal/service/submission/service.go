// Package submission là service layer của US5.2 (nộp bài) và US5.3 (chấm
// điểm): Học viên nộp bài tập tự luận/trắc nghiệm trước hạn; bài trắc
// nghiệm được tự động chấm ngay khi nộp, bài tự luận chờ giảng viên sở hữu
// khóa học chấm thủ công.
package submission

import (
	"context"
	"errors"
	"time"

	"educonnect-lms/backend/internal/domain/assignment"
	"educonnect-lms/backend/internal/domain/course"
	"educonnect-lms/backend/internal/domain/curriculum"
	"educonnect-lms/backend/internal/domain/submission"
)

// AssignmentGetter là tập con method của *assignmentservice.Service mà
// service này cần để kiểm tra hạn nộp/số câu hỏi/đáp án đúng.
type AssignmentGetter interface {
	Get(ctx context.Context, id uint) (*assignment.Assignment, error)
}

// CourseGetter là tập con method của course.Repository, dùng để xác nhận
// giảng viên chấm điểm chính là chủ sở hữu khóa học (US5.3).
type CourseGetter interface {
	FindByID(ctx context.Context, id uint) (*course.Course, error)
}

type Service struct {
	submissions submission.Repository
	assignments AssignmentGetter
	lessons     curriculum.LessonRepository
	chapters    curriculum.ChapterRepository
	courses     CourseGetter
}

func NewService(
	submissions submission.Repository,
	assignments AssignmentGetter,
	lessons curriculum.LessonRepository,
	chapters curriculum.ChapterRepository,
	courses CourseGetter,
) *Service {
	return &Service{
		submissions: submissions,
		assignments: assignments,
		lessons:     lessons,
		chapters:    chapters,
		courses:     courses,
	}
}

// Submit hiện thực US5.2: xác nhận Assignment tồn tại, chưa quá hạn nộp,
// học viên chưa nộp bài này trước đó, và (với bài trắc nghiệm) số đáp án
// khớp số câu hỏi — trước khi ghi nhận bài làm. Bài trắc nghiệm được tự
// động chấm điểm ngay (US5.3) bằng cách so khớp Answers với đáp án đúng.
func (s *Service) Submit(ctx context.Context, assignmentID, studentID uint, content string, answers []int) (*submission.Submission, error) {
	a, err := s.assignments.Get(ctx, assignmentID)
	if err != nil {
		return nil, err
	}

	if a.DueAt() != nil && time.Now().UTC().After(*a.DueAt()) {
		return nil, submission.ErrPastDue
	}

	if a.Kind() == assignment.TypeQuiz && len(answers) != len(a.Questions()) {
		return nil, submission.ErrAnswerCountMismatch
	}

	_, err = s.submissions.FindByAssignmentAndStudent(ctx, assignmentID, studentID)
	if err == nil {
		return nil, submission.ErrAlreadySubmitted
	}
	if !errors.Is(err, submission.ErrNotFound) {
		return nil, err
	}

	sub, err := submission.NewSubmission(assignmentID, studentID, content, answers)
	if err != nil {
		return nil, err
	}

	if a.Kind() == assignment.TypeQuiz {
		if err := sub.Grade(autoScore(a.Questions(), answers), ""); err != nil {
			return nil, err
		}
	}

	if err := s.submissions.Create(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

// Grade hiện thực US5.3: giảng viên chấm điểm 1 bài nộp (chủ yếu dùng cho
// bài tự luận — bài trắc nghiệm đã được tự động chấm khi nộp). Chỉ giảng
// viên sở hữu khóa học (qua chuỗi submission → assignment → lesson →
// chapter → course) hoặc quản trị viên mới được chấm.
func (s *Service) Grade(ctx context.Context, submissionID, graderID uint, isAdmin bool, score float64, feedback string) (*submission.Submission, error) {
	sub, err := s.submissions.FindByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}

	if !isAdmin {
		owns, err := s.teacherOwnsSubmission(ctx, sub, graderID)
		if err != nil {
			return nil, err
		}
		if !owns {
			return nil, submission.ErrNotFound // không tiết lộ sự tồn tại của bài nộp khóa học người khác
		}
	}

	if err := sub.Grade(score, feedback); err != nil {
		return nil, err
	}
	if err := s.submissions.Update(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

// ListByAssignment phục vụ trang chấm điểm của giảng viên (US5.3).
func (s *Service) ListByAssignment(ctx context.Context, assignmentID uint) ([]*submission.Submission, error) {
	return s.submissions.ListByAssignment(ctx, assignmentID)
}

func (s *Service) teacherOwnsSubmission(ctx context.Context, sub *submission.Submission, teacherID uint) (bool, error) {
	a, err := s.assignments.Get(ctx, sub.AssignmentID())
	if err != nil {
		return false, err
	}
	l, err := s.lessons.FindByID(ctx, a.LessonID())
	if err != nil {
		return false, err
	}
	ch, err := s.chapters.FindByID(ctx, l.ChapterID())
	if err != nil {
		return false, err
	}
	c, err := s.courses.FindByID(ctx, ch.CourseID())
	if err != nil {
		return false, err
	}
	return c.TeacherID() == teacherID, nil
}

// autoScore tính điểm 0-10 theo tỉ lệ câu trả lời đúng cho bài trắc nghiệm.
func autoScore(questions []assignment.Question, answers []int) float64 {
	correct := 0
	for i, q := range questions {
		if i < len(answers) && answers[i] == q.CorrectIndex {
			correct++
		}
	}
	return float64(correct) / float64(len(questions)) * 10
}
