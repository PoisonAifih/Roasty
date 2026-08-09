package services

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poisonaifih/roasty/backend/internal/models"
)

// aiConcurrency bounds simultaneous OpenRouter calls. Row narratives are
// independent, so they fan out instead of running one at a time.
const aiConcurrency = 6

type InventoryService struct {
	pool *pgxpool.Pool
	ai   *AIClient
}

func NewInventoryService(pool *pgxpool.Pool, ai *AIClient) *InventoryService {
	return &InventoryService{pool: pool, ai: ai}
}

func (s *InventoryService) Suggestions(ctx context.Context) ([]models.InventorySuggestion, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT b.id::text, b.origin, b.variety, b.stock_kg,
		       COALESCE(SUM(s.qty_kg), 0) AS total_sold,
		       COALESCE(MAX(t.strength), 0) AS trend
		FROM beans b
		LEFT JOIN sales s ON s.bean_id = b.id AND s.sold_at >= CURRENT_DATE - 28
		LEFT JOIN trend_signals t ON t.origin = b.origin
		GROUP BY b.id, b.origin, b.variety, b.stock_kg
		ORDER BY b.origin`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.InventorySuggestion
	for rows.Next() {
		var sug models.InventorySuggestion
		var totalSold float64
		if err := rows.Scan(&sug.BeanID, &sug.Origin, &sug.Variety, &sug.StockKg, &totalSold, &sug.TrendBoost); err != nil {
			return nil, err
		}
		sug.AvgDailyKg = totalSold / 28
		if sug.AvgDailyKg <= 0 {
			sug.DaysOfCover = 999
		} else {
			sug.DaysOfCover = math.Round((sug.StockKg/sug.AvgDailyKg)*10) / 10
		}
		sug.Urgency = urgency(sug.DaysOfCover, sug.TrendBoost)
		out = append(out, sug)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	s.fillRestockTexts(ctx, out)

	rank := map[string]int{"high": 0, "med": 1, "low": 2}
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if rank[out[j].Urgency] < rank[out[i].Urgency] ||
				(rank[out[j].Urgency] == rank[out[i].Urgency] && out[j].DaysOfCover < out[i].DaysOfCover) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out, nil
}

func urgency(days float64, trend int) string {
	score := days
	if trend >= 4 {
		score -= 5
	} else if trend >= 3 {
		score -= 3
	}
	if score < 14 {
		return "high"
	}
	if score < 28 {
		return "med"
	}
	return "low"
}

// fillRestockTexts populates Suggestion for every item concurrently, capped at
// aiConcurrency in-flight requests. Each goroutine writes to its own index, so
// no further synchronisation is needed.
func (s *InventoryService) fillRestockTexts(ctx context.Context, items []models.InventorySuggestion) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, aiConcurrency)
	for i := range items {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			items[i].Suggestion = s.restockText(ctx, items[i])
		}(i)
	}
	wg.Wait()
}

func (s *InventoryService) restockText(ctx context.Context, sug models.InventorySuggestion) string {
	fallback := fmt.Sprintf(
		"Restock %s now: stock %.0fkg, ~%.0f days cover, trend %d, urgency %s.",
		sug.Origin, sug.StockKg, sug.DaysOfCover, sug.TrendBoost, sug.Urgency,
	)
	features := fmt.Sprintf(
		"origin=%s stock=%.0f avg_daily=%.2f days_cover=%.1f trend=%d urgency=%s",
		sug.Origin, sug.StockKg, sug.AvgDailyKg, sug.DaysOfCover, sug.TrendBoost, sug.Urgency,
	)
	res, err := s.ai.Infer(ctx, "RESTOCK", features)
	if err != nil || strings.TrimSpace(res.Text) == "" {
		return fallback
	}
	return strings.TrimSpace(res.Text)
}
