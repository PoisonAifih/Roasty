package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poisonaifih/roasty/backend/internal/models"
)

type CRMService struct {
	pool *pgxpool.Pool
	ai   *AIClient
}

func NewCRMService(pool *pgxpool.Pool, ai *AIClient) *CRMService {
	return &CRMService{pool: pool, ai: ai}
}

func (s *CRMService) FollowUps(ctx context.Context) ([]models.CRMFollowUp, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT cs.id::text, cs.name, cs.payment_status, COALESCE(cs.phone, ''),
		       array_agg(o.ordered_at ORDER BY o.ordered_at) AS dates
		FROM coffee_shops cs
		JOIN orders o ON o.shop_id = cs.id
		GROUP BY cs.id, cs.name, cs.payment_status, cs.phone
		ORDER BY cs.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.CRMFollowUp
	today := time.Now().Truncate(24 * time.Hour)
	for rows.Next() {
		var f models.CRMFollowUp
		var dates []time.Time
		if err := rows.Scan(&f.ShopID, &f.ShopName, &f.PaymentStatus, &f.Phone, &dates); err != nil {
			return nil, err
		}
		f.AvgIntervalDays = avgInterval(dates)
		last := dates[len(dates)-1]
		f.DaysSinceOrder = today.Sub(last.Truncate(24 * time.Hour)).Hours() / 24
		next := last.Add(time.Duration(f.AvgIntervalDays*24) * time.Hour)
		f.PredictedReorder = next.Format("2006-01-02")
		f.NeedsFollowUp = f.DaysSinceOrder >= f.AvgIntervalDays-3
		f.LatePayRisk = f.PaymentStatus == "overdue" || f.PaymentStatus == "credit"
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	s.fillCRMMessages(ctx, out)
	return out, nil
}

// fillCRMMessages populates Message for every shop concurrently, capped at
// aiConcurrency in-flight requests. Each goroutine owns one index.
func (s *CRMService) fillCRMMessages(ctx context.Context, items []models.CRMFollowUp) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, aiConcurrency)
	for i := range items {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			items[i].Message = s.crmText(ctx, items[i])
		}(i)
	}
	wg.Wait()
}

func avgInterval(dates []time.Time) float64 {
	if len(dates) < 2 {
		return 14
	}
	var sum float64
	for i := 1; i < len(dates); i++ {
		sum += dates[i].Sub(dates[i-1]).Hours() / 24
	}
	return sum / float64(len(dates)-1)
}

func (s *CRMService) crmText(ctx context.Context, f models.CRMFollowUp) string {
	fallback := fmt.Sprintf(
		"Follow up with %s: usually reorders around day %.0f, last order %.0f days ago (%s). Payment status: %s.",
		f.ShopName, f.AvgIntervalDays, f.DaysSinceOrder, f.PredictedReorder, f.PaymentStatus,
	)
	features := fmt.Sprintf(
		"shop=%s payment=%s avg_interval=%.0f days_since=%.0f predicted=%s need_follow=%v late_risk=%v",
		f.ShopName, f.PaymentStatus, f.AvgIntervalDays, f.DaysSinceOrder, f.PredictedReorder, f.NeedsFollowUp, f.LatePayRisk,
	)
	res, err := s.ai.Infer(ctx, "CRM", features)
	if err != nil || strings.TrimSpace(res.Text) == "" {
		return fallback
	}
	return strings.TrimSpace(res.Text)
}
