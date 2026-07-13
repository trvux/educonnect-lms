package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"educonnect-lms/backend/internal/domain/notification"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

// CreateMany ghi toàn bộ thông báo (fan-out tới nhiều học viên) trong 1
// giao dịch — hoặc tất cả thành công, hoặc không có thông báo nào được tạo.
func (r *NotificationRepository) CreateMany(ctx context.Context, notifications []*notification.Notification) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres: mở giao dịch tạo notification lỗi: %w", err)
	}
	defer tx.Rollback(ctx)

	const q = `
		INSERT INTO notifications (recipient_id, course_id, title, message, read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	for _, n := range notifications {
		var id uint
		err := tx.QueryRow(ctx, q, n.RecipientID(), n.CourseID(), n.Title(), n.Message(), n.Read(), n.CreatedAt()).Scan(&id)
		if err != nil {
			return fmt.Errorf("postgres: tạo notification lỗi: %w", err)
		}
		n.SetID(id)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres: commit giao dịch tạo notification lỗi: %w", err)
	}
	return nil
}

func (r *NotificationRepository) FindByID(ctx context.Context, id uint) (*notification.Notification, error) {
	const q = `SELECT id, recipient_id, course_id, title, message, read, created_at FROM notifications WHERE id = $1`
	n, err := scanNotification(r.pool.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, notification.ErrNotFound
		}
		return nil, fmt.Errorf("postgres: đọc dữ liệu notification lỗi: %w", err)
	}
	return n, nil
}

func (r *NotificationRepository) ListByRecipient(ctx context.Context, recipientID uint) ([]*notification.Notification, error) {
	const q = `
		SELECT id, recipient_id, course_id, title, message, read, created_at
		FROM notifications WHERE recipient_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, recipientID)
	if err != nil {
		return nil, fmt.Errorf("postgres: lấy danh sách notification lỗi: %w", err)
	}
	defer rows.Close()

	var result []*notification.Notification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres: đọc dòng notification lỗi: %w", err)
		}
		result = append(result, n)
	}
	return result, rows.Err()
}

func (r *NotificationRepository) CountUnread(ctx context.Context, recipientID uint) (int, error) {
	const q = `SELECT count(*) FROM notifications WHERE recipient_id = $1 AND read = false`
	var count int
	if err := r.pool.QueryRow(ctx, q, recipientID).Scan(&count); err != nil {
		return 0, fmt.Errorf("postgres: đếm notification chưa đọc lỗi: %w", err)
	}
	return count, nil
}

func (r *NotificationRepository) Update(ctx context.Context, n *notification.Notification) error {
	const q = `UPDATE notifications SET read = $2 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, n.ID(), n.Read())
	if err != nil {
		return fmt.Errorf("postgres: cập nhật notification lỗi: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return notification.ErrNotFound
	}
	return nil
}

func scanNotification(row rowScanner) (*notification.Notification, error) {
	var (
		id, recipientID, courseID uint
		title, message            string
		read                      bool
		createdAt                 time.Time
	)
	if err := row.Scan(&id, &recipientID, &courseID, &title, &message, &read, &createdAt); err != nil {
		return nil, err
	}
	return notification.Rehydrate(id, recipientID, courseID, title, message, read, createdAt), nil
}
