// Package submission là domain của US5.2: Học viên nộp bài tập (tự luận)
// hoặc làm bài trắc nghiệm trước hạn, gắn với 1 Assignment.
package submission

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidAssignmentID = errors.New("submission: assignment id là bắt buộc")
	ErrInvalidStudentID    = errors.New("submission: student id là bắt buộc")
	ErrEmptySubmission     = errors.New("submission: phải có nội dung bài làm (tự luận) hoặc đáp án (trắc nghiệm)")
	ErrAlreadySubmitted    = errors.New("submission: học viên đã nộp bài này rồi")
	ErrPastDue             = errors.New("submission: đã quá hạn nộp bài")
	ErrAnswerCountMismatch = errors.New("submission: số lượng đáp án không khớp số câu hỏi")
	ErrNotFound            = errors.New("submission: không tìm thấy")
	ErrInvalidScore        = errors.New("submission: điểm phải nằm trong khoảng 0-10")
	// ErrTimeExpired (US5.4): bài trắc nghiệm có giới hạn thời gian, học
	// viên nộp sau khi đã hết giờ tính từ lúc bắt đầu làm (quizattempt).
	ErrTimeExpired = errors.New("submission: đã hết thời gian làm bài trắc nghiệm")
)

// Submission là bài làm của 1 học viên cho 1 Assignment. Content dùng cho
// bài tự luận (essay), Answers (chỉ số lựa chọn theo từng câu hỏi) dùng cho
// bài trắc nghiệm (quiz) — việc đối chiếu đúng loại với Assignment thuộc về
// service layer vì cần đọc aggregate Assignment (xem service/submission).
//
// Score/Feedback/GradedAt phục vụ US5.3: bài trắc nghiệm được service tự
// động chấm ngay khi nộp (so khớp Answers với đáp án đúng của Assignment),
// bài tự luận chờ giảng viên chấm thủ công qua Grade().
type Submission struct {
	id           uint
	assignmentID uint
	studentID    uint
	content      string
	answers      []int
	submittedAt  time.Time
	score        *float64
	feedback     string
	gradedAt     *time.Time
}

func NewSubmission(assignmentID, studentID uint, content string, answers []int) (*Submission, error) {
	if assignmentID == 0 {
		return nil, ErrInvalidAssignmentID
	}
	if studentID == 0 {
		return nil, ErrInvalidStudentID
	}
	if content == "" && len(answers) == 0 {
		return nil, ErrEmptySubmission
	}
	return &Submission{
		assignmentID: assignmentID,
		studentID:    studentID,
		content:      content,
		answers:      answers,
		submittedAt:  time.Now().UTC(),
	}, nil
}

func Rehydrate(id, assignmentID, studentID uint, content string, answers []int, submittedAt time.Time, score *float64, feedback string, gradedAt *time.Time) *Submission {
	return &Submission{
		id:           id,
		assignmentID: assignmentID,
		studentID:    studentID,
		content:      content,
		answers:      answers,
		submittedAt:  submittedAt,
		score:        score,
		feedback:     feedback,
		gradedAt:     gradedAt,
	}
}

func (s *Submission) SetID(id uint) { s.id = id }

// Grade ghi nhận điểm/nhận xét của giảng viên (chấm thủ công bài tự luận)
// hoặc kết quả chấm tự động (bài trắc nghiệm, xem service/submission.Submit).
func (s *Submission) Grade(score float64, feedback string) error {
	if score < 0 || score > 10 {
		return ErrInvalidScore
	}
	s.score = &score
	s.feedback = feedback
	now := time.Now().UTC()
	s.gradedAt = &now
	return nil
}

func (s *Submission) IsGraded() bool { return s.score != nil }

func (s *Submission) ID() uint               { return s.id }
func (s *Submission) AssignmentID() uint     { return s.assignmentID }
func (s *Submission) StudentID() uint        { return s.studentID }
func (s *Submission) Content() string        { return s.content }
func (s *Submission) Answers() []int         { return s.answers }
func (s *Submission) SubmittedAt() time.Time { return s.submittedAt }
func (s *Submission) Score() *float64        { return s.score }
func (s *Submission) Feedback() string       { return s.feedback }
func (s *Submission) GradedAt() *time.Time   { return s.gradedAt }

type Repository interface {
	Create(ctx context.Context, s *Submission) error
	FindByID(ctx context.Context, id uint) (*Submission, error)
	// FindByAssignmentAndStudent trả về ErrNotFound nếu học viên chưa nộp —
	// dùng để chặn nộp trùng (US5.2).
	FindByAssignmentAndStudent(ctx context.Context, assignmentID, studentID uint) (*Submission, error)
	ListByAssignment(ctx context.Context, assignmentID uint) ([]*Submission, error)
	// Update lưu lại điểm/nhận xét sau khi chấm (US5.3).
	Update(ctx context.Context, s *Submission) error
}
