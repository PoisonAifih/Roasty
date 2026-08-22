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
  active?: boolean
}

export type StockAdjustment = {
  id: string
  bean_id: string
  old_stock: number
  new_stock: number
  note: string
  adjusted_at: string
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
  phone: string
  avg_interval_days: number
  days_since_order: number
  predicted_reorder_date: string
  needs_follow_up: boolean
  late_pay_risk: boolean
  message: string
}

export type BasketLine = {
  bean_id: string
  origin: string
  variety: string
  qty_kg: number
  price_per_kg: number
  cost: number
  projected_profit: number
}

export type BasketPlan = {
  strategy: string
  rationale: string
  lines: BasketLine[]
  total_cost: number
  total_kg: number
  projected_profit: number
  budget_used_pct: number
}

export type BriefingItem = {
  kind: 'payment' | 'restock' | 'reorder' | string
  title: string
  detail: string
  impact_idr: number
  severity: 'high' | 'med' | 'low' | string
}

export type Briefing = {
  summary: string
  items: BriefingItem[]
}

export type TraceEvent = {
  type: 'step' | 'tool_call' | 'tool_result' | 'final' | 'error'
  tool?: string
  message: string
  step?: number
}

export type Notification = {
  id: string
  kind: 'low_stock' | 'payment_overdue' | 'reorder_due' | string
  title: string
  body: string
  severity: 'high' | 'med' | 'low' | string
  read: boolean
  created_at: string
}

export type NotificationsResponse = {
  unread: number
  notifications: Notification[]
}

export const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8014'

/** Opens the agent SSE stream. Returns a closer so callers can abort. */
export function streamAgent(
  goal: string,
  onEvent: (ev: TraceEvent) => void,
  onDone: () => void,
): () => void {
  const src = new EventSource(`${API_URL}/api/agent/stream?goal=${encodeURIComponent(goal)}`)

  src.onmessage = (e) => {
    try {
      onEvent(JSON.parse(e.data) as TraceEvent)
    } catch {
      // A malformed frame should not kill the whole run.
    }
  }
  src.addEventListener('done', () => {
    src.close()
    onDone()
  })
  src.onerror = () => {
    src.close()
    onDone()
  }

  return () => src.close()
}

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

  recordSale: (body: { bean_id: string; qty_kg: number; sold_at?: string }) =>
    request<{ status: string }>('/api/sales', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  adjustStock: (beanId: string, stockKg: number) =>
    request<{ status: string }>(`/api/beans/${encodeURIComponent(beanId)}/stock`, {
      method: 'PATCH',
      body: JSON.stringify({ stock_kg: stockKg }),
    }),
  briefing: () => request<Briefing>('/api/briefing'),
  basket: (body: { budget: number; max_kg?: number; channel?: string }) =>
    request<BasketPlan[]>('/api/scout/basket', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  markContacted: (shopId: string) =>
    request<{ status: string }>(`/api/shops/${encodeURIComponent(shopId)}/contacted`, {
      method: 'POST',
    }),
  notifications: () => request<NotificationsResponse>('/api/notifications'),
  markRead: (id: string) =>
    request<{ status: string }>(`/api/notifications/${encodeURIComponent(id)}/read`, {
      method: 'POST',
    }),
  markAllRead: () =>
    request<{ status: string }>('/api/notifications/read-all', { method: 'POST' }),
  allBeans: () => request<Bean[]>('/api/beans/all'),
  createBean: (body: Omit<Bean, 'id' | 'active'>) =>
    request<Bean>('/api/beans', { method: 'POST', body: JSON.stringify(body) }),
  updateBean: (id: string, body: Omit<Bean, 'id' | 'active' | 'stock_kg'>) =>
    request<{ status: string }>(`/api/beans/${encodeURIComponent(id)}`, {
      method: 'PATCH',
      body: JSON.stringify(body),
    }),
  disableBean: (id: string) =>
    request<{ status: string }>(`/api/beans/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  restoreBean: (id: string) =>
    request<{ status: string }>(`/api/beans/${encodeURIComponent(id)}/restore`, { method: 'POST' }),
  beanAdjustments: (id: string) =>
    request<StockAdjustment[]>(`/api/beans/${encodeURIComponent(id)}/adjustments`),
}
