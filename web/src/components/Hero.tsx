const section: React.CSSProperties = {
  maxWidth: 1100,
  margin: '0 auto',
  padding: '80px 24px 64px',
  display: 'grid',
  gridTemplateColumns: '1.1fr 0.9fr',
  gap: 48,
  alignItems: 'center',
}

const eyebrow: React.CSSProperties = {
  display: 'inline-block',
  fontSize: 13,
  fontWeight: 700,
  color: '#5b5fc7',
  background: '#eef0ff',
  padding: '6px 12px',
  borderRadius: 999,
  marginBottom: 20,
}

const h1: React.CSSProperties = {
  fontSize: 52,
  lineHeight: 1.05,
  letterSpacing: -1.5,
  fontWeight: 850,
  marginBottom: 20,
}

const sub: React.CSSProperties = {
  fontSize: 19,
  lineHeight: 1.5,
  color: '#52525b',
  maxWidth: 480,
  marginBottom: 32,
}

const storeRow: React.CSSProperties = { display: 'flex', gap: 12, flexWrap: 'wrap' }

const storeBtn: React.CSSProperties = {
  display: 'inline-flex',
  alignItems: 'center',
  gap: 10,
  background: '#0a0a0a',
  color: '#fff',
  padding: '13px 20px',
  borderRadius: 14,
  fontWeight: 600,
  fontSize: 15,
}

const phone: React.CSSProperties = {
  margin: '0 auto',
  width: 260,
  height: 520,
  borderRadius: 40,
  background: 'linear-gradient(160deg, #5b5fc7, #8f7bff)',
  border: '10px solid #0a0a0a',
  boxShadow: '0 30px 60px rgba(91,95,199,0.28)',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  color: 'rgba(255,255,255,0.9)',
  fontWeight: 700,
  fontSize: 18,
}

export default function Hero() {
  return (
    <section style={section}>
      <div>
        <span style={eyebrow}>Crowdsourced robot training data</span>
        <h1 style={h1}>Film challenges. Train robots.</h1>
        <p style={sub}>
          HumanLoop turns everyday moments into the data that teaches robots how to act in the real
          world. Record short physical challenges, earn credits, and help build smarter machines.
        </p>
        <div style={storeRow}>
          <a href="#" style={storeBtn}>Download on the App Store</a>
          <a href="#" style={{ ...storeBtn, background: '#5b5fc7' }}>Get it on Google Play</a>
        </div>
      </div>
      <div style={phone}>App preview</div>
    </section>
  )
}
