import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { AlertCircle, Bookmark, BookmarkCheck, Eye, EyeOff, ExternalLink, Loader2, RefreshCw, Search, Sparkles } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { api } from '../api/client'
import type { Feed, Item } from '../types'
import { stripHTML, timeAgo } from '../utils'

type Filter = 'all' | 'unread' | 'bookmarked'

export function Dashboard() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [filter, setFilter] = useState<Filter>('all')
  const [search, setSearch] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')

  // Feed id from URL query param (?feed=xxx)
  const feedId = searchParams.get('feed') ?? ''

  useEffect(() => {
    const id = setTimeout(() => setDebouncedSearch(search), 300)
    return () => clearTimeout(id)
  }, [search])

  const { data: feeds = [] } = useQuery<Feed[]>({
    queryKey: ['feeds'],
    queryFn: api.feeds.list,
  })

  const {
    data: items = [],
    isLoading,
    refetch,
  } = useQuery<Item[]>({
    queryKey: ['items', filter, feedId, debouncedSearch],
    queryFn: () =>
      api.items.list({
        feedId: feedId || undefined,
        unread: filter === 'unread',
        bookmarked: filter === 'bookmarked',
        q: debouncedSearch || undefined,
        limit: 50,
      }),
  })

  const qc = useQueryClient()

  const updateItem = useMutation({
    mutationFn: ({ id, data }: { id: string; data: { is_read?: boolean; is_bookmarked?: boolean } }) =>
      api.items.update(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['items'] }),
  })

  const selectedFeed = feeds.find((f) => f.id === feedId)

  return (
    <div className="max-w-3xl mx-auto px-4 py-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-5">
        <h1 className="text-xl font-semibold text-zinc-800">
          {selectedFeed ? selectedFeed.title || selectedFeed.url : 'All Items'}
        </h1>
        <button
          onClick={() => refetch()}
          className="flex items-center gap-1.5 text-sm text-zinc-500 hover:text-zinc-800 transition-colors"
        >
          <RefreshCw size={14} />
          Refresh
        </button>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 mb-5">
        {/* Tab filter */}
        <div className="flex items-center gap-1 bg-white border border-zinc-200 rounded-lg p-1">
          {(['all', 'unread', 'bookmarked'] as Filter[]).map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={[
                'px-3 py-1.5 rounded-md text-sm font-medium capitalize transition-colors',
                filter === f
                  ? 'bg-zinc-900 text-white'
                  : 'text-zinc-500 hover:text-zinc-800',
              ].join(' ')}
            >
              {f}
            </button>
          ))}
        </div>

        {/* Feed selector */}
        {feeds.length > 0 && (
          <select
            value={feedId}
            onChange={(e) => {
              if (e.target.value) {
                setSearchParams({ feed: e.target.value })
              } else {
                setSearchParams({})
              }
            }}
            className="text-sm border border-zinc-200 rounded-lg px-3 py-2 bg-white text-zinc-700 focus:outline-none focus:ring-2 focus:ring-zinc-300"
          >
            <option value="">All feeds</option>
            {feeds.map((f) => (
              <option key={f.id} value={f.id}>
                {f.title || f.url}
              </option>
            ))}
          </select>
        )}

        {/* Search */}
        <div className="relative flex-1">
          <Search size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-400" />
          <input
            type="text"
            placeholder="Search articles…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-3 py-2 text-sm border border-zinc-200 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-zinc-300"
          />
        </div>
      </div>

      {/* Items */}
      {isLoading ? (
        <div className="text-center py-20 text-zinc-400 text-sm">Loading…</div>
      ) : items.length === 0 ? (
        <div className="text-center py-20 text-zinc-400 text-sm">
          No items found.{' '}
          {feeds.length === 0 && (
            <a href="/feeds" className="text-blue-500 hover:underline">
              Add your first feed →
            </a>
          )}
        </div>
      ) : (
        <ul className="space-y-3">
          {items.map((item) => (
            <ItemCard
              key={item.id}
              item={item}
              onToggleRead={() =>
                updateItem.mutate({ id: item.id, data: { is_read: !item.is_read } })
              }
              onToggleBookmark={() =>
                updateItem.mutate({ id: item.id, data: { is_bookmarked: !item.is_bookmarked } })
              }
            />
          ))}
        </ul>
      )}
    </div>
  )
}

function ItemCard({
  item,
  onToggleRead,
  onToggleBookmark,
}: {
  item: Item
  onToggleRead: () => void
  onToggleBookmark: () => void
}) {
  const qc = useQueryClient()
  const [showSummary, setShowSummary] = useState(false)

  const summarize = useMutation({
    mutationFn: () => api.items.summarize(item.id),
    onSuccess: (updated) => {
      qc.setQueriesData<Item[]>({ queryKey: ['items'] }, (prev) =>
        prev?.map((it) => (it.id === item.id ? { ...it, ai_summary: updated.ai_summary } : it))
      )
      qc.invalidateQueries({ queryKey: ['items'] })
      setShowSummary(true)
    },
  })

  const excerpt = stripHTML(item.content).slice(0, 200)

  return (
    <li
      className={[
        'bg-white border border-zinc-200 rounded-xl p-4 transition-opacity',
        item.is_read ? 'opacity-55' : '',
      ].join(' ')}
    >
      <div className="flex items-start gap-3">
        {/* Content */}
        <div className="flex-1 min-w-0">
          {/* Meta */}
          <div className="flex items-center gap-2 text-xs text-zinc-400 mb-1">
            <span className="font-medium text-blue-600 truncate max-w-[140px]">
              {item.feed_title}
            </span>
            <span>·</span>
            <span className="shrink-0">{timeAgo(item.published_at || item.created_at)}</span>
          </div>

          {/* Title */}
          <a
            href={item.link}
            target="_blank"
            rel="noopener noreferrer"
            className="group flex items-start gap-1 text-sm font-semibold text-zinc-900 hover:text-blue-600 leading-snug"
            onClick={() => {
              if (!item.is_read) onToggleRead()
            }}
          >
            <span className="line-clamp-2">{item.title}</span>
            <ExternalLink
              size={12}
              className="shrink-0 mt-0.5 opacity-0 group-hover:opacity-100 transition-opacity"
            />
          </a>

          {/* Excerpt */}
          {excerpt && (
            <p className="text-xs text-zinc-500 mt-1 line-clamp-2 leading-relaxed">
              {excerpt}
            </p>
          )}

          {/* AI Summary */}
          <div className="mt-2 space-y-1">
            <button
              onClick={() => {
                if (item.ai_summary) {
                  setShowSummary((v) => !v)
                } else if (!summarize.isPending) {
                  summarize.mutate()
                }
              }}
              className={[
                'flex items-center gap-1.5 text-xs transition-colors disabled:opacity-50',
                item.ai_summary
                  ? 'text-indigo-500 hover:text-indigo-700'
                  : 'text-zinc-400 hover:text-indigo-600',
              ].join(' ')}
            >
              {summarize.isPending ? (
                <Loader2 size={12} className="animate-spin" />
              ) : (
                <Sparkles size={12} />
              )}
              {summarize.isPending
                ? '요약 중…'
                : item.ai_summary
                  ? showSummary ? 'AI 요약 닫기' : 'AI 요약 보기'
                  : 'AI 요약'}
            </button>
            {summarize.isError && (
              <div className="flex items-center gap-1 text-xs text-red-500">
                <AlertCircle size={11} />
                {(summarize.error as Error)?.message || 'AI 요약 실패'}
              </div>
            )}
            {item.ai_summary && showSummary && (
              <div className="px-3 py-2 bg-indigo-50 border border-indigo-100 rounded-lg text-xs text-indigo-700 leading-relaxed">
                {item.ai_summary}
              </div>
            )}
          </div>

        </div>

        {/* Actions */}
        <div className="flex flex-col gap-1.5 shrink-0">
          <button
            onClick={onToggleBookmark}
            title={item.is_bookmarked ? 'Remove bookmark' : 'Bookmark'}
            className={[
              'p-1.5 rounded-md transition-colors',
              item.is_bookmarked
                ? 'text-amber-500 hover:bg-amber-50'
                : 'text-zinc-300 hover:text-zinc-500 hover:bg-zinc-50',
            ].join(' ')}
          >
            {item.is_bookmarked ? <BookmarkCheck size={15} /> : <Bookmark size={15} />}
          </button>
          <button
            onClick={onToggleRead}
            title={item.is_read ? 'Mark unread' : 'Mark read'}
            className={[
              'p-1.5 rounded-md transition-colors',
              item.is_read
                ? 'text-zinc-300 hover:text-zinc-500 hover:bg-zinc-50'
                : 'text-zinc-400 hover:text-zinc-600 hover:bg-zinc-50',
            ].join(' ')}
          >
            {item.is_read ? <EyeOff size={15} /> : <Eye size={15} />}
          </button>
        </div>
      </div>
    </li>
  )
}
