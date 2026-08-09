export type Bean = {
  id: string
  origin: string
  variety: string
  channel: string
  price_per_kg: number
  stock_kg: number
  harvest_estimate_kg: number
  humidity: number
  quality_score: number
  sell_price_per_kg: number
}

export type ScoutRecommendation = {
  bean: Bean
  fit_score: number
  breakdown: {
    price_fit: number
    quality: number
    supply_risk: number
    margin: number
  }
  report: {
    strengths: string
    risks: string
    potential: string
  }
  predicted_profit_per_kg: number
}

export type ScoutShop = {
  name: string
  presence: 'online' | 'offline' | string
  url: string
  snippet: string
  note: string
}

export type InventorySuggestion = {
  bean_id: string
  origin: string
  variety: string
  stock_kg: number
  avg_daily_kg: number
  days_of_cover: number
  trend_boost: number
  urgency: 'high' | 'med' | 'low'
  suggestion: string
}

export type CRMFollowUp = {
  shop_id: string
  shop_name: string
  payment_status: string
  avg_interval_days: number
  days_since_order: number
  predicted_reorder_date: string
  needs_follow_up: boolean
  late_pay_risk: boolean
  message: string
}

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8014'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    headers: { 'Content-Type': 'application/json', ...(init?.headers || {}) },
    ...init,
  })
  if (!res.ok) {
    const text = await res.text()
    try {
      const parsed = JSON.parse(text) as { error?: string }
      if (parsed.error) throw new Error(parsed.error)
    } catch (e) {
      if (!(e instanceof SyntaxError)) throw e
    }
    throw new Error(text.trim() || res.statusText)
  }
  return res.json() as Promise<T>
}

export const api = {
  beans: () => request<Bean[]>('/api/beans'),
  recommend: (opts: { budget?: number; weight_kg?: number; channel?: string }) =>
    request<ScoutRecommendation[]>('/api/scout/recommend', {
      method: 'POST',
      body: JSON.stringify({
        budget: opts.budget ?? 0,
        weight_kg: opts.weight_kg ?? 0,
        channel: opts.channel || '',
      }),
    }),
  scoutShops: (origin: string, variety: string) =>
    request<ScoutShop[]>(
      `/api/scout/shops?origin=${encodeURIComponent(origin)}&variety=${encodeURIComponent(variety)}`,
    ),
  inventory: () => request<InventorySuggestion[]>('/api/inventory/suggestions'),
  crm: () => request<CRMFollowUp[]>('/api/crm/follow-ups'),
}
