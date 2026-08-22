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

// Suggestions returns restock signals with an AI narrative per row.
func (s *InventoryService) Suggestions(ctx context.Context) ([]models.InventorySuggestion, error) {
	out, err := s.snapshot(ctx)
	if err != nil {
		return nil, err
	}
	s.fillRestockTexts(ctx, out)
	return out, nil
}

// snapshot computes the same signals without calling the AI. The agent uses
// this: it reasons over the numbers itself, so per-row prose would be wasted
// tokens and latency.
func (s *InventoryService) snapshot(ctx context.Context) ([]models.InventorySuggestion, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT b.id::text, b.origin, b.variety, b.stock_kg,
		       COALESCE(SUM(s.qty_kg), 0) AS total_sold,
		       COALESCE(MAX(t.strength), 0) AS trend
		FROM beans b
		LEFT JOIN sales s ON s.bean_id = b.id AND s.sold_at >= CURRENT_DATE - 28
		LEFT JOIN trend_signals t ON t.origin = b.origin
		WHERE b.active = true
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

// ListAllBeans returns all beans including inactive, for the management page.
func (s *InventoryService) ListAllBeans(ctx context.Context) ([]models.Bean, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, origin, variety, channel, price_per_kg, stock_kg,
		       harvest_estimate_kg, humidity, quality_score, sell_price_per_kg, active
		FROM beans ORDER BY active DESC, origin`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var beans []models.Bean
	for rows.Next() {
		var b models.Bean
		if err := rows.Scan(&b.ID, &b.Origin, &b.Variety, &b.Channel, &b.PricePerKg, &b.StockKg,
			&b.HarvestEstimateKg, &b.Humidity, &b.QualityScore, &b.SellPricePerKg, &b.Active); err != nil {
			return nil, err
		}
		b.Channel = normalizeChannel(b.Channel)
		beans = append(beans, b)
	}
	return beans, rows.Err()
}

// RecentAdjustments returns the last N stock adjustments for a bean.
func (s *InventoryService) RecentAdjustments(ctx context.Context, beanID string, limit int) ([]models.StockAdjustment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, bean_id::text, old_stock, new_stock, COALESCE(note, ''), adjusted_at
		FROM stock_adjustments
		WHERE bean_id = $1::uuid
		ORDER BY adjusted_at DESC
		LIMIT $2`, beanID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.StockAdjustment
	for rows.Next() {
		var a models.StockAdjustment
		if err := rows.Scan(&a.ID, &a.BeanID, &a.OldStock, &a.NewStock, &a.Note, &a.AdjustedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

type CreateBeanInput struct {
	Origin            string  `json:"origin"`
	Variety           string  `json:"variety"`
	Channel           string  `json:"channel"`
	PricePerKg        float64 `json:"price_per_kg"`
	SellPricePerKg    float64 `json:"sell_price_per_kg"`
	StockKg           float64 `json:"stock_kg"`
	HarvestEstimateKg float64 `json:"harvest_estimate_kg"`
	Humidity          float64 `json:"humidity"`
	QualityScore      float64 `json:"quality_score"`
}

func (s *InventoryService) CreateBean(ctx context.Context, in CreateBeanInput) (*models.Bean, error) {
	if strings.TrimSpace(in.Origin) == "" {
		return nil, fmt.Errorf("origin required")
	}
	if strings.TrimSpace(in.Variety) == "" {
		return nil, fmt.Errorf("variety required")
	}
	if in.PricePerKg <= 0 {
		return nil, fmt.Errorf("price_per_kg must be > 0")
	}
	if in.SellPricePerKg <= 0 {
		return nil, fmt.Errorf("sell_price_per_kg must be > 0")
	}
	channel := normalizeChannel(in.Channel)
	var b models.Bean
	err := s.pool.QueryRow(ctx, `
		INSERT INTO beans (origin, variety, channel, price_per_kg, sell_price_per_kg, stock_kg,
		                   harvest_estimate_kg, humidity, quality_score)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id::text, origin, variety, channel, price_per_kg, stock_kg,
		          harvest_estimate_kg, humidity, quality_score, sell_price_per_kg, active`,
		in.Origin, in.Variety, channel, in.PricePerKg, in.SellPricePerKg, in.StockKg,
		in.HarvestEstimateKg, in.Humidity, in.QualityScore,
	).Scan(&b.ID, &b.Origin, &b.Variety, &b.Channel, &b.PricePerKg, &b.StockKg,
		&b.HarvestEstimateKg, &b.Humidity, &b.QualityScore, &b.SellPricePerKg, &b.Active)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

type UpdateBeanInput struct {
	Origin            string  `json:"origin"`
	Variety           string  `json:"variety"`
	Channel           string  `json:"channel"`
	PricePerKg        float64 `json:"price_per_kg"`
	SellPricePerKg    float64 `json:"sell_price_per_kg"`
	HarvestEstimateKg float64 `json:"harvest_estimate_kg"`
	Humidity          float64 `json:"humidity"`
	QualityScore      float64 `json:"quality_score"`
}

func (s *InventoryService) UpdateBean(ctx context.Context, beanID string, in UpdateBeanInput) error {
	if beanID == "" {
		return fmt.Errorf("bean id required")
	}
	if strings.TrimSpace(in.Origin) == "" {
		return fmt.Errorf("origin required")
	}
	if in.PricePerKg <= 0 {
		return fmt.Errorf("price_per_kg must be > 0")
	}
	if in.SellPricePerKg <= 0 {
		return fmt.Errorf("sell_price_per_kg must be > 0")
	}
	channel := normalizeChannel(in.Channel)
	tag, err := s.pool.Exec(ctx, `
		UPDATE beans
		SET origin=$1, variety=$2, channel=$3, price_per_kg=$4, sell_price_per_kg=$5,
		    harvest_estimate_kg=$6, humidity=$7, quality_score=$8
		WHERE id=$9`,
		in.Origin, in.Variety, channel, in.PricePerKg, in.SellPricePerKg,
		in.HarvestEstimateKg, in.Humidity, in.QualityScore, beanID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("bean %s not found", beanID)
	}
	return nil
}

// SetBeanActive enables or disables a bean (soft delete / restore).
func (s *InventoryService) SetBeanActive(ctx context.Context, beanID string, active bool) error {
	if beanID == "" {
		return fmt.Errorf("bean id required")
	}
	tag, err := s.pool.Exec(ctx, `UPDATE beans SET active=$1 WHERE id=$2`, active, beanID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("bean %s not found", beanID)
	}
	return nil
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
