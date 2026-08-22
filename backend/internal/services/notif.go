package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// notifCooldownHours prevents the cron from re-alerting about the same issue
// until either the previous notification is marked read or this window passes.
const notifCooldownHours = 24

// Notification mirrors the DB row returned to callers.
type Notification struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`     // low_stock | payment_overdue | reorder_due
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Severity  string    `json:"severity"` // high | med | low
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

// NotificationService runs the inventory+CRM check and manages the
// notifications table. It is designed to be called from a cron goroutine.
type NotificationService struct {
	pool *pgxpool.Pool
	inv  *InventoryService
	crm  *CRMService
}

func NewNotificationService(pool *pgxpool.Pool, inv *InventoryService, crm *CRMService) *NotificationService {
	return &NotificationService{pool: pool, inv: inv, crm: crm}
}

// Check runs both the inventory and CRM checks, inserting a notification for
// each issue that has not been notified about within the cooldown window.
func (s *NotificationService) Check(ctx context.Context) error {
	if err := s.checkInventory(ctx); err != nil {
		return fmt.Errorf("inventory check: %w", err)
	}
	if err := s.checkCRM(ctx); err != nil {
		return fmt.Errorf("crm check: %w", err)
	}
	return nil
}

func (s *NotificationService) checkInventory(ctx context.Context) error {
	stocks, err := s.inv.snapshot(ctx)
	if err != nil {
		return err
	}
	for _, st := range stocks {
		if st.Urgency == "low" {
			continue
		}
		title := fmt.Sprintf("Restock %s", st.Origin)
		body := fmt.Sprintf(
			"Sisa stok %.0f kg (~%.0f hari). Rata-rata penjualan %.1f kg/hari. Urgency: %s.",
			st.StockKg, st.DaysOfCover, st.AvgDailyKg, st.Urgency,
		)
		if err := s.upsert(ctx, "low_stock", title, body, st.Urgency); err != nil {
			return err
		}
	}
	return nil
}

func (s *NotificationService) checkCRM(ctx context.Context) error {
	followUps, err := s.crm.snapshotFollowUps(ctx)
	if err != nil {
		return err
	}
	for _, f := range followUps {
		switch {
		case f.PaymentStatus == "overdue":
			title := fmt.Sprintf("Tagih pembayaran: %s", f.ShopName)
			body := fmt.Sprintf(
				"Status overdue. Terakhir order %.0f hari lalu. Prediksi reorder: %s.",
				f.DaysSinceOrder, f.PredictedReorder,
			)
			if err := s.upsert(ctx, "payment_overdue", title, body, "high"); err != nil {
				return err
			}

		case f.NeedsFollowUp:
			sev := "low"
			if f.LatePayRisk {
				sev = "med"
			}
			title := fmt.Sprintf("Follow up: %s", f.ShopName)
			body := fmt.Sprintf(
				"Reorder rata-rata setiap %.0f hari. Sudah %.0f hari sejak order terakhir.",
				f.AvgIntervalDays, f.DaysSinceOrder,
			)
			if err := s.upsert(ctx, "reorder_due", title, body, sev); err != nil {
				return err
			}
		}
	}
	return nil
}

// upsert inserts a notification only when there is no existing unread one with
// the same (kind, title) within the cooldown window — prevents spam.
func (s *NotificationService) upsert(ctx context.Context, kind, title, body, severity string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO notifications (kind, title, body, severity)
		SELECT $1, $2, $3, $4
		WHERE NOT EXISTS (
			SELECT 1 FROM notifications
			WHERE kind = $1
			  AND title = $2
			  AND read  = false
			  AND created_at > NOW() - make_interval(hours => $5)
		)`,
		kind, title, body, severity, notifCooldownHours,
	)
	return err
}

// List returns recent notifications, newest first.
func (s *NotificationService) List(ctx context.Context, limit int) ([]Notification, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, kind, title, body, severity, read, created_at
		FROM   notifications
		ORDER  BY created_at DESC
		LIMIT  $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.Kind, &n.Title, &n.Body, &n.Severity, &n.Read, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// UnreadCount returns the number of unread notifications.
func (s *NotificationService) UnreadCount(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE read = false`).Scan(&n)
	return n, err
}

// MarkRead marks a single notification as read.
func (s *NotificationService) MarkRead(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `UPDATE notifications SET read = true WHERE id = $1::uuid`, id)
	return err
}

// MarkAllRead marks every unread notification as read.
func (s *NotificationService) MarkAllRead(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `UPDATE notifications SET read = true WHERE read = false`)
	return err
}
