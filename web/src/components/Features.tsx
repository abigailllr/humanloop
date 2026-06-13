const items = [
  { title: 'Film short challenges', body: 'Pick a task like sorting, folding, or pick-and-place and record it in seconds with your phone camera.' },
  { title: 'Earn credits', body: 'Every accepted clip earns credits. Climb the leaderboard and unlock badges as you contribute.' },
  { title: 'Train real robots', body: 'Your clips become structured training data that teaches robots real-world manipulation skills.' },
  { title: 'Privacy first', body: 'You choose what to record and can delete your account and data from the app at any time.' },
]

const section: React.CSSProperties = {
  maxWidth: 1100,
  margin: '0 auto',
  padding: '64px 24px',
}

const heading: React.CSSProperties = {
  fontSize: 34,
  fontWeight: 800,
  letterSpacing: -0.8,
  marginBottom: 8,
}

const lede: React.CSSProperties = { color: '#52525b', fontSize: 17, marginBottom: 40 }

const grid: React.CSSProperties = {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fit, minmax(230px, 1fr))',
  gap: 20,
}

const card: React.CSSProperties = {
  background: '#fafafb',
  border: '1px solid #f0f0f2',
  borderRadius: 18,
  padding: 24,
}

const cardTitle: React.CSSProperties = { fontSize: 18, fontWeight: 700, marginBottom: 8 }
const cardBody: React.CSSProperties = { color: '#52525b', fontSize: 15, lineHeight: 1.55 }

export default function Features() {
  return (
    <section style={section} id="features">
      <h2 style={heading}>Built for contributors</h2>
      <p style={lede}>A simple loop: film, earn, repeat — while helping robots learn.</p>
      <div style={grid}>
        {items.map((it) => (
          <div key={it.title} style={card}>
            <div style={cardTitle}>{it.title}</div>
            <div style={cardBody}>{it.body}</div>
          </div>
        ))}
      </div>
    </section>
  )
}
