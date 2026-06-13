const steps = [
  { n: '1', title: 'Pick a challenge', body: 'Browse the feed and choose a physical task to record.' },
  { n: '2', title: 'Film it', body: 'Capture a short clip with your phone. Submit it in the app.' },
  { n: '3', title: 'Earn & repeat', body: 'Get credits for accepted clips and climb the leaderboard.' },
]

const section: React.CSSProperties = {
  background: '#0a0a0a',
  color: '#fff',
}

const inner: React.CSSProperties = {
  maxWidth: 1100,
  margin: '0 auto',
  padding: '72px 24px',
}

const heading: React.CSSProperties = { fontSize: 34, fontWeight: 800, letterSpacing: -0.8, marginBottom: 40 }

const grid: React.CSSProperties = {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))',
  gap: 28,
}

const num: React.CSSProperties = {
  width: 44,
  height: 44,
  borderRadius: 12,
  background: '#5b5fc7',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  fontWeight: 800,
  fontSize: 18,
  marginBottom: 16,
}

const stepTitle: React.CSSProperties = { fontSize: 19, fontWeight: 700, marginBottom: 8 }
const stepBody: React.CSSProperties = { color: 'rgba(255,255,255,0.66)', fontSize: 15, lineHeight: 1.55 }

export default function Steps() {
  return (
    <section style={section} id="how">
      <div style={inner}>
        <h2 style={heading}>How it works</h2>
        <div style={grid}>
          {steps.map((s) => (
            <div key={s.n}>
              <div style={num}>{s.n}</div>
              <div style={stepTitle}>{s.title}</div>
              <div style={stepBody}>{s.body}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
