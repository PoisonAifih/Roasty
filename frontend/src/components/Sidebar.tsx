import { NavLink } from 'react-router-dom'
import { useEffect, useState, type ReactNode } from 'react'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'
import { api } from '@/api/client'

const links = [
  { to: '/', label: 'Dashboard', end: true },
  { to: '/agent', label: 'Sourcing Agent', end: false },
  { to: '/scout', label: 'Bean Scout', end: false },
  { to: '/inventory', label: 'Smart Inventory', end: false },
  { to: '/beans', label: 'Bean Management', end: false },
  { to: '/crm', label: 'Customer & Payment', end: false },
  { to: '/notifications', label: 'Notifications', end: false },
]

export function Sidebar({ children }: { children: ReactNode }) {
  const [unread, setUnread] = useState(0)

  useEffect(() => {
    let cancelled = false
    async function poll() {
      try {
        const data = await api.notifications()
        if (!cancelled) setUnread(data.unread)
      } catch {
        // non-fatal
      }
    }
    void poll()
    const id = setInterval(() => void poll(), 60_000)
    return () => {
      cancelled = true
      clearInterval(id)
    }
  }, [])

  return (
    <div className="grid min-h-screen grid-cols-1 md:grid-cols-[250px_1fr]">
      <aside className="flex flex-col gap-7 border-b border-sidebar-border bg-sidebar p-5 text-sidebar-foreground md:border-b-0 md:border-r">
        <div className="flex items-center gap-3 px-1 pb-4">
          <span
            className="h-8 w-8 shrink-0 rounded-full bg-sidebar-primary shadow-[inset_0_0_0_2px_rgba(248,244,225,0.35)]"
            aria-hidden
          />
          <div>
            <strong className="font-heading block text-[1.55rem] font-bold leading-tight tracking-wide">
              Roasty
            </strong>
            <small className="block text-[0.78rem] tracking-wide">Roastery Scout AI</small>
          </div>
        </div>
        <Separator className="bg-sidebar-border" />

        <nav className="flex flex-1 flex-row flex-wrap gap-1 md:flex-col">
          {links.map((l) => (
            <NavLink key={l.to} to={l.to} end={l.end}>
              {({ isActive }) => (
                <Button
                  variant="ghost"
                  className={cn(
                    'font-heading h-auto w-full justify-start rounded-none px-3 py-2.5 text-[0.98rem] font-bold',
                    isActive
                      ? 'border-b-2 border-sidebar-primary md:border-b-0 md:border-l-2'
                      : 'border-b-2 border-transparent md:border-b-0 md:border-l-2',
                  )}
                >
                  <span className="flex w-full items-center justify-between">
                    {l.label}
                    {l.to === '/notifications' && unread > 0 && (
                      <span className="ml-2 inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-destructive px-1 text-[0.7rem] font-bold text-white">
                        {unread}
                      </span>
                    )}
                  </span>
                </Button>
              )}
            </NavLink>
          ))}
        </nav>

        <div className="hidden border-t border-sidebar-border px-1 pt-4 md:block">
          <span className="font-heading block text-[0.95rem] font-semibold">Hello !</span>
        </div>
      </aside>

      <div className="min-w-0">
        <main className="mx-auto max-w-[1080px] px-4 py-5 md:px-9 md:py-8">{children}</main>
      </div>
    </div>
  )
}
