'use client'

import { useState } from 'react'
import type { Challenge } from '@/lib/api'
import { createChallenge, updateChallenge, deleteChallenge } from '@/lib/api'
import ChallengeForm from './ChallengeForm'

interface Props {
  initial: Challenge[]
}

type Modal = { mode: 'create' } | { mode: 'edit'; challenge: Challenge } | null

const difficultyColor: Record<string, string> = {
  easy: '#34c759',
  medium: '#ff9f0a',
  hard: '#ff3b30',
}

export default function ChallengeList({ initial }: Props) {
  const [list, setList] = useState(initial)
  const [modal, setModal] = useState<Modal>(null)

  async function handleCreate(data: { title: string; description: string; difficulty: string }) {
    const created = await createChallenge(data)
    setList(prev => [created, ...prev])
    setModal(null)
  }

  async function handleUpdate(id: string, data: { title: string; description: string; difficulty: string }) {
    const updated = await updateChallenge(id, data)
    setList(prev => prev.map(c => c.id === id ? updated : c))
    setModal(null)
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this challenge?')) return
    await deleteChallenge(id)
    setList(prev => prev.filter(c => c.id !== id))
  }

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h2 style={{ fontSize: 22, fontWeight: 700 }}>Challenges</h2>
        <button onClick={() => setModal({ mode: 'create' })} style={{
          background: '#5b5fc7', color: '#fff', border: 'none',
          padding: '10px 20px', borderRadius: 10, fontWeight: 600,
          cursor: 'pointer', fontSize: 14,
        }}>
          + New Challenge
        </button>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {list.length === 0 && (
          <p style={{ color: '#6e6e73', textAlign: 'center', padding: '40px 0' }}>No challenges yet.</p>
        )}
        {list.map(c => (
          <div key={c.id} style={{
            background: '#fff', borderRadius: 16, padding: '20px 24px',
            border: '1px solid #e5e5ea', display: 'flex', alignItems: 'center', gap: 16,
          }}>
            <div style={{ flex: 1 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 4 }}>
                <span style={{ fontWeight: 700, fontSize: 16 }}>{c.title}</span>
                <span style={{
                  fontSize: 11, fontWeight: 700, padding: '2px 8px', borderRadius: 6,
                  background: (difficultyColor[c.difficulty?.toLowerCase()] ?? '#8e8e93') + '20',
                  color: difficultyColor[c.difficulty?.toLowerCase()] ?? '#8e8e93',
                  textTransform: 'uppercase', letterSpacing: 0.5,
                }}>
                  {c.difficulty}
                </span>
              </div>
              <p style={{ color: '#6e6e73', fontSize: 14 }}>{c.description}</p>
              <p style={{ color: '#aeaeb2', fontSize: 12, marginTop: 6 }}>{c.submissions} submissions · ID: {c.id}</p>
            </div>
            <div style={{ display: 'flex', gap: 8 }}>
              <button onClick={() => setModal({ mode: 'edit', challenge: c })} style={{
                padding: '8px 16px', borderRadius: 8, border: '1px solid #e5e5ea',
                background: '#fff', cursor: 'pointer', fontSize: 13, fontWeight: 600,
              }}>
                Edit
              </button>
              <button onClick={() => handleDelete(c.id)} style={{
                padding: '8px 16px', borderRadius: 8, border: 'none',
                background: '#ff3b3010', color: '#ff3b30', cursor: 'pointer', fontSize: 13, fontWeight: 600,
              }}>
                Delete
              </button>
            </div>
          </div>
        ))}
      </div>

      {modal && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.4)',
          display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 100,
        }} onClick={() => setModal(null)}>
          <div style={{
            background: '#fff', borderRadius: 20, padding: 32, width: '100%',
            maxWidth: 480, margin: 24, boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
          }} onClick={e => e.stopPropagation()}>
            <h3 style={{ fontWeight: 700, fontSize: 18, marginBottom: 24 }}>
              {modal.mode === 'create' ? 'New Challenge' : 'Edit Challenge'}
            </h3>
            <ChallengeForm
              initial={modal.mode === 'edit' ? modal.challenge : undefined}
              onSave={data => modal.mode === 'create' ? handleCreate(data) : handleUpdate(modal.challenge.id, data)}
              onCancel={() => setModal(null)}
            />
          </div>
        </div>
      )}
    </>
  )
}
