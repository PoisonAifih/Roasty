import { useEffect, useState } from 'react'
import { api, type CRMFollowUp } from '@/api/client'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'

// Indonesian roasteries run on WhatsApp, so the follow-up note is only useful
// if it can actually be sent. Prefill the message and let the owner edit it.
function waLink(c: CRMFollowUp) {
  const greeting = `Halo ${c.shop_name}, selamat siang!`
  const body =
    c.payment_status === 'overdue'
      ? 'Kami ingin menindaklanjuti pembayaran yang masih tertunda. Mohon infonya ya. Terima kasih!'
      : `Biasanya pesanan berikutnya sekitar ${c.predicted_reorder_date}. Apakah mau kami siapkan stoknya?`
  return `https://wa.me/${c.phone}?text=${encodeURIComponent(`${greeting} ${body}`)}`
}

export function CRMPage() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [items, setItems] = useState<CRMFollowUp[]>([])
  const [selected, setSelected] = useState<CRMFollowUp | null>(null)
  const [marking, setMarking] = useState(false)

  async function markDone(shopId: string) {
    setMarking(true)
    setError('')
    try {
      await api.markContacted(shopId)
      await load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to mark contacted')
    } finally {
      setMarking(false)
    }
  }

  async function load() {
    setLoading(true)
    setError('')
    try {
      const data = await api.crm()
      setItems(data)
      setSelected((prev) => data.find((d) => d.shop_id === prev?.shop_id) ?? data[0] ?? null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load CRM')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  const followUps = items.filter((i) => i.needs_follow_up).length
  const payRisk = items.filter((i) => i.late_pay_risk).length

  return (
    <section className="flex flex-col gap-5">
      <header className="flex flex-wrap items-end justify-between gap-4 border-b border-border pb-4">
        <div>
          <h1 className="font-heading m-0 text-[clamp(1.9rem,3.2vw,2.45rem)] font-bold tracking-wide">
            Coffee Shop Customer Tracker
          </h1>
          <p className="mt-2 max-w-xl text-[1.02rem] text-muted-foreground">
            Manage customer relationships and payment status.
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
            {followUps} need follow-up
          </Badge>
          <Badge variant="destructive" className="rounded-none">
            {payRisk} payment risk
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
          Loading coffee shop data…
        </p>
      )}
      {!loading && items.length === 0 && !error && (
        <p className="border-y border-dashed border-border py-5 italic text-muted-foreground">
          No coffee shop data yet.
        </p>
      )}

      {items.length > 0 && (
        <div className="grid grid-cols-1 items-start gap-4 md:grid-cols-[0.95fr_1.05fr]">
          <Card className="rounded-none shadow-none">
            <CardHeader className="pb-2">
              <CardTitle className="font-heading text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
                Coffee shops
              </CardTitle>
              <Separator />
            </CardHeader>
            <CardContent>
              <ul className="m-0 flex list-none flex-col p-0">
                {items.map((i) => (
                  <li key={i.shop_id} className="border-t border-border first:border-t-0">
                    <button
                      type="button"
                      className={cn(
                        'flex w-full items-center gap-3 bg-transparent px-1.5 py-3.5 text-left hover:bg-accent',
                        selected?.shop_id === i.shop_id &&
                          'bg-accent shadow-[inset_2px_0_0_var(--destructive)]',
                      )}
                      onClick={() => setSelected(i)}
                    >
                      <span className="h-2.5 w-2.5 shrink-0 rounded-full bg-destructive" />
                      <span className="min-w-0 flex-1">
                        <strong className="font-heading block font-semibold">{i.shop_name}</strong>
                        <small className="block text-sm text-muted-foreground">
                          Predicted {i.predicted_reorder_date}
                        </small>
                      </span>
                      <Badge
                        variant={
                          i.payment_status === 'overdue'
                            ? 'destructive'
                            : i.payment_status === 'credit'
                              ? 'secondary'
                              : 'outline'
                        }
                      >
                        {i.payment_status}
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
                <div className="mb-1 flex items-center justify-between gap-3 border-b border-border pb-3.5">
                  <h2 className="font-heading m-0 text-[1.45rem] font-semibold">
                    {selected.shop_name}
                  </h2>
                  <Badge
                    variant={
                      selected.payment_status === 'overdue'
                        ? 'destructive'
                        : selected.payment_status === 'credit'
                          ? 'secondary'
                          : 'outline'
                    }
                  >
                    {selected.payment_status}
                  </Badge>
                </div>
                <div className="my-4 grid grid-cols-2 border border-border">
                  {[
                    ['Average interval', `${selected.avg_interval_days.toFixed(0)} days`],
                    ['Since last order', `${selected.days_since_order.toFixed(0)} days`],
                    ['Predicted reorder', selected.predicted_reorder_date],
                    ['Status', selected.needs_follow_up ? 'Needs follow-up' : 'On track'],
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
                    Follow-up note
                  </h3>
                  <p className="m-0">{selected.message}</p>

                  <div className="mt-4 flex flex-wrap items-center gap-2">
                    {selected.phone ? (
                      <Button asChild className="rounded-none">
                        <a href={waLink(selected)} target="_blank" rel="noreferrer">
                          Send WhatsApp
                        </a>
                      </Button>
                    ) : (
                      <span className="text-sm text-muted-foreground">
                        No phone number on file.
                      </span>
                    )}
                    <Button
                      variant="outline"
                      className="rounded-none"
                      disabled={marking}
                      onClick={() => void markDone(selected.shop_id)}
                    >
                      {marking ? 'Saving…' : 'Mark contacted'}
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      )}
    </section>
  )
}
