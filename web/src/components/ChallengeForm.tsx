'use client'

import { useState } from 'react'
import type { Challenge } from '@/lib/api'

interface Props {
  initial?: Partial<Challenge>
  onSave: (data: { title: string; description: string; difficulty: string }) => Promise<void>
  onCancel: () => void
}

export default function ChallengeForm({ initial, onSave, onCancel }: Props) {
  const [title, setTitle] = useState(initial?.title ?? '')
  const [description, setDescription] = useState(initial?.description ?? '')
  const [difficulty, setDifficulty] = useState(initial?.difficulty ?? 'Easy')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    setError('')
    try {
      await onSave({ title, description, difficulty })
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to save')
      setSaving(false)
    }
  }

  const inputStyle: React.CSSProperties = {
    width: '100%', padding: '10px 14px', borderRadius: 10,
    border: '1px solid #e5e5ea', fontSize: 15, outline: 'none',
    fontFamily: 'inherit',
  }

  return (
    <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <div>
        <label style={{ fontSize: 13, fontWeight: 600, display: 'block', marginBottom: 6 }}>Title</label>
        <input style={inputStyle} value={title} onChange={e => setTitle(e.target.value)} required />
      </div>
      <div>
        <label style={{ fontSize: 13, fontWeight: 600, display: 'block', marginBottom: 6 }}>Description</label>
        <textarea
          style={{ ...inputStyle, minHeight: 80, resize: 'vertical' }}
          value={description}
          onChange={e => setDescription(e.target.value)}
          required
        />
      </div>
      <div>
        <label style={{ fontSize: 13, fontWeight: 600, display: 'block', marginBottom: 6 }}>Difficulty</label>
        <select style={inputStyle} value={difficulty} onChange={e => setDifficulty(e.target.value)}>
          <option>Easy</option>
          <option>Medium</option>
          <option>Hard</option>
        </select>
      </div>
      {error && <p style={{ color: '#ff3b30', fontSize: 13 }}>{error}</p>}
      <div style={{ display: 'flex', gap: 10, justifyContent: 'flex-end' }}>
        <button type="button" onClick={onCancel} style={{
          padding: '10px 20px', borderRadius: 10, border: '1px solid #e5e5ea',
          background: '#fff', fontWeight: 600, cursor: 'pointer', fontSize: 14,
        }}>
          Cancel
        </button>
        <button type="submit" disabled={saving} style={{
          padding: '10px 20px', borderRadius: 10, border: 'none',
          background: '#5b5fc7', color: '#fff', fontWeight: 600,
          cursor: saving ? 'not-allowed' : 'pointer', fontSize: 14, opacity: saving ? 0.7 : 1,
        }}>
          {saving ? 'Saving…' : 'Save'}
        </button>
      </div>
    </form>
  )
}
