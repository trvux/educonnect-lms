package material_test

import (
	"testing"

	"educonnect-lms/backend/internal/domain/material"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMaterial(t *testing.T) {
	m, err := material.NewMaterial(1, "slide-bai-1.pdf", "uploads/lesson-1/slide-bai-1.pdf")
	require.NoError(t, err)
	assert.Equal(t, "slide-bai-1.pdf", m.FileName())
	assert.Equal(t, material.FileTypePDF, m.FileType())

	_, err = material.NewMaterial(0, "a.pdf", "path")
	assert.ErrorIs(t, err, material.ErrInvalidLessonID)

	_, err = material.NewMaterial(1, "", "path")
	assert.ErrorIs(t, err, material.ErrEmptyFileName)

	_, err = material.NewMaterial(1, "a.pdf", "")
	assert.ErrorIs(t, err, material.ErrEmptyFilePath)
}

func TestNewMaterial_FileTypeWhitelist(t *testing.T) {
	tests := []struct {
		fileName string
		want     material.FileType
	}{
		{"slide.pdf", material.FileTypePDF},
		{"BaiGiang.DOC", material.FileTypeDoc}, // không phân biệt hoa/thường
		{"report.docx", material.FileTypeDoc},
		{"diem.xls", material.FileTypeExcel},
		{"diem.xlsx", material.FileTypeExcel},
		{"slide.ppt", material.FileTypePPT},
		{"slide.pptx", material.FileTypePPT},
		{"baigiang.mp4", material.FileTypeVideo},
		{"baigiang.webm", material.FileTypeVideo},
		{"baigiang.mov", material.FileTypeVideo},
		{"mon-hoc.zip", material.FileTypeArchive},
		{"mon-hoc.rar", material.FileTypeArchive},
		{"mon-hoc.7z", material.FileTypeArchive},
	}
	for _, tt := range tests {
		t.Run(tt.fileName, func(t *testing.T) {
			m, err := material.NewMaterial(1, tt.fileName, "path/"+tt.fileName)
			require.NoError(t, err)
			assert.Equal(t, tt.want, m.FileType())
		})
	}

	for _, bad := range []string{"virus.exe", "script.sh", "noext"} {
		t.Run(bad, func(t *testing.T) {
			_, err := material.NewMaterial(1, bad, "path/"+bad)
			assert.ErrorIs(t, err, material.ErrUnsupportedFileType)
		})
	}
}
