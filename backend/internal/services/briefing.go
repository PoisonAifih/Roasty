package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
)

// The dashboard shows counts and leaves the owner to work out what matters.
// A briefing decides instead: every signal is converted to an estimated
// rupiah impact so unrelated problems can be ranked against each other.

type BriefingItem struct {
	Kind     string  `json:"kind"` // payment | restock | reorder
	Title    string  `json:"title"`
	Detail   string  `json:"detail"`
	Impact   float64 `json:"impact_idr"`
	Severity string  `json:"severity"`
}

type Briefing struct {
	Summary string         `json:"summary"`
	Items   []BriefingItem `json:"items"`
}

// stockoutHorizonDays bounds how far ahead lost sales are counted, so a bean
// with one day of cover is not credited with a year of missed revenue.
const stockoutHorizonDays = 30

func (ag *Agent) DailyBriefing(ctx context.Context) (Briefing, error) {
	stock, err := ag.inv.snapshot(ctx)
	if err != nil {
		return Briefing{}, err
	}
	beans, err := ag.scout.ListBeans(ctx)
	if err != nil {
		return Briefing{}, err
	}
	sellPrice := map[string]float64{}
	for _, b := range beans {
		sellPrice[b.Origin] = b.SellPricePerKg
	}

	var items []BriefingItem

	for _, s := range stock {
		if s.Urgency == "low" || s.AvgDailyKg <= 0 {
			continue
		}
		// Revenue at risk = days we would be out of stock inside the horizon,
		// times daily throughput, times sell price.
		daysShort := stockoutHorizonDays - s.DaysOfCover
		if daysShort <= 0 {
			continue
		}
		impact := daysShort * s.AvgDailyKg * sellPrice[s.Origin]
		items = append(items, BriefingItem{
			Kind:     "restock",
			Title:    fmt.Sprintf("Restock %s", s.Origin),
			Detail:   fmt.Sprintf("%.0fkg left, about %.0f days of cover at %.1fkg/day.", s.StockKg, s.DaysOfCover, s.AvgDailyKg),
			Impact:   math.Round(impact),
			Severity: s.Urgency,
		})
	}

	crm, err := ag.crm.snapshotFollowUps(ctx)
	if err != nil {
		return Briefing{}, err
	}
	for _, c := range crm {
		switch {
		case c.PaymentStatus == "overdue":
			// Overdue risk scales with how long the money has been outstanding.
			impact := c.DaysSinceOrder * 50_000
			items = append(items, BriefingItem{
				Kind:     "payment",
				Title:    fmt.Sprintf("Chase payment from %s", c.ShopName),
				Detail:   fmt.Sprintf("Overdue, last order %.0f days ago.", c.DaysSinceOrder),
				Impact:   math.Round(impact),
				Severity: "high",
			})
		case c.NeedsFollowUp:
			severity := "low"
			if c.LatePayRisk {
				severity = "med"
			}
			items = append(items, BriefingItem{
				Kind:     "reorder",
				Title:    fmt.Sprintf("Follow up %s", c.ShopName),
				Detail:   fmt.Sprintf("Reorders roughly every %.0f days; %.0f days since the last one.", c.AvgIntervalDays, c.DaysSinceOrder),
				Impact:   math.Round(c.AvgIntervalDays * 20_000),
				Severity: severity,
			})
		}
	}

	sort.Slice(items, func(i, j int) bool { return items[i].Impact > items[j].Impact })
	if len(items) > 6 {
		items = items[:6]
	}

	return Briefing{Summary: ag.briefingSummary(ctx, items), Items: items}, nil
}

func (ag *Agent) briefingSummary(ctx context.Context, items []BriefingItem) string {
	if len(items) == 0 {
		return "Nothing urgent today. Stock cover and customer follow-ups are all within normal range."
	}

	var total float64
	lines := make([]string, 0, len(items))
	for _, it := range items {
		total += it.Impact
		lines = append(lines, fmt.Sprintf("%s (%s, ~IDR %.0f)", it.Title, it.Severity, it.Impact))
	}
	fallback := fmt.Sprintf("%d actions today, about IDR %.0f at stake. Start with: %s.",
		len(items), total, items[0].Title)

	res, err := ag.ai.Infer(ctx, "BRIEFING",
		"Write one short paragraph in English summarising today's priorities: "+strings.Join(lines, "; "))
	if err != nil || strings.TrimSpace(res.Text) == "" {
		return fallback
	}
	return strings.TrimSpace(res.Text)
}
