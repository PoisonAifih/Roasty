import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type Bean, type CRMFollowUp, type InventorySuggestion } from '@/api/client'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'

const modules = [
  {
    to: '/scout',
    title: 'Bean Scout',
    desc: 'Search where to buy coffee beans. ',
  },
  {
    to: '/inventory',
    title: 'Smart Inventory',
    desc: 'Monitor stock and restock suggestions.',
  },
  {
    to: '/crm',
    title: 'Customer & Payment',
    desc: 'Customer reorders and payment risk.',
  },
]

type Summary = {
  beans: Bean[]
  inventory: InventorySuggestion[]
  crm: CRMFollowUp[]
}

export function HomePage() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [data, setData] = useState<Summary | null>(null)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      setLoading(true)
      setError('')
      try {
        const [beans, inventory, crm] = await Promise.all([
          api.beans(),
          api.inventory(),
          api.crm(),
        ])
        if (!cancelled) setData({ beans, inventory, crm })
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Failed to load summary')
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [])

  const beansCount = data?.beans.length ?? 0
  const totalStock = data?.beans.reduce((s, b) => s + b.stock_kg, 0) ?? 0
  const highUrgency = data?.inventory.filter((i) => i.urgency === 'high') ?? []
  const medUrgency = data?.inventory.filter((i) => i.urgency === 'med').length ?? 0
  const shops = data?.crm.length ?? 0
  const followUps = data?.crm.filter((c) => c.needs_follow_up) ?? []
  const payRisk = data?.crm.filter((c) => c.late_pay_risk) ?? []
  const overdue = data?.crm.filter((c) => c.payment_status === 'overdue').length ?? 0

  const topRestock = (data?.inventory ?? []).slice(0, 3)
  const topFollow = [...(data?.crm ?? [])]
    .sort(
      (a, b) =>
        Number(b.needs_follow_up) - Number(a.needs_follow_up) ||
        Number(b.late_pay_risk) - Number(a.late_pay_risk),
    )
    .slice(0, 3)

  return (
    <section className="flex flex-col gap-5">
      <header className="border-b border-border pb-4">
        <h1 className="font-heading m-0 text-[clamp(1.9rem,3.2vw,2.45rem)] font-bold tracking-wide">
          Dashboard
        </h1>
      </header>

      {loading && (
        <p className="border-y border-dashed border-border py-5 italic text-muted-foreground">
          Loading summary…
        </p>
      )}
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {data && (
        <>
          <Card className="rounded-none py-0 shadow-none">
            <CardContent className="grid grid-cols-1 p-0 sm:grid-cols-2 lg:grid-cols-4">
              <Stat
                label="Bean origins"
                value={beansCount}
                hint={`${totalStock.toFixed(0)} kg total stock`}
                className="border-b sm:border-r lg:border-b-0"
              />
              <Stat
                label="Critical restock"
                value={highUrgency.length}
                hint={`${medUrgency} medium urgency`}
                emphasize={highUrgency.length > 0}
                className="border-b lg:border-b-0 lg:border-r"
              />
              <Stat
                label="Coffee shops"
                value={shops}
                hint={`${followUps.length} need follow-up`}
                className="border-b sm:border-r sm:border-b-0 lg:border-r"
              />
              <Stat
                label="Payment risk"
                value={payRisk.length}
                hint={`${overdue} overdue`}
                emphasize={payRisk.length > 0}
              />
            </CardContent>
          </Card>

          <div className="grid grid-cols-1 items-start gap-4 md:grid-cols-2">
            <Card className="rounded-none shadow-none">
              <CardHeader className="pb-2">
                <CardTitle className="font-heading text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
                  Restock priority
                </CardTitle>
                <Separator />
              </CardHeader>
              <CardContent>
                {topRestock.length === 0 ? (
                  <p className="text-muted-foreground">No stock data.</p>
                ) : (
                  <ul className="m-0 flex list-none flex-col p-0">
                    {topRestock.map((i) => (
                      <li key={i.bean_id} className="border-t border-border first:border-t-0">
                        <Link
                          to="/inventory"
                          className="flex w-full items-center gap-3 px-1.5 py-3.5 hover:bg-accent"
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
                        </Link>
                      </li>
                    ))}
                  </ul>
                )}
              </CardContent>
            </Card>

            <Card className="rounded-none shadow-none">
              <CardHeader className="pb-2">
                <CardTitle className="font-heading text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
                  Customer follow-up
                </CardTitle>
                <Separator />
              </CardHeader>
              <CardContent>
                {topFollow.length === 0 ? (
                  <p className="text-muted-foreground">No CRM data.</p>
                ) : (
                  <ul className="m-0 flex list-none flex-col p-0">
                    {topFollow.map((c) => (
                      <li key={c.shop_id} className="border-t border-border first:border-t-0">
                        <Link
                          to="/crm"
                          className="flex w-full items-center gap-3 px-1.5 py-3.5 hover:bg-accent"
                        >
                          <span className="h-2.5 w-2.5 shrink-0 rounded-full bg-destructive" />
                          <span className="min-w-0 flex-1">
                            <strong className="font-heading block font-semibold">{c.shop_name}</strong>
                            <small className="block text-sm text-muted-foreground">
                              {c.needs_follow_up ? 'Needs follow-up' : 'On track'} · predicted{' '}
                              {c.predicted_reorder_date}
                            </small>
                          </span>
                          <Badge
                            variant={
                              c.payment_status === 'overdue'
                                ? 'destructive'
                                : c.payment_status === 'credit'
                                  ? 'secondary'
                                  : 'outline'
                            }
                          >
                            {c.payment_status}
                          </Badge>
                        </Link>
                      </li>
                    ))}
                  </ul>
                )}
              </CardContent>
            </Card>
          </div>
        </>
      )}

      <div className="grid grid-cols-1 gap-4 pt-1 md:grid-cols-3">
        {modules.map((m) => (
          <Card
            key={m.to}
            className="rounded-none shadow-none transition hover:border-destructive"
          >
            <CardHeader>
              <CardTitle className="font-heading text-[1.35rem]">{m.title}</CardTitle>
              <CardDescription>{m.desc}</CardDescription>
            </CardHeader>
            <CardContent>
              <Link
                to={m.to}
                className="font-heading border-t border-border pt-2.5 text-[0.92rem] font-semibold text-destructive"
              >
                Open module →
              </Link>
            </CardContent>
          </Card>
        ))}
      </div>
    </section>
  )
}

function Stat({
  label,
  value,
  hint,
  emphasize,
  className,
}: {
  label: string
  value: number
  hint: string
  emphasize?: boolean
  className?: string
}) {
  return (
    <div className={cn('border-border p-4', className)}>
      <span className="block text-xs uppercase tracking-wider text-muted-foreground">{label}</span>
      <strong
        className={cn(
          'font-heading mt-1 block text-3xl font-bold leading-tight',
          emphasize && 'text-destructive',
        )}
      >
        {value}
      </strong>
      <small className="text-sm text-muted-foreground">{hint}</small>
    </div>
  )
}
