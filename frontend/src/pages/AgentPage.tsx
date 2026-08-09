import { useEffect, useRef, useState, type FormEvent } from 'react'
import { streamAgent, type TraceEvent } from '@/api/client'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'

const examples = [
  'I have 50 million rupiah to restock for the next three months. What should I buy?',
  'Which origin is closest to running out, and is it worth reordering?',
  'Compare Gayo and Toraja for a high-margin single origin.',
]

// The tool trace is the whole point of this page: without it the agent's work
// is invisible and indistinguishable from a single chat completion.
const toolLabels: Record<string, string> = {
  list_beans: 'Reading bean catalogue',
  score_beans: 'Scoring beans against constraints',
  check_inventory: 'Checking stock levels',
  sales_history: 'Pulling sales history',
  build_basket: 'Allocating budget across origins',
  market_price: 'Searching live market listings',
}

export function AgentPage() {
  const [goal, setGoal] = useState('')
  const [running, setRunning] = useState(false)
  const [events, setEvents] = useState<TraceEvent[]>([])
  const closeRef = useRef<(() => void) | null>(null)
  const endRef = useRef<HTMLDivElement | null>(null)

  // Close the stream if the user navigates away mid-run.
  useEffect(() => () => closeRef.current?.(), [])

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
  }, [events])

  function run(text: string) {
    const trimmed = text.trim()
    if (!trimmed || running) return

    closeRef.current?.()
    setEvents([])
    setRunning(true)
    closeRef.current = streamAgent(
      trimmed,
      (ev) => setEvents((prev) => [...prev, ev]),
      () => setRunning(false),
    )
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    run(goal)
  }

  const final = events.find((e) => e.type === 'final')
  const errored = events.find((e) => e.type === 'error')
  const steps = events.filter((e) => e.type !== 'final')
  const toolsUsed = new Set(events.filter((e) => e.type === 'tool_call').map((e) => e.tool))

  return (
    <section className="flex flex-col gap-5">
      <header className="border-b border-border pb-4">
        <h1 className="font-heading m-0 text-[clamp(1.9rem,3.2vw,2.45rem)] font-bold tracking-wide">
          Sourcing Agent
        </h1>
        <p className="mt-2 max-w-2xl text-[1.02rem] text-muted-foreground">
          Ask in plain language. The agent queries your own stock and sales, searches live
          listings, and works out a purchase plan — every step shown below.
        </p>
      </header>

      <form className="flex flex-col gap-3 sm:flex-row" onSubmit={onSubmit}>
        <Input
          value={goal}
          onChange={(e) => setGoal(e.target.value)}
          placeholder="e.g. 50 million rupiah, restock for three months"
          className="h-11 flex-1 rounded-none"
        />
        <Button type="submit" size="lg" disabled={running || !goal.trim()} className="rounded-none">
          {running ? 'Working…' : 'Ask agent'}
        </Button>
      </form>

      {events.length === 0 && !running && (
        <div className="flex flex-col gap-2 border-y border-dashed border-border py-4">
          <span className="text-xs uppercase tracking-wider text-muted-foreground">Try</span>
          {examples.map((ex) => (
            <button
              key={ex}
              type="button"
              className="bg-transparent text-left text-muted-foreground underline decoration-border underline-offset-4 hover:text-destructive"
              onClick={() => {
                setGoal(ex)
                run(ex)
              }}
            >
              {ex}
            </button>
          ))}
        </div>
      )}

      {steps.length > 0 && (
        <Card className="rounded-none shadow-none">
          <CardHeader className="pb-2">
            <CardTitle className="font-heading flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
              Agent trace
              {toolsUsed.size > 0 && (
                <Badge variant="outline" className="rounded-none">
                  {toolsUsed.size} tools used
                </Badge>
              )}
            </CardTitle>
            <Separator />
          </CardHeader>
          <CardContent>
            <ol className="m-0 flex list-none flex-col p-0">
              {steps.map((ev, i) => (
                <li key={i} className="flex gap-3 border-t border-border py-2.5 first:border-t-0">
                  <span
                    className={cn(
                      'mt-1.5 h-2 w-2 shrink-0 rounded-full',
                      ev.type === 'error' ? 'bg-destructive' : 'bg-secondary',
                    )}
                  />
                  <div className="min-w-0 flex-1">
                    <span className="font-heading block text-sm font-semibold">
                      {ev.type === 'tool_call'
                        ? (toolLabels[ev.tool ?? ''] ?? ev.tool)
                        : ev.type === 'tool_result'
                          ? `${ev.tool} returned`
                          : ev.type === 'error'
                            ? 'Error'
                            : ev.message}
                    </span>
                    {ev.type !== 'step' && (
                      <span className="mt-0.5 block break-words font-mono text-xs text-muted-foreground">
                        {ev.message}
                      </span>
                    )}
                  </div>
                </li>
              ))}
            </ol>
            <div ref={endRef} />
          </CardContent>
        </Card>
      )}

      {errored && !final && (
        <Alert variant="destructive">
          <AlertDescription>{errored.message}</AlertDescription>
        </Alert>
      )}

      {final && (
        <Card className="rounded-none border-destructive shadow-none">
          <CardHeader className="pb-2">
            <CardTitle className="font-heading text-xs font-semibold uppercase tracking-[0.12em] text-destructive">
              Recommendation
            </CardTitle>
            <Separator />
          </CardHeader>
          <CardContent>
            <p className="m-0 whitespace-pre-wrap text-[1.02rem] leading-relaxed">
              {final.message}
            </p>
          </CardContent>
        </Card>
      )}
    </section>
  )
}
