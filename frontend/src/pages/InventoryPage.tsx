import { useEffect, useState, type FormEvent } from 'react'
import { api, type InventorySuggestion, type StockAdjustment } from '@/api/client'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'

export function InventoryPage() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [items, setItems] = useState<InventorySuggestion[]>([])
  const [selected, setSelected] = useState<InventorySuggestion | null>(null)

  async function load() {
    setLoading(true)
    setError('')
    try {
      const data = await api.inventory()
      setItems(data)
      setSelected((prev) => data.find((d) => d.bean_id === prev?.bean_id) ?? data[0] ?? null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load stock')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  const highCount = items.filter((i) => i.urgency === 'high').length

  return (
    <section className="flex flex-col gap-5">
      <header className="flex flex-wrap items-end justify-between gap-4 border-b border-border pb-4">
        <div>
          <h1 className="font-heading m-0 text-[clamp(1.9rem,3.2vw,2.45rem)] font-bold tracking-wide">
            Smart Inventory
          </h1>
          <p className="mt-2 max-w-xl text-[1.02rem] text-muted-foreground">
            Track inventory and restock recommendations.
          </p>
        </div>
        <Button
          variant="outline"
          type="button"
          onClick={() => void load()}
          disabled={loading}
          className="rounded-none"
        >
          {loading ? 'Loading…' : 'Refresh'}
        </Button>
      </header>

      {items.length > 0 && (
        <div className="flex flex-wrap gap-2.5">
          <Badge variant="secondary" className="rounded-none">
            {highCount} high urgency
          </Badge>
          <Badge variant="outline" className="rounded-none">
            {items.length} origins tracked
          </Badge>
        </div>
      )}

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      {loading && items.length === 0 && (
        <p className="border-y border-dashed border-border py-5 italic text-muted-foreground">
          Loading stock data…
        </p>
      )}
      {!loading && items.length === 0 && !error && (
        <p className="border-y border-dashed border-border py-5 italic text-muted-foreground">
          No stock data yet.
        </p>
      )}

      {items.length > 0 && (
        <div className="grid grid-cols-1 items-start gap-4 md:grid-cols-[0.95fr_1.05fr]">
          <Card className="rounded-none shadow-none">
            <CardHeader className="pb-2">
              <CardTitle className="font-heading text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
                Stock
              </CardTitle>
              <Separator />
            </CardHeader>
            <CardContent>
              <ul className="m-0 flex list-none flex-col p-0">
                {items.map((i) => (
                  <li key={i.bean_id} className="border-t border-border first:border-t-0">
                    <button
                      type="button"
                      className={cn(
                        'flex w-full items-center gap-3 bg-transparent px-1.5 py-3.5 text-left hover:bg-accent',
                        selected?.bean_id === i.bean_id &&
                          'bg-accent shadow-[inset_2px_0_0_var(--destructive)]',
                      )}
                      onClick={() => setSelected(i)}
                    >
                      <span className="h-2.5 w-2.5 shrink-0 rounded-full bg-destructive" />
                      <span className="min-w-0 flex-1">
                        <strong className="font-heading block font-semibold">{i.origin}</strong>
                        <small className="block text-sm text-muted-foreground">
                          {i.stock_kg} kg ·{' '}
                          {i.days_of_cover >= 999 ? 'no sales data' : `${i.days_of_cover}d cover`}
                        </small>
                      </span>
                      <Badge
                        variant={
                          i.urgency === 'high'
                            ? 'destructive'
                            : i.urgency === 'med'
                              ? 'secondary'
                              : 'outline'
                        }
                      >
                        {i.urgency}
                      </Badge>
                    </button>
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>

          {selected && <DetailPanel key={selected.bean_id} item={selected} onMutate={load} />}
        </div>
      )}
    </section>
  )
}

function DetailPanel({
  item,
  onMutate,
}: {
  item: InventorySuggestion
  onMutate: () => Promise<void>
}) {
  // Adjust stock form
  const [adjStock, setAdjStock] = useState(String(item.stock_kg))
  const [adjBusy, setAdjBusy] = useState(false)
  const [adjMsg, setAdjMsg] = useState<{ ok: boolean; text: string } | null>(null)

  // Record sale form
  const [saleQty, setSaleQty] = useState('')
  const [saleDate, setSaleDate] = useState('')
  const [saleBusy, setSaleBusy] = useState(false)
  const [saleMsg, setSaleMsg] = useState<{ ok: boolean; text: string } | null>(null)

  // Adjustment history
  const [adjustments, setAdjustments] = useState<StockAdjustment[]>([])

  useEffect(() => {
    api
      .beanAdjustments(item.bean_id)
      .then(setAdjustments)
      .catch(() => {})
  }, [item.bean_id])

  async function onAdjust(e: FormEvent) {
    e.preventDefault()
    const kg = parseFloat(adjStock)
    if (isNaN(kg) || kg < 0) {
      setAdjMsg({ ok: false, text: 'Enter a valid non-negative number.' })
      return
    }
    setAdjBusy(true)
    setAdjMsg(null)
    try {
      await api.adjustStock(item.bean_id, kg)
      setAdjMsg({ ok: true, text: `Stock set to ${kg} kg.` })
      await onMutate()
    } catch (err) {
      setAdjMsg({ ok: false, text: err instanceof Error ? err.message : 'Failed to adjust stock.' })
    } finally {
      setAdjBusy(false)
    }
  }

  async function onSale(e: FormEvent) {
    e.preventDefault()
    const qty = parseFloat(saleQty)
    if (isNaN(qty) || qty <= 0) {
      setSaleMsg({ ok: false, text: 'Enter a valid quantity greater than zero.' })
      return
    }
    setSaleBusy(true)
    setSaleMsg(null)
    try {
      await api.recordSale({ bean_id: item.bean_id, qty_kg: qty, sold_at: saleDate || undefined })
      setSaleMsg({ ok: true, text: `Recorded sale of ${qty} kg.` })
      setSaleQty('')
      setSaleDate('')
      await onMutate()
    } catch (err) {
      setSaleMsg({ ok: false, text: err instanceof Error ? err.message : 'Failed to record sale.' })
    } finally {
      setSaleBusy(false)
    }
  }

  return (
    <Card className="rounded-none shadow-none">
      <CardContent className="pt-6">
        <h2 className="font-heading m-0 text-[1.45rem] font-semibold">{item.origin}</h2>
        <p className="mt-1 text-muted-foreground">{item.variety}</p>

        <div className="my-4 grid grid-cols-2 border border-border">
          {(
            [
              ['Stock', `${item.stock_kg} kg`],
              [
                'Days of cover',
                item.days_of_cover >= 999 ? 'No sales data' : String(item.days_of_cover),
              ],
              ['Daily average', `${item.avg_daily_kg.toFixed(2)} kg`],
              ['Trend signal', item.trend_boost > 0 ? `+${item.trend_boost}` : '—'],
            ] as const
          ).map(([label, value], idx) => (
            <div
              key={label}
              className={cn(
                'p-3',
                idx % 2 === 0 && 'border-r border-border',
                idx < 2 && 'border-b border-border',
              )}
            >
              <span className="block text-xs uppercase tracking-wider text-muted-foreground">
                {label}
              </span>
              <strong className="font-heading text-[1.05rem] font-semibold">{value}</strong>
            </div>
          ))}
        </div>

        <div className="border-t border-border pt-4">
          <h3 className="font-heading mb-1.5 text-[1.05rem] font-bold text-destructive">
            Restock suggestion
          </h3>
          <p className="m-0 text-sm">{item.suggestion}</p>
        </div>

        {/* ── Adjust stock ───────────────────────────────────────────── */}
        <form onSubmit={(e) => void onAdjust(e)} className="mt-5 border-t border-border pt-4">
          <h3 className="font-heading mb-3 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
            Adjust stock (restock / opname)
          </h3>
          <div className="flex items-end gap-2">
            <div className="flex flex-1 flex-col gap-1.5">
              <Label htmlFor={`adj-${item.bean_id}`}>New total stock (kg)</Label>
              <Input
                id={`adj-${item.bean_id}`}
                type="number"
                min={0}
                step={0.1}
                value={adjStock}
                onChange={(e) => setAdjStock(e.target.value)}
                className="rounded-none"
              />
            </div>
            <Button type="submit" disabled={adjBusy} className="rounded-none">
              {adjBusy ? 'Saving…' : 'Set stock'}
            </Button>
          </div>
          {adjMsg && (
            <p
              className={cn(
                'mt-2 text-sm',
                adjMsg.ok ? 'text-green-600 dark:text-green-400' : 'text-destructive',
              )}
            >
              {adjMsg.text}
            </p>
          )}
        </form>

        {/* ── Record sale ────────────────────────────────────────────── */}
        <form onSubmit={(e) => void onSale(e)} className="mt-5 border-t border-border pt-4">
          <h3 className="font-heading mb-3 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
            Record sale
          </h3>
          <div className="grid grid-cols-[1fr_1fr_auto] items-end gap-2">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor={`sale-qty-${item.bean_id}`}>Qty (kg)</Label>
              <Input
                id={`sale-qty-${item.bean_id}`}
                type="number"
                min={0.1}
                step={0.1}
                placeholder="e.g. 5"
                value={saleQty}
                onChange={(e) => setSaleQty(e.target.value)}
                className="rounded-none"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor={`sale-date-${item.bean_id}`}>Date (optional)</Label>
              <Input
                id={`sale-date-${item.bean_id}`}
                type="date"
                value={saleDate}
                onChange={(e) => setSaleDate(e.target.value)}
                className="rounded-none"
              />
            </div>
            <Button type="submit" disabled={saleBusy} className="rounded-none">
              {saleBusy ? 'Saving…' : 'Record'}
            </Button>
          </div>
          {saleMsg && (
            <p
              className={cn(
                'mt-2 text-sm',
                saleMsg.ok ? 'text-green-600 dark:text-green-400' : 'text-destructive',
              )}
            >
              {saleMsg.text}
            </p>
          )}
        </form>

        {/* ── Adjustment history ─────────────────────────────────────── */}
        <div className="mt-5 border-t border-border pt-4">
          <h3 className="font-heading mb-2 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
            Adjustment history
          </h3>
          {adjustments.length === 0 ? (
            <p className="text-sm text-muted-foreground">No adjustments yet.</p>
          ) : (
            <ul className="m-0 list-none p-0">
              {adjustments.slice(0, 5).map((a) => (
                <li
                  key={a.id}
                  className="flex items-baseline gap-3 border-t border-border py-2 text-sm first:border-t-0"
                >
                  <span className="w-[6.5rem] shrink-0 text-xs text-muted-foreground">
                    {new Date(a.adjusted_at).toLocaleDateString()}
                  </span>
                  <span className="font-mono">
                    {a.old_stock} → {a.new_stock} kg
                  </span>
                  {a.note && <span className="text-muted-foreground">{a.note}</span>}
                </li>
              ))}
            </ul>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
