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

	_, err = material.NewMaterial(0, "a.pdf", "path")
	assert.ErrorIs(t, err, material.ErrInvalidLessonID)

	_, err = material.NewMaterial(1, "", "path")
	assert.ErrorIs(t, err, material.ErrEmptyFileName)

	_, err = material.NewMaterial(1, "a.pdf", "")
	assert.ErrorIs(t, err, material.ErrEmptyFilePath)
}
