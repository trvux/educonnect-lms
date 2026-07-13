// Package forum là domain của US6.1: Học viên/Giảng viên đăng và trả lời
// câu hỏi trong diễn đàn theo từng khóa học.
package forum

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidCourseID      = errors.New("forum: course id là bắt buộc")
	ErrInvalidAuthorID      = errors.New("forum: author id là bắt buộc")
	ErrEmptyContent         = errors.New("forum: nội dung không được để trống")
	ErrNotFound             = errors.New("forum: không tìm thấy")
	ErrParentCourseMismatch = errors.New("forum: bài trả lời phải cùng khóa học với bài gốc")
)

// Post là 1 bài đăng trong diễn đàn của khóa học — ParentID = nil nghĩa là
// câu hỏi gốc, ParentID != nil nghĩa là trả lời cho bài có ID đó (comment
// lồng nhau, US6.1). AuthorName chỉ được điền khi đọc qua ListByCourse
// (join với bảng users) — không phải dữ liệu gốc của aggregate này.
type Post struct {
	id         uint
	courseID   uint
	authorID   uint
	parentID   *uint
	content    string
	createdAt  time.Time
	authorName string
}

func NewPost(courseID, authorID uint, parentID *uint, content string) (*Post, error) {
	if courseID == 0 {
		return nil, ErrInvalidCourseID
	}
	if authorID == 0 {
		return nil, ErrInvalidAuthorID
	}
	if content == "" {
		return nil, ErrEmptyContent
	}
	return &Post{
		courseID:  courseID,
		authorID:  authorID,
		parentID:  parentID,
		content:   content,
		createdAt: time.Now().UTC(),
	}, nil
}

func Rehydrate(id, courseID, authorID uint, parentID *uint, content string, createdAt time.Time) *Post {
	return &Post{id: id, courseID: courseID, authorID: authorID, parentID: parentID, content: content, createdAt: createdAt}
}

// RehydrateWithAuthor giống Rehydrate nhưng kèm tên tác giả (từ SQL join
// với bảng users) — dùng cho ListByCourse để FE hiển thị mà không cần gọi
// thêm API lấy thông tin user.
func RehydrateWithAuthor(id, courseID, authorID uint, parentID *uint, content string, createdAt time.Time, authorName string) *Post {
	p := Rehydrate(id, courseID, authorID, parentID, content, createdAt)
	p.authorName = authorName
	return p
}

func (p *Post) SetID(id uint) { p.id = id }

func (p *Post) ID() uint             { return p.id }
func (p *Post) CourseID() uint       { return p.courseID }
func (p *Post) AuthorID() uint       { return p.authorID }
func (p *Post) ParentID() *uint      { return p.parentID }
func (p *Post) Content() string      { return p.content }
func (p *Post) CreatedAt() time.Time { return p.createdAt }
func (p *Post) AuthorName() string   { return p.authorName }

type Repository interface {
	Create(ctx context.Context, p *Post) error
	FindByID(ctx context.Context, id uint) (*Post, error)
	// ListByCourse trả về mọi bài đăng (gốc + trả lời) của khóa học, kèm
	// tên tác giả — FE tự dựng cây lồng nhau từ ParentID.
	ListByCourse(ctx context.Context, courseID uint) ([]*Post, error)
}
