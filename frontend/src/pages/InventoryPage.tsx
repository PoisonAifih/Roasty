import { useEffect, useState } from 'react'
import { api, type InventorySuggestion } from '@/api/client'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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
                          {i.stock_kg}kg · {i.days_of_cover} days cover
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

          {selected && (
            <Card className="rounded-none shadow-none">
              <CardContent className="pt-6">
                <h2 className="font-heading m-0 text-[1.45rem] font-semibold">{selected.origin}</h2>
                <p className="mt-1 text-muted-foreground">{selected.variety}</p>
                <div className="my-4 grid grid-cols-2 border border-border">
                  {[
                    ['Stock', `${selected.stock_kg} kg`],
                    ['Days of cover', String(selected.days_of_cover)],
                    ['Daily average', `${selected.avg_daily_kg.toFixed(2)} kg`],
                    ['Trend', String(selected.trend_boost)],
                  ].map(([label, value], idx) => (
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
                <div className="mt-1.5 border-t border-border pt-4">
                  <h3 className="font-heading mb-1.5 text-[1.05rem] font-bold text-destructive">
                    Restock suggestion
                  </h3>
                  <p className="m-0">{selected.suggestion}</p>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      )}
    </section>
  )
}
