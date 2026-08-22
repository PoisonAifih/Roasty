import { useEffect, useState, type FormEvent } from 'react'
import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from 'recharts'
import { api, type ScoutRecommendation, type ScoutShop } from '@/api/client'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from '@/components/ui/chart'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'

const fitChartConfig = {
  value: {
    label: 'Score',
    color: 'var(--chart-1)',
  },
} satisfies ChartConfig

function labelChannel(channel: string) {
  const c = channel.trim().toLowerCase()
  if (c === 'petani' || c === 'farmer') return 'Farmer'
  if (c === 'tengkulak' || c === 'agen' || c === 'middleman' || c === 'middle man') return 'Middleman'
  if (!c) return '—'
  return channel.charAt(0).toUpperCase() + channel.slice(1)
}

export function ScoutPage() {
  const [budget, setBudget] = useState('')
  const [weightKg, setWeightKg] = useState('')
  const [channel, setChannel] = useState('all')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [recs, setRecs] = useState<ScoutRecommendation[]>([])
  const [selected, setSelected] = useState<ScoutRecommendation | null>(null)
  const [shops, setShops] = useState<ScoutShop[]>([])
  const [shopsLoading, setShopsLoading] = useState(false)
  const [shopsError, setShopsError] = useState('')

  useEffect(() => {
    if (!selected) {
      setShops([])
      setShopsError('')
      return
    }
    let cancelled = false
    setShopsLoading(true)
    setShopsError('')
    api
      .scoutShops(selected.bean.origin, selected.bean.variety)
      .then((data) => {
        if (!cancelled) setShops(data)
      })
      .catch((err) => {
        if (!cancelled) {
          setShops([])
          setShopsError(err instanceof Error ? err.message : 'Failed to search shops')
        }
      })
      .finally(() => {
        if (!cancelled) setShopsLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [selected?.bean.id, selected?.bean.origin, selected?.bean.variety])

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const data = await api.recommend({
        budget: budget === '' ? 0 : Number(budget),
        weight_kg: weightKg === '' ? 0 : Number(weightKg),
        channel: channel === 'all' ? '' : channel,
      })
      setRecs(data)
      setSelected(data[0] ?? null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to call API')
    } finally {
      setLoading(false)
    }
  }

  const chartData = selected
    ? [
        { name: 'Price', value: selected.breakdown.price_fit },
        { name: 'Quality', value: selected.breakdown.quality },
        { name: 'Supply', value: selected.breakdown.supply_risk },
        { name: 'Margin', value: selected.breakdown.margin },
      ]
    : []

  const onlineShops = shops.filter((s) => s.presence === 'online')
  const offlineShops = shops.filter((s) => s.presence === 'offline')
  const shopNote = shops.find((s) => s.note)?.note ?? ''

  return (
    <section className="flex flex-col gap-5">
      <header className="border-b border-border pb-4">
        <h1 className="font-heading m-0 text-[clamp(1.9rem,3.2vw,2.45rem)] font-bold tracking-wide">
          Bean Scout
        </h1>
        <p className="mt-2 max-w-xl text-[1.02rem] text-muted-foreground">
          Filter by budget and/or weight.
        </p>
      </header>

      <form
        className="grid grid-cols-1 items-end gap-4 pt-1 md:grid-cols-[1fr_1fr_1fr_auto]"
        onSubmit={onSubmit}
      >
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="budget">Budget (IDR)</Label>
          <Input
            id="budget"
            type="number"
            min={0}
            step={100000}
            placeholder="Optional"
            value={budget}
            onChange={(e) => setBudget(e.target.value)}
            className="rounded-none"
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="weight">Weight (kg)</Label>
          <Input
            id="weight"
            type="number"
            min={0}
            step={1}
            placeholder="Optional"
            value={weightKg}
            onChange={(e) => setWeightKg(e.target.value)}
            className="rounded-none"
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label>Channel</Label>
          <Select value={channel} onValueChange={setChannel}>
            <SelectTrigger className="w-full rounded-none">
              <SelectValue placeholder="All" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All</SelectItem>
              <SelectItem value="farmer">Farmer</SelectItem>
              <SelectItem value="middleman">Middleman</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <Button type="submit" disabled={loading} className="rounded-none" size="lg">
          {loading ? 'Analyzing…' : 'Run scout'}
        </Button>
      </form>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {recs.length === 0 && !error && (
        <p className="border-y border-dashed border-border py-5 italic text-muted-foreground">
          Set budget and/or weight (or leave both empty), then run scout.
        </p>
      )}

      {recs.length > 0 && (
        <div className="grid grid-cols-1 items-start gap-4 md:grid-cols-[0.95fr_1.05fr]">
          <Card className="rounded-none shadow-none">
            <CardHeader className="pb-2">
              <CardTitle className="font-heading text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
                Recommendations
              </CardTitle>
              <Separator />
            </CardHeader>
            <CardContent>
              <ul className="m-0 flex list-none flex-col p-0">
                {recs.map((r) => (
                  <li key={r.bean.id} className="border-t border-border first:border-t-0">
                    <button
                      type="button"
                      className={cn(
                        'flex w-full items-center gap-3 bg-transparent px-1.5 py-3.5 text-left hover:bg-accent',
                        selected?.bean.id === r.bean.id &&
                          'bg-accent shadow-[inset_2px_0_0_var(--destructive)]',
                      )}
                      onClick={() => setSelected(r)}
                    >
                      <span className="h-2.5 w-2.5 shrink-0 rounded-full bg-destructive" />
                      <span className="min-w-0 flex-1">
                        <strong className="font-heading block font-semibold">{r.bean.origin}</strong>
                        <small className="block text-sm text-muted-foreground">
                          {r.bean.variety} · IDR {r.bean.price_per_kg.toLocaleString('en-US')}/kg
                        </small>
                      </span>
                      <Badge variant="outline" className="min-w-9 justify-center rounded-none">
                        {r.fit_score.toFixed(0)}
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
                <div className="mb-4 flex items-start justify-between gap-4 border-b border-border pb-4">
                  <div>
                    <h2 className="font-heading m-0 text-[1.45rem] font-semibold">
                      {selected.bean.origin}
                    </h2>
                    <p className="mt-1 text-muted-foreground">
                      {labelChannel(selected.bean.channel)} · profit ~IDR{' '}
                      {selected.predicted_profit_per_kg.toLocaleString('en-US')}/kg
                    </p>
                  </div>
                  <div className="text-right">
                    <span className="block text-xs uppercase tracking-wider text-muted-foreground">
                      Fit score
                    </span>
                    <strong className="font-heading text-4xl font-bold text-destructive">
                      {selected.fit_score.toFixed(0)}
                    </strong>
                  </div>
                </div>

                <ChartContainer config={fitChartConfig} className="mb-4 aspect-auto h-44 w-full">
                  <BarChart data={chartData} accessibilityLayer>
                    <CartesianGrid vertical={false} />
                    <XAxis dataKey="name" tickLine={false} axisLine={false} />
                    <YAxis domain={[0, 100]} width={28} tickLine={false} axisLine={false} />
                    <ChartTooltip content={<ChartTooltipContent />} />
                    <Bar dataKey="value" fill="var(--color-value)" radius={0} />
                  </BarChart>
                </ChartContainer>

                <dl className="m-0 grid gap-3.5">
                  {(
                    [
                      ['Strengths', selected.report.strengths],
                      ['Risks', selected.report.risks],
                      ['Potential', selected.report.potential],
                    ] as const
                  ).map(([title, body]) => (
                    <div key={title}>
                      <dt className="font-heading text-base font-bold text-destructive">{title}</dt>
                      <dd className="m-0 border-b border-border pb-2 pt-1">{body}</dd>
                    </div>
                  ))}
                </dl>

                <div className="mt-5 border-t border-border pt-4">
                  <h3 className="font-heading mb-3 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
                    Linked shops · live search
                  </h3>
                  {shopsLoading && <p className="text-muted-foreground">Searching online…</p>}
                  {shopsError && (
                    <Alert variant="destructive" className="mb-3">
                      <AlertDescription>{shopsError}</AlertDescription>
                    </Alert>
                  )}
                  {!shopsLoading && !shopsError && shopNote && (
                    <p className="mb-3 text-sm text-muted-foreground">{shopNote}</p>
                  )}
                  {!shopsLoading && !shopsError && (
                    <div className="grid gap-4">
                      <ShopGroup title="Online" items={onlineShops} />
                      <ShopGroup title="Offline" items={offlineShops} />
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      )}
    </section>
  )
}

function ShopGroup({ title, items }: { title: string; items: ScoutShop[] }) {
  if (items.length === 0) {
    return (
      <div>
        <Badge variant="outline" className="rounded-none">
          {title}
        </Badge>
        <p className="mt-2 text-muted-foreground">
          None found. Use the Google link when available.
        </p>
      </div>
    )
  }
  return (
    <div>
      <Badge variant="outline" className="rounded-none">
        {title}
      </Badge>
      <ul className="m-0 mt-2 flex list-none flex-col gap-2 p-0">
        {items.map((s) => (
          <li key={`${s.presence}-${s.url}`} className="border-b border-border pb-2">
            <a
              href={s.url}
              target="_blank"
              rel="noreferrer"
              className="font-heading font-semibold underline decoration-border underline-offset-2 hover:text-destructive"
            >
              {s.name}
            </a>
            {s.snippet && <p className="m-0 mt-1 text-sm text-muted-foreground">{s.snippet}</p>}
          </li>
        ))}
      </ul>
    </div>
  )
}
