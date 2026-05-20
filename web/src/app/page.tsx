import Link from 'next/link'
import { getStats } from '@/lib/api'

export default async function Home() {
  const stats = await getStats()

  return (
    <main style={{ maxWidth: 960, margin: '0 auto', padding: '48px 24px' }}>
      <div style={{ marginBottom: 40 }}>
        <h1 style={{ fontSize: 32, fontWeight: 800, letterSpacing: -0.5 }}>HumanLoop</h1>
        <p style={{ color: '#6e6e73', marginTop: 4 }}>Admin Dashboard</p>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(160px, 1fr))', gap: 16, marginBottom: 40 }}>
        {[
          { label: 'Total Submissions', value: stats?.total_submissions ?? '—' },
          { label: 'Verified', value: stats?.verified ?? '—' },
          { label: 'Synthetic Rejected', value: stats?.synthetic ?? '—' },
          { label: 'Credits Issued', value: stats?.credits_issued ?? '—' },
          { label: 'Challenges', value: stats?.challenges ?? '—' },
        ].map(s => (
          <div key={s.label} style={{ background: '#fff', borderRadius: 16, padding: '20px 24px', border: '1px solid #e5e5ea' }}>
            <div style={{ fontSize: 28, fontWeight: 800 }}>{s.value}</div>
            <div style={{ fontSize: 13, color: '#6e6e73', marginTop: 4 }}>{s.label}</div>
          </div>
        ))}
      </div>

      <div style={{ display: 'flex', gap: 12 }}>
        <Link href="/challenges" style={{
          background: '#5b5fc7', color: '#fff', padding: '12px 24px',
          borderRadius: 12, fontWeight: 600, fontSize: 15,
        }}>
          Manage Challenges
        </Link>
        <Link href="/stats" style={{
          background: '#fff', color: '#1d1d1f', padding: '12px 24px',
          borderRadius: 12, fontWeight: 600, fontSize: 15, border: '1px solid #e5e5ea',
        }}>
          View Stats
        </Link>
      </div>
    </main>
  )
}
