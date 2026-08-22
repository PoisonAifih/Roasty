import { useEffect, useState, type FormEvent } from 'react'
import { api, type Bean, type StockAdjustment } from '@/api/client'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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

type Mode = { type: 'view'; bean: Bean } | { type: 'add' }

export function BeansPage() {
  const [beans, setBeans] = useState<Bean[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [mode, setMode] = useState<Mode | null>(null)

  async function load() {
    setLoading(true)
    setError('')
    try {
      const data = await api.allBeans()
      setBeans(data)
      setMode((prev) => {
        if (prev?.type === 'view') {
          const fresh = data.find((b) => b.id === prev.bean.id)
          return fresh ? { type: 'view', bean: fresh } : { type: 'view', bean: data[0] }
        }
        if (prev?.type === 'add') return prev
        return data.length > 0 ? { type: 'view', bean: data[0] } : null
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load beans')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  const activeCount = beans.filter((b) => b.active !== false).length

  return (
    <section className="flex flex-col gap-5">
      <header className="flex flex-wrap items-end justify-between gap-4 border-b border-border pb-4">
        <div>
          <h1 className="font-heading m-0 text-[clamp(1.9rem,3.2vw,2.45rem)] font-bold tracking-wide">
            Bean Management
          </h1>
          <p className="mt-2 max-w-xl text-[1.02rem] text-muted-foreground">
            Add, edit prices, or disable origins.
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            className="rounded-none"
            onClick={() => void load()}
            disabled={loading}
          >
            {loading ? 'Loading…' : 'Refresh'}
          </Button>
          <Button
            className="rounded-none"
            onClick={() => setMode({ type: 'add' })}
            disabled={mode?.type === 'add'}
          >
            Add Bean
          </Button>
        </div>
      </header>

      {beans.length > 0 && (
        <div className="flex flex-wrap gap-2.5">
          <Badge variant="secondary" className="rounded-none">
            {activeCount} active
          </Badge>
          <Badge variant="outline" className="rounded-none">
            {beans.length - activeCount} disabled
          </Badge>
        </div>
      )}

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {loading && beans.length === 0 && (
        <p className="border-y border-dashed border-border py-5 italic text-muted-foreground">
          Loading…
        </p>
      )}

      {!loading && beans.length === 0 && !error && (
        <p className="border-y border-dashed border-border py-5 italic text-muted-foreground">
          No beans yet. Add one to get started.
        </p>
      )}

      {(beans.length > 0 || mode?.type === 'add') && (
        <div className="grid grid-cols-1 items-start gap-4 md:grid-cols-[0.95fr_1.05fr]">
          {/* Left: bean list */}
          <Card className="rounded-none shadow-none">
            <CardHeader className="pb-2">
              <CardTitle className="font-heading text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
                All beans
              </CardTitle>
              <Separator />
            </CardHeader>
            <CardContent>
              <ul className="m-0 flex list-none flex-col p-0">
                {beans.map((b) => (
                  <li key={b.id} className="border-t border-border first:border-t-0">
                    <button
                      type="button"
                      className={cn(
                        'flex w-full items-center gap-3 bg-transparent px-1.5 py-3.5 text-left hover:bg-accent',
                        mode?.type === 'view' &&
                          mode.bean.id === b.id &&
                          'bg-accent shadow-[inset_2px_0_0_var(--destructive)]',
                        b.active === false && 'opacity-50',
                      )}
                      onClick={() => setMode({ type: 'view', bean: b })}
                    >
                      <span
                        className={cn(
                          'h-2.5 w-2.5 shrink-0 rounded-full',
                          b.active === false ? 'bg-muted-foreground' : 'bg-destructive',
                        )}
                      />
                      <span className="min-w-0 flex-1">
                        <strong className="font-heading block font-semibold">{b.origin}</strong>
                        <small className="block text-sm text-muted-foreground">
                          {b.variety} · IDR {b.price_per_kg.toLocaleString('en-US')}/kg
                        </small>
                      </span>
                      {b.active === false && (
                        <Badge variant="outline" className="rounded-none text-[0.7rem]">
                          disabled
                        </Badge>
                      )}
                    </button>
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>

          {/* Right: edit or add panel */}
          {mode?.type === 'view' && (
            <EditPanel key={mode.bean.id} bean={mode.bean} onMutate={load} />
          )}
          {mode?.type === 'add' && (
            <AddPanel
              onMutate={load}
              onCancel={() =>
                setMode(beans.length > 0 ? { type: 'view', bean: beans[0] } : null)
              }
            />
          )}
        </div>
      )}
    </section>
  )
}

// ── Edit panel ────────────────────────────────────────────────────────────────

function EditPanel({ bean, onMutate }: { bean: Bean; onMutate: () => Promise<void> }) {
  const [form, setForm] = useState({
    origin: bean.origin,
    variety: bean.variety,
    channel: bean.channel || 'farmer',
    price_per_kg: String(bean.price_per_kg),
    sell_price_per_kg: String(bean.sell_price_per_kg),
    harvest_estimate_kg: String(bean.harvest_estimate_kg),
    humidity: String(bean.humidity),
    quality_score: String(bean.quality_score),
  })
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null)
  const [toggleBusy, setToggleBusy] = useState(false)
  const [adjustments, setAdjustments] = useState<StockAdjustment[]>([])

  useEffect(() => {
    api
      .beanAdjustments(bean.id)
      .then(setAdjustments)
      .catch(() => {})
  }, [bean.id])

  function set(k: keyof typeof form, v: string) {
    setForm((f) => ({ ...f, [k]: v }))
    setMsg(null)
  }

  async function onSave(e: FormEvent) {
    e.preventDefault()
    const pricePerKg = parseFloat(form.price_per_kg)
    const sellPricePerKg = parseFloat(form.sell_price_per_kg)
    if (isNaN(pricePerKg) || pricePerKg <= 0) {
      setMsg({ ok: false, text: 'Buy price must be > 0.' })
      return
    }
    if (isNaN(sellPricePerKg) || sellPricePerKg <= 0) {
      setMsg({ ok: false, text: 'Sell price must be > 0.' })
      return
    }
    setBusy(true)
    setMsg(null)
    try {
      await api.updateBean(bean.id, {
        origin: form.origin,
        variety: form.variety,
        channel: form.channel,
        price_per_kg: pricePerKg,
        sell_price_per_kg: sellPricePerKg,
        harvest_estimate_kg: parseFloat(form.harvest_estimate_kg) || 0,
        humidity: parseFloat(form.humidity) || 60,
        quality_score: parseFloat(form.quality_score) || 70,
      })
      setMsg({ ok: true, text: 'Saved.' })
      await onMutate()
    } catch (err) {
      setMsg({ ok: false, text: err instanceof Error ? err.message : 'Failed to save.' })
    } finally {
      setBusy(false)
    }
  }

  async function onToggle() {
    setToggleBusy(true)
    setMsg(null)
    try {
      if (bean.active === false) {
        await api.restoreBean(bean.id)
      } else {
        await api.disableBean(bean.id)
      }
      await onMutate()
    } catch (err) {
      setMsg({ ok: false, text: err instanceof Error ? err.message : 'Failed.' })
    } finally {
      setToggleBusy(false)
    }
  }

  return (
    <Card className="rounded-none shadow-none">
      <CardContent className="pt-6">
        <div className="mb-4 flex items-start justify-between gap-4 border-b border-border pb-4">
          <div>
            <h2 className="font-heading m-0 text-[1.45rem] font-semibold">{bean.origin}</h2>
            <p className="mt-1 text-muted-foreground">{bean.variety}</p>
          </div>
          <Button
            variant="outline"
            size="sm"
            className={cn(
              'mt-1 rounded-none',
              bean.active !== false && 'border-destructive text-destructive hover:bg-destructive/10',
            )}
            onClick={() => void onToggle()}
            disabled={toggleBusy}
          >
            {toggleBusy ? '…' : bean.active === false ? 'Restore' : 'Disable'}
          </Button>
        </div>

        <form onSubmit={(e) => void onSave(e)} className="flex flex-col gap-3">
          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1.5">
              <Label>Origin</Label>
              <Input
                value={form.origin}
                onChange={(e) => set('origin', e.target.value)}
                className="rounded-none"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Variety</Label>
              <Input
                value={form.variety}
                onChange={(e) => set('variety', e.target.value)}
                className="rounded-none"
              />
            </div>
          </div>

          <div className="flex flex-col gap-1.5">
            <Label>Channel</Label>
            <Select value={form.channel} onValueChange={(v) => set('channel', v)}>
              <SelectTrigger className="w-full rounded-none">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="farmer">Farmer</SelectItem>
                <SelectItem value="middleman">Middleman</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1.5">
              <Label>Buy price (IDR/kg)</Label>
              <Input
                type="number"
                min={1}
                step={500}
                value={form.price_per_kg}
                onChange={(e) => set('price_per_kg', e.target.value)}
                className="rounded-none"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Sell price (IDR/kg)</Label>
              <Input
                type="number"
                min={1}
                step={500}
                value={form.sell_price_per_kg}
                onChange={(e) => set('sell_price_per_kg', e.target.value)}
                className="rounded-none"
              />
            </div>
          </div>

          <div className="grid grid-cols-3 gap-3">
            <div className="flex flex-col gap-1.5">
              <Label>Harvest est. (kg)</Label>
              <Input
                type="number"
                min={0}
                step={10}
                value={form.harvest_estimate_kg}
                onChange={(e) => set('harvest_estimate_kg', e.target.value)}
                className="rounded-none"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Humidity (%)</Label>
              <Input
                type="number"
                min={0}
                max={100}
                step={1}
                value={form.humidity}
                onChange={(e) => set('humidity', e.target.value)}
                className="rounded-none"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Quality score</Label>
              <Input
                type="number"
                min={0}
                max={100}
                step={1}
                value={form.quality_score}
                onChange={(e) => set('quality_score', e.target.value)}
                className="rounded-none"
              />
            </div>
          </div>

          {msg && (
            <p className={cn('text-sm', msg.ok ? 'text-green-600 dark:text-green-400' : 'text-destructive')}>
              {msg.text}
            </p>
          )}

          <Button type="submit" disabled={busy} className="rounded-none">
            {busy ? 'Saving…' : 'Save changes'}
          </Button>
        </form>

        {/* Adjustment history */}
        <div className="mt-5 border-t border-border pt-4">
          <h3 className="font-heading mb-2 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
            Stock adjustment history
          </h3>
          {adjustments.length === 0 ? (
            <p className="text-sm text-muted-foreground">No adjustments yet.</p>
          ) : (
            <ul className="m-0 list-none p-0">
              {adjustments.slice(0, 10).map((a) => (
                <li key={a.id} className="flex items-baseline gap-3 border-t border-border py-2 first:border-t-0 text-sm">
                  <span className="w-[7rem] shrink-0 text-xs text-muted-foreground">
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

// ── Add panel ─────────────────────────────────────────────────────────────────

function AddPanel({ onMutate, onCancel }: { onMutate: () => Promise<void>; onCancel: () => void }) {
  const [form, setForm] = useState({
    origin: '',
    variety: '',
    channel: 'farmer',
    price_per_kg: '',
    sell_price_per_kg: '',
    stock_kg: '0',
    harvest_estimate_kg: '0',
    humidity: '60',
    quality_score: '70',
  })
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null)

  function set(k: keyof typeof form, v: string) {
    setForm((f) => ({ ...f, [k]: v }))
    setMsg(null)
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    const pricePerKg = parseFloat(form.price_per_kg)
    const sellPricePerKg = parseFloat(form.sell_price_per_kg)
    if (!form.origin.trim()) {
      setMsg({ ok: false, text: 'Origin is required.' })
      return
    }
    if (!form.variety.trim()) {
      setMsg({ ok: false, text: 'Variety is required.' })
      return
    }
    if (isNaN(pricePerKg) || pricePerKg <= 0) {
      setMsg({ ok: false, text: 'Buy price must be > 0.' })
      return
    }
    if (isNaN(sellPricePerKg) || sellPricePerKg <= 0) {
      setMsg({ ok: false, text: 'Sell price must be > 0.' })
      return
    }
    setBusy(true)
    setMsg(null)
    try {
      await api.createBean({
        origin: form.origin.trim(),
        variety: form.variety.trim(),
        channel: form.channel,
        price_per_kg: pricePerKg,
        sell_price_per_kg: sellPricePerKg,
        stock_kg: parseFloat(form.stock_kg) || 0,
        harvest_estimate_kg: parseFloat(form.harvest_estimate_kg) || 0,
        humidity: parseFloat(form.humidity) || 60,
        quality_score: parseFloat(form.quality_score) || 70,
      })
      setMsg({ ok: true, text: 'Bean added.' })
      await onMutate()
    } catch (err) {
      setMsg({ ok: false, text: err instanceof Error ? err.message : 'Failed to add bean.' })
      setBusy(false)
    }
  }

  return (
    <Card className="rounded-none shadow-none">
      <CardHeader className="pb-2">
        <CardTitle className="font-heading text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
          New bean
        </CardTitle>
        <Separator />
      </CardHeader>
      <CardContent>
        <form onSubmit={(e) => void onSubmit(e)} className="flex flex-col gap-3">
          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1.5">
              <Label>Origin *</Label>
              <Input
                placeholder="e.g. Gayo, Aceh"
                value={form.origin}
                onChange={(e) => set('origin', e.target.value)}
                className="rounded-none"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Variety *</Label>
              <Input
                placeholder="e.g. Arabica"
                value={form.variety}
                onChange={(e) => set('variety', e.target.value)}
                className="rounded-none"
              />
            </div>
          </div>

          <div className="flex flex-col gap-1.5">
            <Label>Channel</Label>
            <Select value={form.channel} onValueChange={(v) => set('channel', v)}>
              <SelectTrigger className="w-full rounded-none">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="farmer">Farmer</SelectItem>
                <SelectItem value="middleman">Middleman</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1.5">
              <Label>Buy price (IDR/kg) *</Label>
              <Input
                type="number"
                min={1}
                step={500}
                placeholder="e.g. 85000"
                value={form.price_per_kg}
                onChange={(e) => set('price_per_kg', e.target.value)}
                className="rounded-none"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Sell price (IDR/kg) *</Label>
              <Input
                type="number"
                min={1}
                step={500}
                placeholder="e.g. 145000"
                value={form.sell_price_per_kg}
                onChange={(e) => set('sell_price_per_kg', e.target.value)}
                className="rounded-none"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1.5">
              <Label>Initial stock (kg)</Label>
              <Input
                type="number"
                min={0}
                step={0.1}
                value={form.stock_kg}
                onChange={(e) => set('stock_kg', e.target.value)}
                className="rounded-none"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Harvest est. (kg)</Label>
              <Input
                type="number"
                min={0}
                step={10}
                value={form.harvest_estimate_kg}
                onChange={(e) => set('harvest_estimate_kg', e.target.value)}
                className="rounded-none"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1.5">
              <Label>Humidity (%)</Label>
              <Input
                type="number"
                min={0}
                max={100}
                step={1}
                value={form.humidity}
                onChange={(e) => set('humidity', e.target.value)}
                className="rounded-none"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label>Quality score (0–100)</Label>
              <Input
                type="number"
                min={0}
                max={100}
                step={1}
                value={form.quality_score}
                onChange={(e) => set('quality_score', e.target.value)}
                className="rounded-none"
              />
            </div>
          </div>

          {msg && (
            <p className={cn('text-sm', msg.ok ? 'text-green-600 dark:text-green-400' : 'text-destructive')}>
              {msg.text}
            </p>
          )}

          <div className="flex gap-2">
            <Button type="submit" disabled={busy} className="flex-1 rounded-none">
              {busy ? 'Adding…' : 'Add Bean'}
            </Button>
            <Button type="button" variant="outline" className="rounded-none" onClick={onCancel}>
              Cancel
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
