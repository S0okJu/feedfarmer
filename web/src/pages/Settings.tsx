import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { CheckCircle, Circle, Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { api } from '../api/client'
import type { AIConfig } from '../types'

const PROVIDERS = ['ollama', 'openai', 'anthropic'] as const

export function Settings() {
  const qc = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({
    name: '',
    provider: 'ollama',
    base_url: '',
    model: '',
    is_active: true,
  })

  const { data: configs = [] } = useQuery<AIConfig[]>({
    queryKey: ['ai-configs'],
    queryFn: api.aiConfigs.list,
  })

  const createConfig = useMutation({
    mutationFn: () => api.aiConfigs.create(form),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['ai-configs'] })
      setShowForm(false)
      setForm({ name: '', provider: 'ollama', base_url: '', model: '', is_active: true })
    },
  })

  const deleteConfig = useMutation({
    mutationFn: (id: string) => api.aiConfigs.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['ai-configs'] }),
  })

  const activateConfig = useMutation({
    mutationFn: (id: string) => api.aiConfigs.activate(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['ai-configs'] }),
  })

  return (
    <div className="max-w-2xl mx-auto px-4 py-6">
      <h1 className="text-xl font-semibold text-zinc-800 mb-6">Settings</h1>

      {/* AI Providers */}
      <section>
        <div className="flex items-center justify-between mb-3">
          <div>
            <h2 className="text-sm font-semibold text-zinc-700">AI Providers</h2>
            <p className="text-xs text-zinc-400 mt-0.5">
              Configure Ollama, OpenAI, or Anthropic for summarization and tagging.
            </p>
          </div>
          <button
            onClick={() => setShowForm((v) => !v)}
            className="flex items-center gap-1.5 text-sm bg-zinc-900 text-white px-3 py-1.5 rounded-lg hover:bg-zinc-700 transition-colors"
          >
            <Plus size={14} />
            Add
          </button>
        </div>

        {/* Add form */}
        {showForm && (
          <div className="bg-white border border-zinc-200 rounded-xl p-4 mb-3 space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-medium text-zinc-600 mb-1">Name</label>
                <input
                  type="text"
                  placeholder="My Ollama"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  className="w-full text-sm border border-zinc-200 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-zinc-300"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-zinc-600 mb-1">Provider</label>
                <select
                  value={form.provider}
                  onChange={(e) => setForm({ ...form, provider: e.target.value })}
                  className="w-full text-sm border border-zinc-200 rounded-lg px-3 py-2 bg-white focus:outline-none focus:ring-2 focus:ring-zinc-300"
                >
                  {PROVIDERS.map((p) => (
                    <option key={p} value={p}>{p}</option>
                  ))}
                </select>
              </div>
            </div>
            <div>
              <label className="block text-xs font-medium text-zinc-600 mb-1">Base URL</label>
              <input
                type="text"
                placeholder="http://localhost:11434"
                value={form.base_url}
                onChange={(e) => setForm({ ...form, base_url: e.target.value })}
                className="w-full text-sm border border-zinc-200 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-zinc-300"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-zinc-600 mb-1">Model</label>
              <input
                type="text"
                placeholder="gemma3:4b"
                value={form.model}
                onChange={(e) => setForm({ ...form, model: e.target.value })}
                className="w-full text-sm border border-zinc-200 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-zinc-300"
              />
            </div>
            <div className="flex items-center gap-2">
              <input
                id="is_active"
                type="checkbox"
                checked={form.is_active}
                onChange={(e) => setForm({ ...form, is_active: e.target.checked })}
                className="rounded border-zinc-300"
              />
              <label htmlFor="is_active" className="text-sm text-zinc-600">Set as active provider</label>
            </div>
            <div className="flex gap-2 pt-1">
              <button
                onClick={() => createConfig.mutate()}
                disabled={!form.base_url || createConfig.isPending}
                className="text-sm bg-zinc-900 text-white px-4 py-1.5 rounded-lg hover:bg-zinc-700 disabled:opacity-50 transition-colors"
              >
                {createConfig.isPending ? 'Saving…' : 'Save'}
              </button>
              <button
                onClick={() => setShowForm(false)}
                className="text-sm text-zinc-500 hover:text-zinc-800 px-3 py-1.5 transition-colors"
              >
                Cancel
              </button>
            </div>
            {createConfig.isError && (
              <p className="text-xs text-red-500">{(createConfig.error as Error)?.message}</p>
            )}
          </div>
        )}

        {/* Config list */}
        {configs.length === 0 && !showForm ? (
          <div className="text-center py-10 text-zinc-400 text-sm border border-dashed border-zinc-200 rounded-xl">
            No AI provider configured yet.
          </div>
        ) : (
          <ul className="space-y-2">
            {configs.map((cfg) => (
              <li
                key={cfg.id}
                className="bg-white border border-zinc-200 rounded-xl px-4 py-3 flex items-center gap-3"
              >
                {/* Active indicator */}
                <button
                  onClick={() => activateConfig.mutate(cfg.id)}
                  title={cfg.is_active ? 'Active' : 'Set as active'}
                  className={cfg.is_active ? 'text-emerald-500' : 'text-zinc-300 hover:text-zinc-500'}
                >
                  {cfg.is_active ? <CheckCircle size={18} /> : <Circle size={18} />}
                </button>

                {/* Info */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-zinc-800">
                      {cfg.name || cfg.provider}
                    </span>
                    <span className="text-xs px-1.5 py-0.5 bg-zinc-100 text-zinc-500 rounded">
                      {cfg.provider}
                    </span>
                    {cfg.is_active && (
                      <span className="text-xs px-1.5 py-0.5 bg-emerald-50 text-emerald-600 rounded font-medium">
                        active
                      </span>
                    )}
                  </div>
                  <div className="text-xs text-zinc-400 mt-0.5 truncate">
                    {cfg.base_url} · {cfg.model}
                  </div>
                </div>

                {/* Delete */}
                <button
                  onClick={() => deleteConfig.mutate(cfg.id)}
                  title="Delete"
                  className="text-zinc-300 hover:text-red-500 transition-colors"
                >
                  <Trash2 size={15} />
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  )
}
