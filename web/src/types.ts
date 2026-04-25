export interface Feed {
  id: string
  url: string
  title: string
  description: string
  last_fetched_at: string | null
  fetch_interval_minutes: number
  item_count: number
  created_at: string
}

export interface Item {
  id: string
  feed_id: string
  feed_title: string
  title: string
  link: string
  content: string
  published_at: string | null
  ai_summary: string
  ai_tags: string[]
  ai_score: number
  is_read: boolean
  is_bookmarked: boolean
  created_at: string
}

export interface AIConfig {
  id: string
  name: string
  provider: string
  base_url: string
  model: string
  is_active: boolean
  created_at: string
}

export interface ItemsParams {
  feedId?: string
  unread?: boolean
  bookmarked?: boolean
  q?: string
  limit?: number
  offset?: number
}
