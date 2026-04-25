import type { Feed, Item, ItemsParams } from '../types'

async function req<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) {
    const msg = await res.text().catch(() => res.statusText)
    throw new Error(msg || `HTTP ${res.status}`)
  }
  if (res.status === 204) return null as T
  return res.json() as Promise<T>
}

export const api = {
  feeds: {
    list: () => req<Feed[]>('/api/feeds'),

    create: (url: string, intervalMinutes = 60) =>
      req<Feed>('/api/feeds', {
        method: 'POST',
        body: JSON.stringify({ url, fetch_interval_minutes: intervalMinutes }),
      }),

    delete: (id: string) =>
      req<null>(`/api/feeds/${id}`, { method: 'DELETE' }),

    refresh: (id: string) =>
      req<{ status: string }>(`/api/feeds/${id}/refresh`, { method: 'POST' }),
  },

  items: {
    list: (params: ItemsParams = {}) => {
      const qs = new URLSearchParams()
      if (params.feedId)    qs.set('feed_id',   params.feedId)
      if (params.unread)    qs.set('unread',     'true')
      if (params.bookmarked) qs.set('bookmarked', 'true')
      if (params.q)         qs.set('q',          params.q)
      if (params.limit)     qs.set('limit',      String(params.limit))
      if (params.offset)    qs.set('offset',     String(params.offset))
      return req<Item[]>(`/api/items?${qs}`)
    },

    get: (id: string) => req<Item>(`/api/items/${id}`),

    update: (id: string, data: { is_read?: boolean; is_bookmarked?: boolean }) =>
      req<Item>(`/api/items/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(data),
      }),
  },
}
