import Link from 'next/link'
import { getStats } from '@/lib/api'

export default async function StatsPage() {
  const stats = await getStats()

  const rows = stats ? [
    { label: 'Total Submissions', value: stats.total_submissions },
    { label: 'Verified', value: stats.verified },
    { label: 'Synthetic Rejected', value: stats.synthetic },
    { label: 'Failed', value: stats.failed },
    { label: 'Credits Issued', value: stats.credits_issued },
    { label: 'Active Challenges', value: stats.challenges },
    { label: 'Acceptance Rate', value: stats.total_submissions > 0 ? `${Math.round(stats.verified / stats.total_submissions * 100)}%` : '—' },
  ] : []

  return (
    <main style={{ maxWidth: 960, margin: '0 auto', padding: '48px 24px' }}>
      <div style={{ marginBottom: 32 }}>
        <Link href="/" style={{ color: '#5b5fc7', fontSize: 14, fontWeight: 600 }}>← Dashboard</Link>
        <h1 style={{ fontSize: 32, fontWeight: 800, letterSpacing: -0.5, marginTop: 12 }}>Stats</h1>
      </div>

      {!stats ? (
        <p style={{ color: '#6e6e73' }}>Stats unavailable — backend may be offline or API_KEY not set.</p>
      ) : (
        <div style={{ background: '#fff', borderRadius: 20, border: '1px solid #e5e5ea', overflow: 'hidden' }}>
          {rows.map((row, i) => (
            <div key={row.label} style={{
              display: 'flex', justifyContent: 'space-between', alignItems: 'center',
              padding: '18px 24px',
              borderTop: i === 0 ? 'none' : '1px solid #f2f2f7',
            }}>
              <span style={{ fontSize: 15, color: '#3a3a3c' }}>{row.label}</span>
              <span style={{ fontSize: 18, fontWeight: 700 }}>{row.value}</span>
            </div>
          ))}
        </div>
      )}
    </main>
  )
}
