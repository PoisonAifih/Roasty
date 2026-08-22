import { useEffect, useState } from 'react'
import { api, type Notification } from '@/api/client'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'

const severityLabel: Record<string, string> = {
  high: 'High',
  med: 'Med',
  low: 'Low',
}

const kindLabel: Record<string, string> = {
  low_stock: 'Low Stock',
  payment_overdue: 'Payment Overdue',
  reorder_due: 'Reorder Due',
}

export function NotificationsPage() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [items, setItems] = useState<Notification[]>([])
  const [unread, setUnread] = useState(0)
  const [markingAll, setMarkingAll] = useState(false)

  async function load() {
    setLoading(true)
    setError('')
    try {
      const data = await api.notifications()
      setItems(data.notifications ?? [])
      setUnread(data.unread)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load notifications')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  async function markRead(id: string) {
    try {
      await api.markRead(id)
      setItems((prev) => prev.map((n) => (n.id === id ? { ...n, read: true } : n)))
      setUnread((c) => Math.max(0, c - 1))
    } catch {
      // ignore — notification stays unread visually
    }
  }

  async function markAll() {
    setMarkingAll(true)
    try {
      await api.markAllRead()
      setItems((prev) => prev.map((n) => ({ ...n, read: true })))
      setUnread(0)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to mark all read')
    } finally {
      setMarkingAll(false)
    }
  }

  return (
    <section className="flex flex-col gap-5">
      <header className="border-b border-border pb-4">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="font-heading m-0 text-[clamp(1.9rem,3.2vw,2.45rem)] font-bold tracking-wide">
              Notifications
            </h1>
            <p className="mt-2 max-w-xl text-[1.02rem] text-muted-foreground">
              Alerts from inventory and CRM checks — updated every hour.
            </p>
          </div>
          {unread > 0 && (
            <Button
              variant="outline"
              className="mt-1 rounded-none"
              onClick={markAll}
              disabled={markingAll}
            >
              {markingAll ? 'Marking…' : `Mark all read (${unread})`}
            </Button>
          )}
        </div>
      </header>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {loading && (
        <p className="border-y border-dashed border-border py-5 italic text-muted-foreground">
          Loading…
        </p>
      )}

      {!loading && items.length === 0 && !error && (
        <p className="border-y border-dashed border-border py-5 italic text-muted-foreground">
          No notifications yet. The cron runs on startup and every hour.
        </p>
      )}

      {!loading && items.length > 0 && (
        <Card className="rounded-none shadow-none">
          <CardHeader className="pb-2">
            <CardTitle className="font-heading text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
              Recent alerts
            </CardTitle>
            <Separator />
          </CardHeader>
          <CardContent className="p-0">
            <ul className="m-0 list-none p-0">
              {items.map((n) => (
                <li
                  key={n.id}
                  className={cn(
                    'flex items-start gap-4 border-t border-border px-5 py-4 first:border-t-0',
                    !n.read && 'bg-accent/40',
                  )}
                >
                  <span
                    className={cn(
                      'mt-1.5 h-2 w-2 shrink-0 rounded-full',
                      n.severity === 'high'
                        ? 'bg-destructive'
                        : n.severity === 'med'
                          ? 'bg-yellow-500'
                          : 'bg-secondary',
                    )}
                  />
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="font-heading font-semibold">{n.title}</span>
                      <Badge variant="outline" className="rounded-none text-[0.7rem]">
                        {kindLabel[n.kind] ?? n.kind}
                      </Badge>
                      <Badge
                        variant="outline"
                        className={cn(
                          'rounded-none text-[0.7rem]',
                          n.severity === 'high' && 'border-destructive text-destructive',
                        )}
                      >
                        {severityLabel[n.severity] ?? n.severity}
                      </Badge>
                    </div>
                    <p className="mt-1 text-sm text-muted-foreground">{n.body}</p>
                    <span className="mt-1 block text-xs text-muted-foreground/60">
                      {new Date(n.created_at).toLocaleString()}
                    </span>
                  </div>
                  {!n.read && (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="shrink-0 rounded-none text-xs"
                      onClick={() => void markRead(n.id)}
                    >
                      Dismiss
                    </Button>
                  )}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}
    </section>
  )
}
