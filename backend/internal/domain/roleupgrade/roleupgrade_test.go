package roleupgrade_test

import (
	"testing"

	"educonnect-lms/backend/internal/domain/roleupgrade"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequest(t *testing.T) {
	r, err := roleupgrade.NewRequest(1, "Em muốn dạy khóa Golang cho các bạn khác")
	require.NoError(t, err)
	assert.Equal(t, uint(1), r.UserID())
	assert.Equal(t, roleupgrade.StatusPending, r.Status())

	_, err = roleupgrade.NewRequest(1, "")
	assert.ErrorIs(t, err, roleupgrade.ErrEmptyReason)
}

func TestRequest_Approve(t *testing.T) {
	r, err := roleupgrade.NewRequest(1, "ly do")
	require.NoError(t, err)

	require.NoError(t, r.Approve(99))
	assert.Equal(t, roleupgrade.StatusApproved, r.Status())
	require.NotNil(t, r.ReviewedBy())
	assert.Equal(t, uint(99), *r.ReviewedBy())

	// Đã duyệt rồi thì không duyệt/từ chối lại được.
	assert.ErrorIs(t, r.Approve(99), roleupgrade.ErrNotPending)
	assert.ErrorIs(t, r.Reject(99), roleupgrade.ErrNotPending)
}

func TestRequest_Reject(t *testing.T) {
	r, err := roleupgrade.NewRequest(1, "ly do")
	require.NoError(t, err)

	require.NoError(t, r.Reject(99))
	assert.Equal(t, roleupgrade.StatusRejected, r.Status())
}
