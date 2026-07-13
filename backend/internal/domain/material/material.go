// Package material là domain của US4.1: Giảng viên tải lên tài liệu/video
// bài giảng (PDF, slide...) gắn với 1 Lesson cụ thể.
package material

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrInvalidLessonID     = errors.New("material: lesson id là bắt buộc")
	ErrEmptyFileName       = errors.New("material: tên file là bắt buộc")
	ErrEmptyFilePath       = errors.New("material: đường dẫn lưu trữ là bắt buộc")
	ErrNotFound            = errors.New("material: không tìm thấy")
	ErrUnsupportedFileType = errors.New("material: định dạng file không được hỗ trợ (chỉ nhận PDF/Word/Excel/PowerPoint/video/nén)")
)

// FileType là nhóm định dạng tài liệu (US4.4), phân loại theo phần mở rộng
// tên file — không sniff nội dung thật của file (magic bytes), vì các định
// dạng Office hiện đại (.docx/.xlsx/.pptx) đều là file ZIP bên trong, Go
// stdlib (http.DetectContentType) không phân biệt được chúng với .zip
// thường; phân loại theo tên là cách thực tế và đơn giản nhất ở quy mô
// dự án này.
type FileType string

const (
	FileTypePDF     FileType = "pdf"
	FileTypeDoc     FileType = "doc"
	FileTypeExcel   FileType = "excel"
	FileTypePPT     FileType = "ppt"
	FileTypeVideo   FileType = "video"
	FileTypeArchive FileType = "archive"
)

var extensionToFileType = map[string]FileType{
	".pdf":  FileTypePDF,
	".doc":  FileTypeDoc,
	".docx": FileTypeDoc,
	".xls":  FileTypeExcel,
	".xlsx": FileTypeExcel,
	".ppt":  FileTypePPT,
	".pptx": FileTypePPT,
	".mp4":  FileTypeVideo,
	".webm": FileTypeVideo,
	".mov":  FileTypeVideo,
	".zip":  FileTypeArchive,
	".rar":  FileTypeArchive,
	".7z":   FileTypeArchive,
}

// classifyFileType xác định FileType từ tên file; trả ErrUnsupportedFileType
// nếu phần mở rộng không nằm trong whitelist (US4.4).
func classifyFileType(fileName string) (FileType, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	t, ok := extensionToFileType[ext]
	if !ok {
		return "", ErrUnsupportedFileType
	}
	return t, nil
}

// Material đại diện cho 1 file tài liệu (PDF/slide/video) đã upload,
// gắn với 1 Lesson. FilePath là đường dẫn lưu trữ nội bộ (US4.1 dùng local
// disk storage cho bản demo; có thể thay bằng object storage sau này vì
// domain không phụ thuộc trực tiếp implementation lưu trữ — xem
// internal/platform/storage.FileStorage).
type Material struct {
	id         uint
	lessonID   uint
	fileName   string
	filePath   string
	fileType   FileType
	uploadedAt time.Time
}

func NewMaterial(lessonID uint, fileName, filePath string) (*Material, error) {
	if lessonID == 0 {
		return nil, ErrInvalidLessonID
	}
	if fileName == "" {
		return nil, ErrEmptyFileName
	}
	if filePath == "" {
		return nil, ErrEmptyFilePath
	}
	fileType, err := classifyFileType(fileName)
	if err != nil {
		return nil, err
	}
	return &Material{lessonID: lessonID, fileName: fileName, filePath: filePath, fileType: fileType, uploadedAt: time.Now().UTC()}, nil
}

func Rehydrate(id, lessonID uint, fileName, filePath string, fileType FileType, uploadedAt time.Time) *Material {
	return &Material{id: id, lessonID: lessonID, fileName: fileName, filePath: filePath, fileType: fileType, uploadedAt: uploadedAt}
}

func (m *Material) SetID(id uint) { m.id = id }

func (m *Material) ID() uint              { return m.id }
func (m *Material) LessonID() uint        { return m.lessonID }
func (m *Material) FileName() string      { return m.fileName }
func (m *Material) FilePath() string      { return m.filePath }
func (m *Material) FileType() FileType    { return m.fileType }
func (m *Material) UploadedAt() time.Time { return m.uploadedAt }

type Repository interface {
	Create(ctx context.Context, m *Material) error
	FindByID(ctx context.Context, id uint) (*Material, error)
	ListByLesson(ctx context.Context, lessonID uint) ([]*Material, error)
	Delete(ctx context.Context, id uint) error
}
