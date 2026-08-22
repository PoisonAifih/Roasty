package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poisonaifih/roasty/backend/internal/models"
)

const (
	wPrice   = 0.30
	wQuality = 0.25
	wSupply  = 0.20
	wMargin  = 0.25
)

type ScoutService struct {
	pool *pgxpool.Pool
	ai   *AIClient
}

func NewScoutService(pool *pgxpool.Pool, ai *AIClient) *ScoutService {
	return &ScoutService{pool: pool, ai: ai}
}

type RecommendInput struct {
	Budget   float64 `json:"budget"`
	WeightKg float64 `json:"weight_kg"`
	Channel  string  `json:"channel"`
}

func (s *ScoutService) ListBeans(ctx context.Context) ([]models.Bean, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, origin, variety, channel, price_per_kg, stock_kg,
		       harvest_estimate_kg, humidity, quality_score, sell_price_per_kg, active
		FROM beans WHERE active = true ORDER BY origin`)
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

func (s *ScoutService) Recommend(ctx context.Context, in RecommendInput) ([]models.ScoutRecommendation, error) {
	beans, err := s.ListBeans(ctx)
	if err != nil {
		return nil, err
	}
	in.Channel = normalizeChannel(in.Channel)

	var out []models.ScoutRecommendation
	for _, b := range beans {
		if in.Channel != "" && b.Channel != in.Channel {
			continue
		}
		rec := scoreBean(b, in.Budget, in.WeightKg)
		rec.Report = s.scoutReport(ctx, rec, in)
		out = append(out, rec)
	}

	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].FitScore > out[i].FitScore {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out, nil
}

// scoreOnly ranks beans without generating the per-bean AI narrative, which
// is what the agent wants: it reasons over the raw numbers itself.
func (s *ScoutService) scoreOnly(ctx context.Context, in RecommendInput) ([]models.ScoutRecommendation, error) {
	beans, err := s.ListBeans(ctx)
	if err != nil {
		return nil, err
	}
	in.Channel = normalizeChannel(in.Channel)

	var out []models.ScoutRecommendation
	for _, b := range beans {
		if in.Channel != "" && b.Channel != in.Channel {
			continue
		}
		out = append(out, scoreBean(b, in.Budget, in.WeightKg))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].FitScore > out[j].FitScore })
	return out, nil
}

func scoreBean(b models.Bean, budget, weightKg float64) models.ScoutRecommendation {
	priceFit := priceFitScore(b, budget, weightKg)

	quality := clamp(b.QualityScore, 0, 100)

	harvestScore := clamp(30+b.HarvestEstimateKg/40, 0, 100)
	humidityScore := clamp(100-math.Abs(b.Humidity-65)*2, 0, 100)
	supply := clamp(0.65*harvestScore+0.35*humidityScore, 0, 100)

	if weightKg > 0 {
		cover := (b.StockKg + b.HarvestEstimateKg*0.08) / weightKg
		supply = clamp(0.55*supply+0.45*clamp(cover*100, 0, 100), 0, 100)
	}

	marginPerKg := b.SellPricePerKg - b.PricePerKg
	margin := clamp(marginPerKg/900, 0, 100)

	fit := wPrice*priceFit + wQuality*quality + wSupply*supply + wMargin*margin

	return models.ScoutRecommendation{
		Bean:     b,
		FitScore: math.Round(fit*10) / 10,
		Breakdown: models.ScoreBreakdown{
			PriceFit:   math.Round(priceFit*10) / 10,
			Quality:    math.Round(quality*10) / 10,
			SupplyRisk: math.Round(supply*10) / 10,
			Margin:     math.Round(margin*10) / 10,
		},
		PredictedProfitPerKg: math.Round(marginPerKg*100) / 100,
	}
}

func priceFitScore(b models.Bean, budget, weightKg float64) float64 {
	hasBudget := budget > 0
	hasWeight := weightKg > 0

	switch {
	case hasBudget && hasWeight:
		cost := weightKg * b.PricePerKg
		budgetCover := clamp(budget/cost*100, 0, 100)
		stockCover := clamp((b.StockKg+b.HarvestEstimateKg*0.08)/weightKg*100, 0, 100)
		return clamp(0.65*budgetCover+0.35*stockCover, 0, 100)
	case hasBudget:
		affordableKg := budget / b.PricePerKg
		return clamp(affordableKg/60*100, 0, 100)
	case hasWeight:
		stockCover := clamp((b.StockKg+b.HarvestEstimateKg*0.08)/weightKg*100, 0, 100)
		costPressure := clamp(100-(weightKg*b.PricePerKg)/80_000, 0, 100)
		return clamp(0.7*stockCover+0.3*costPressure, 0, 100)
	default:
		return 72
	}
}

func (s *ScoutService) scoutReport(ctx context.Context, rec models.ScoutRecommendation, in RecommendInput) models.ScoutReport {
	features := fmt.Sprintf(
		"Write the report in English. origin=%s variety=%s channel=%s price=%.0f quality=%.0f supply=%.0f margin=%.0f fit=%.1f profit=%.0f budget=%.0f weight_kg=%.1f",
		rec.Bean.Origin, rec.Bean.Variety, normalizeChannel(rec.Bean.Channel),
		rec.Bean.PricePerKg, rec.Breakdown.Quality, rec.Breakdown.SupplyRisk,
		rec.Breakdown.Margin, rec.FitScore, rec.PredictedProfitPerKg, in.Budget, in.WeightKg,
	)

	fallback := models.ScoutReport{
		Strengths: fmt.Sprintf("Quality %.0f and margin ~IDR %.0f/kg for %s.", rec.Breakdown.Quality, rec.PredictedProfitPerKg, rec.Bean.Origin),
		Risks:     fmt.Sprintf("Supply score %.0f (harvest/humidity); %s channel needs contract checks.", rec.Breakdown.SupplyRisk, rec.Bean.Channel),
		Potential: fmt.Sprintf("Fit score %.1f — good roast-and-sell candidate if it fits your constraints.", rec.FitScore),
	}

	res, err := s.ai.Infer(ctx, "SCOUT", features)
	if err != nil {
		return fallback
	}
	out := fallback
	if res.Strengths != "" {
		out.Strengths = res.Strengths
	}
	if res.Risks != "" {
		out.Risks = res.Risks
	}
	if res.Potential != "" {
		out.Potential = res.Potential
	}
	if res.Strengths == "" && res.Risks == "" && res.Potential == "" && res.Text != "" {
		parts := strings.Split(res.Text, "|")
		if len(parts) >= 1 {
			out.Strengths = strings.TrimSpace(parts[0])
		}
		if len(parts) >= 2 {
			out.Risks = strings.TrimSpace(parts[1])
		}
		if len(parts) >= 3 {
			out.Potential = strings.TrimSpace(parts[2])
		}
	}
	return out
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func normalizeChannel(channel string) string {
	switch strings.ToLower(strings.TrimSpace(channel)) {
	case "petani", "farmer":
		return "farmer"
	case "tengkulak", "agen", "middle man", "middleman":
		return "middleman"
	default:
		return strings.ToLower(strings.TrimSpace(channel))
	}
}
