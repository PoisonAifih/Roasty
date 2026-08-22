package models

import "time"

type Bean struct {
	ID                string  `json:"id"`
	Origin            string  `json:"origin"`
	Variety           string  `json:"variety"`
	Channel           string  `json:"channel"`
	PricePerKg        float64 `json:"price_per_kg"`
	StockKg           float64 `json:"stock_kg"`
	HarvestEstimateKg float64 `json:"harvest_estimate_kg"`
	Humidity          float64 `json:"humidity"`
	QualityScore      float64 `json:"quality_score"`
	SellPricePerKg    float64 `json:"sell_price_per_kg"`
	Active            bool    `json:"active"`
}

type StockAdjustment struct {
	ID         string    `json:"id"`
	BeanID     string    `json:"bean_id"`
	OldStock   float64   `json:"old_stock"`
	NewStock   float64   `json:"new_stock"`
	Note       string    `json:"note"`
	AdjustedAt time.Time `json:"adjusted_at"`
}

type ScoreBreakdown struct {
	PriceFit   float64 `json:"price_fit"`
	Quality    float64 `json:"quality"`
	SupplyRisk float64 `json:"supply_risk"`
	Margin     float64 `json:"margin"`
}

type ScoutReport struct {
	Strengths string `json:"strengths"`
	Risks     string `json:"risks"`
	Potential string `json:"potential"`
}

type ScoutRecommendation struct {
	Bean                 Bean           `json:"bean"`
	FitScore             float64        `json:"fit_score"`
	Breakdown            ScoreBreakdown `json:"breakdown"`
	Report               ScoutReport    `json:"report"`
	PredictedProfitPerKg float64        `json:"predicted_profit_per_kg"`
}

type ScoutShop struct {
	Name     string `json:"name"`
	Presence string `json:"presence"`
	URL      string `json:"url"`
	Snippet  string `json:"snippet"`
	Note     string `json:"note"`
}

type InventorySuggestion struct {
	BeanID      string  `json:"bean_id"`
	Origin      string  `json:"origin"`
	Variety     string  `json:"variety"`
	StockKg     float64 `json:"stock_kg"`
	AvgDailyKg  float64 `json:"avg_daily_kg"`
	DaysOfCover float64 `json:"days_of_cover"`
	TrendBoost  int     `json:"trend_boost"`
	Urgency     string  `json:"urgency"`
	Suggestion  string  `json:"suggestion"`
}

type CRMFollowUp struct {
	ShopID           string  `json:"shop_id"`
	ShopName         string  `json:"shop_name"`
	PaymentStatus    string  `json:"payment_status"`
	Phone            string  `json:"phone"`
	AvgIntervalDays  float64 `json:"avg_interval_days"`
	DaysSinceOrder   float64 `json:"days_since_order"`
	PredictedReorder string  `json:"predicted_reorder_date"`
	NeedsFollowUp    bool    `json:"needs_follow_up"`
	LatePayRisk      bool    `json:"late_pay_risk"`
	Message          string  `json:"message"`
}
