import { useQuery } from '@tanstack/react-query'
import { Layers, Rss } from 'lucide-react'
import { NavLink, Outlet } from 'react-router-dom'
import { api } from '../api/client'
import type { Feed } from '../types'

function navClass({ isActive }: { isActive: boolean }) {
  return [
    'flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium transition-colors',
    isActive
      ? 'bg-zinc-700 text-white'
      : 'text-zinc-400 hover:bg-zinc-800 hover:text-white',
  ].join(' ')
}

export function Layout() {
  const { data: feeds = [] } = useQuery<Feed[]>({
    queryKey: ['feeds'],
    queryFn: api.feeds.list,
  })

  return (
    <div className="flex h-screen overflow-hidden">
      {/* Sidebar */}
      <aside className="w-56 shrink-0 bg-zinc-900 flex flex-col border-r border-zinc-800">
        {/* Brand */}
        <div className="px-5 py-4 border-b border-zinc-800">
          <span className="text-white font-bold text-base tracking-tight">
            🌾 FeedFarmer
          </span>
        </div>

        {/* Nav */}
        <nav className="p-3 space-y-0.5">
          <NavLink to="/" end className={navClass}>
            <Layers size={16} />
            All Items
          </NavLink>
          <NavLink to="/feeds" className={navClass}>
            <Rss size={16} />
            Manage Feeds
          </NavLink>
        </nav>

        {/* Feed list */}
        {feeds.length > 0 && (
          <div className="px-3 pt-2">
            <p className="text-xs font-semibold text-zinc-500 uppercase px-2 mb-1">
              Feeds
            </p>
            <nav className="space-y-0.5 overflow-y-auto max-h-80">
              {feeds.map((f) => (
                <NavLink
                  key={f.id}
                  to={`/?feed=${f.id}`}
                  className="flex items-center justify-between px-2 py-1.5 rounded text-xs text-zinc-400 hover:bg-zinc-800 hover:text-white transition-colors"
                >
                  <span className="truncate">{f.title || f.url}</span>
                  <span className="ml-1 text-zinc-600 shrink-0">{f.item_count}</span>
                </NavLink>
              ))}
            </nav>
          </div>
        )}

        {/* Footer */}
        <div className="mt-auto p-4 border-t border-zinc-800">
          <p className="text-xs text-zinc-600">
            {feeds.length} feed{feeds.length !== 1 ? 's' : ''} subscribed
          </p>
        </div>
      </aside>

      {/* Main */}
      <main className="flex-1 overflow-y-auto">
        <Outlet />
      </main>
    </div>
  )
}
