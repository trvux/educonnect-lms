package assignment_test

import (
	"testing"

	"educonnect-lms/backend/internal/domain/assignment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validQuestions() []assignment.Question {
	return []assignment.Question{
		{Content: "1 + 1 = ?", Options: []string{"1", "2", "3"}, CorrectIndex: 1},
	}
}

func TestNewAssignment_Essay(t *testing.T) {
	a, err := assignment.NewAssignment(1, "Bai tap tuan 1", "Nop file .go", assignment.TypeEssay, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "Bai tap tuan 1", a.Title())
	assert.Equal(t, assignment.TypeEssay, a.Kind())
	assert.Empty(t, a.Questions())
}

func TestNewAssignment_Quiz(t *testing.T) {
	a, err := assignment.NewAssignment(1, "Trac nghiem chuong 1", "", assignment.TypeQuiz, validQuestions(), nil)
	require.NoError(t, err)
	assert.Equal(t, assignment.TypeQuiz, a.Kind())
	assert.Len(t, a.Questions(), 1)
}

func TestNewAssignment_Validation(t *testing.T) {
	_, err := assignment.NewAssignment(0, "Bai tap", "", assignment.TypeEssay, nil, nil)
	assert.ErrorIs(t, err, assignment.ErrInvalidLessonID)

	_, err = assignment.NewAssignment(1, "", "", assignment.TypeEssay, nil, nil)
	assert.ErrorIs(t, err, assignment.ErrEmptyTitle)

	_, err = assignment.NewAssignment(1, "Bai tap", "", "khong-hop-le", nil, nil)
	assert.ErrorIs(t, err, assignment.ErrInvalidType)

	_, err = assignment.NewAssignment(1, "Trac nghiem", "", assignment.TypeQuiz, nil, nil)
	assert.ErrorIs(t, err, assignment.ErrQuizNeedsQuestions)
}

func TestNewAssignment_InvalidQuestion(t *testing.T) {
	badContent := []assignment.Question{{Content: "", Options: []string{"a", "b"}, CorrectIndex: 0}}
	_, err := assignment.NewAssignment(1, "Trac nghiem", "", assignment.TypeQuiz, badContent, nil)
	assert.ErrorIs(t, err, assignment.ErrInvalidQuestion)

	tooFewOptions := []assignment.Question{{Content: "cau 1", Options: []string{"a"}, CorrectIndex: 0}}
	_, err = assignment.NewAssignment(1, "Trac nghiem", "", assignment.TypeQuiz, tooFewOptions, nil)
	assert.ErrorIs(t, err, assignment.ErrInvalidQuestion)

	outOfRange := []assignment.Question{{Content: "cau 1", Options: []string{"a", "b"}, CorrectIndex: 5}}
	_, err = assignment.NewAssignment(1, "Trac nghiem", "", assignment.TypeQuiz, outOfRange, nil)
	assert.ErrorIs(t, err, assignment.ErrInvalidQuestion)
}
