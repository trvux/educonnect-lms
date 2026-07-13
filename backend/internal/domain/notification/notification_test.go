package notification_test

import (
	"testing"

	"educonnect-lms/backend/internal/domain/notification"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNotification(t *testing.T) {
	n, err := notification.NewNotification(1, 2, "Bai tap moi", "Co bai tap moi trong khoa hoc")
	require.NoError(t, err)
	assert.False(t, n.Read())
	assert.Equal(t, "Bai tap moi", n.Title())
}

func TestNewNotification_Validation(t *testing.T) {
	_, err := notification.NewNotification(0, 2, "tieu de", "")
	assert.ErrorIs(t, err, notification.ErrInvalidRecipientID)

	_, err = notification.NewNotification(1, 0, "tieu de", "")
	assert.ErrorIs(t, err, notification.ErrInvalidCourseID)

	_, err = notification.NewNotification(1, 2, "", "")
	assert.ErrorIs(t, err, notification.ErrEmptyTitle)
}

func TestNotification_MarkRead(t *testing.T) {
	n, err := notification.NewNotification(1, 2, "tieu de", "")
	require.NoError(t, err)
	assert.False(t, n.Read())

	n.MarkRead()
	assert.True(t, n.Read())
}
