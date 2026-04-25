import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, RefreshCw, Trash2, ExternalLink } from 'lucide-react'
import { useState } from 'react'
import { api } from '../api/client'
import type { Feed } from '../types'
import { timeAgo } from '../utils'

export function FeedsPage() {
  const qc = useQueryClient()
  const [url, setUrl] = useState('')
  const [error, setError] = useState('')

  const { data: feeds = [], isLoading } = useQuery<Feed[]>({
    queryKey: ['feeds'],
    queryFn: api.feeds.list,
  })

  const addFeed = useMutation({
    mutationFn: (feedUrl: string) => api.feeds.create(feedUrl),
    onSuccess: () => {
      setUrl('')
      setError('')
      qc.invalidateQueries({ queryKey: ['feeds'] })
    },
    onError: (err: Error) => setError(err.message),
  })

  const deleteFeed = useMutation({
    mutationFn: api.feeds.delete,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['feeds'] })
      qc.invalidateQueries({ queryKey: ['items'] })
    },
  })

  const refreshFeed = useMutation({
    mutationFn: api.feeds.refresh,
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const trimmed = url.trim()
    if (!trimmed) return
    addFeed.mutate(trimmed)
  }

  return (
    <div className="max-w-3xl mx-auto px-4 py-6">
      <h1 className="text-xl font-semibold text-zinc-800 mb-6">Manage Feeds</h1>

      {/* Add feed form */}
      <div className="bg-white border border-zinc-200 rounded-xl p-5 mb-6">
        <h2 className="text-sm font-semibold text-zinc-700 mb-3">Add new feed</h2>
        <form onSubmit={handleSubmit} className="flex gap-2">
          <input
            type="url"
            placeholder="https://example.com/feed.xml"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            required
            className="flex-1 text-sm border border-zinc-200 rounded-lg px-3 py-2 bg-zinc-50 focus:outline-none focus:ring-2 focus:ring-zinc-300 focus:bg-white"
          />
          <button
            type="submit"
            disabled={addFeed.isPending}
            className="flex items-center gap-1.5 px-4 py-2 bg-zinc-900 text-white text-sm font-medium rounded-lg hover:bg-zinc-700 disabled:opacity-50 transition-colors"
          >
            <Plus size={15} />
            {addFeed.isPending ? 'Adding…' : 'Add'}
          </button>
        </form>
        {error && <p className="text-xs text-red-500 mt-2">{error}</p>}
        <p className="text-xs text-zinc-400 mt-2">
          Supports RSS 1.0/2.0, Atom, and JSON Feed formats.
        </p>
      </div>

      {/* Feed list */}
      {isLoading ? (
        <div className="text-center py-20 text-zinc-400 text-sm">Loading…</div>
      ) : feeds.length === 0 ? (
        <div className="text-center py-20 text-zinc-400 text-sm">
          No feeds yet. Add your first RSS feed above.
        </div>
      ) : (
        <ul className="space-y-3">
          {feeds.map((feed) => (
            <FeedRow
              key={feed.id}
              feed={feed}
              onRefresh={() => refreshFeed.mutate(feed.id)}
              onDelete={() => {
                if (window.confirm(`Delete "${feed.title || feed.url}"?`)) {
                  deleteFeed.mutate(feed.id)
                }
              }}
              refreshing={refreshFeed.isPending && refreshFeed.variables === feed.id}
            />
          ))}
        </ul>
      )}
    </div>
  )
}

function FeedRow({
  feed,
  onRefresh,
  onDelete,
  refreshing,
}: {
  feed: Feed
  onRefresh: () => void
  onDelete: () => void
  refreshing: boolean
}) {
  return (
    <li className="bg-white border border-zinc-200 rounded-xl p-4">
      <div className="flex items-start gap-3">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-zinc-800 truncate">
              {feed.title || <span className="text-zinc-400 italic">Untitled</span>}
            </span>
            <a
              href={feed.url}
              target="_blank"
              rel="noopener noreferrer"
              className="text-zinc-400 hover:text-zinc-600 shrink-0"
            >
              <ExternalLink size={12} />
            </a>
          </div>
          <p className="text-xs text-zinc-400 truncate mt-0.5">{feed.url}</p>
          <div className="flex items-center gap-3 mt-2 text-xs text-zinc-400">
            <span>{feed.item_count} items</span>
            <span>·</span>
            <span>
              {feed.last_fetched_at
                ? `Last fetched ${timeAgo(feed.last_fetched_at)}`
                : 'Never fetched'}
            </span>
            <span>·</span>
            <span>Every {feed.fetch_interval_minutes}m</span>
          </div>
        </div>

        <div className="flex items-center gap-1 shrink-0">
          <button
            onClick={onRefresh}
            disabled={refreshing}
            title="Refresh now"
            className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-700 hover:bg-zinc-100 disabled:opacity-40 transition-colors"
          >
            <RefreshCw size={14} className={refreshing ? 'animate-spin' : ''} />
          </button>
          <button
            onClick={onDelete}
            title="Delete feed"
            className="p-1.5 rounded-md text-zinc-400 hover:text-red-500 hover:bg-red-50 transition-colors"
          >
            <Trash2 size={14} />
          </button>
        </div>
      </div>
    </li>
  )
}
