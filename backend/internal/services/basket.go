package services

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/poisonaifih/roasty/backend/internal/models"
)

// A roastery does not buy one bean, it buys a mix under a budget ceiling.
// Ranking beans individually (what scoreOnly does) never answers "what do I
// actually purchase with 50 million rupiah", so this allocates across several
// origins and returns a few strategies to choose between.

type BasketLine struct {
	BeanID     string  `json:"bean_id"`
	Origin     string  `json:"origin"`
	Variety    string  `json:"variety"`
	QtyKg      float64 `json:"qty_kg"`
	PricePerKg float64 `json:"price_per_kg"`
	Cost       float64 `json:"cost"`
	Profit     float64 `json:"projected_profit"`
}

type BasketPlan struct {
	Strategy        string       `json:"strategy"`
	Rationale       string       `json:"rationale"`
	Lines           []BasketLine `json:"lines"`
	TotalCost       float64      `json:"total_cost"`
	TotalKg         float64      `json:"total_kg"`
	ProjectedProfit float64      `json:"projected_profit"`
	BudgetUsedPct   float64      `json:"budget_used_pct"`
}

type BasketInput struct {
	Budget  float64 `json:"budget"`
	MaxKg   float64 `json:"max_kg"`
	Channel string  `json:"channel"`
}

// BuildBaskets returns three allocations of the same budget so the user can
// weigh profit against risk instead of being handed a single answer.
func (s *ScoutService) BuildBaskets(ctx context.Context, in BasketInput) ([]BasketPlan, error) {
	if in.Budget <= 0 {
		return nil, fmt.Errorf("budget must be greater than zero")
	}

	beans, err := s.ListBeans(ctx)
	if err != nil {
		return nil, err
	}
	channel := normalizeChannel(in.Channel)

	var pool []models.Bean
	for _, b := range beans {
		if channel != "" && b.Channel != channel {
			continue
		}
		if b.PricePerKg > 0 {
			pool = append(pool, b)
		}
	}
	if len(pool) == 0 {
		return nil, fmt.Errorf("no beans match channel %q", in.Channel)
	}

	plans := []BasketPlan{
		allocate(pool, in, "max_profit",
			"Concentrates the budget on the two highest-margin origins. Best return per rupiah, most exposed if one harvest disappoints.",
			profile{maxLines: 2, sharePerPick: 0.75},
			func(b models.Bean) float64 {
				return (b.SellPricePerKg - b.PricePerKg) / b.PricePerKg
			}),
		allocate(pool, in, "lowest_risk",
			"Spreads across five origins ranked by quality, harvest cover and ideal humidity. Thinner margin bought in exchange for supply certainty.",
			profile{maxLines: 5, sharePerPick: 0.3},
			func(b models.Bean) float64 {
				harvest := clamp(30+b.HarvestEstimateKg/40, 0, 100)
				humidity := clamp(100-math.Abs(b.Humidity-65)*2, 0, 100)
				return (b.QualityScore*0.5 + harvest*0.25 + humidity*0.25) / 100
			}),
		allocate(pool, in, "balanced",
			"Blends margin and quality over three or four origins - the default choice when nothing about the season is unusual.",
			profile{maxLines: 4, sharePerPick: 0.45},
			func(b models.Bean) float64 {
				margin := (b.SellPricePerKg - b.PricePerKg) / b.PricePerKg
				quality := b.QualityScore / 100
				return margin*0.55 + quality*0.45
			}),
	}
	return plans, nil
}

// profile controls how concentrated a strategy is. Without differing spreads
// the three plans collapse into the same allocation whenever availability,
// rather than budget, is the binding constraint.
type profile struct {
	maxLines     int
	sharePerPick float64
}

// allocate walks beans best-first by the supplied score and spends the budget
// in decreasing slices, so the strongest candidate takes the largest share
// while later ones still get funded. Availability caps every line.
func allocate(pool []models.Bean, in BasketInput, name, rationale string, p profile, score func(models.Bean) float64) BasketPlan {
	ranked := make([]models.Bean, len(pool))
	copy(ranked, pool)
	sort.Slice(ranked, func(i, j int) bool { return score(ranked[i]) > score(ranked[j]) })

	// Diminishing shares keep the top pick dominant without letting it
	// swallow the whole budget.
	sharePerPick := p.sharePerPick
	maxLines := p.maxLines

	plan := BasketPlan{Strategy: name, Rationale: rationale}
	remaining := in.Budget
	remainingKg := in.MaxKg

	for i, b := range ranked {
		if i >= maxLines || remaining <= 0 {
			break
		}
		if in.MaxKg > 0 && remainingKg <= 0 {
			break
		}

		spend := remaining * sharePerPick
		if i == len(ranked)-1 || i == maxLines-1 {
			spend = remaining // last pick mops up the remainder
		}

		qty := spend / b.PricePerKg

		// Never plan to buy more than the origin can actually supply.
		available := b.StockKg + b.HarvestEstimateKg*0.08
		if qty > available {
			qty = available
		}
		if in.MaxKg > 0 && qty > remainingKg {
			qty = remainingKg
		}
		qty = math.Floor(qty)
		if qty < 1 {
			continue
		}

		cost := qty * b.PricePerKg
		profit := qty * (b.SellPricePerKg - b.PricePerKg)

		plan.Lines = append(plan.Lines, BasketLine{
			BeanID:     b.ID,
			Origin:     b.Origin,
			Variety:    b.Variety,
			QtyKg:      qty,
			PricePerKg: b.PricePerKg,
			Cost:       math.Round(cost),
			Profit:     math.Round(profit),
		})
		plan.TotalCost += cost
		plan.TotalKg += qty
		plan.ProjectedProfit += profit
		remaining -= cost
		remainingKg -= qty
	}

	plan.TotalCost = math.Round(plan.TotalCost)
	plan.ProjectedProfit = math.Round(plan.ProjectedProfit)
	if in.Budget > 0 {
		plan.BudgetUsedPct = math.Round(plan.TotalCost / in.Budget * 1000) / 10
	}
	return plan
}
