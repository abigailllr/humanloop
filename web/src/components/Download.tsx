const section: React.CSSProperties = {
  maxWidth: 1100,
  margin: '0 auto',
  padding: '88px 24px',
  textAlign: 'center',
}

const heading: React.CSSProperties = {
  fontSize: 40,
  fontWeight: 850,
  letterSpacing: -1,
  marginBottom: 14,
}

const sub: React.CSSProperties = { color: '#52525b', fontSize: 18, marginBottom: 32 }

const row: React.CSSProperties = { display: 'flex', gap: 12, justifyContent: 'center', flexWrap: 'wrap' }

const btn: React.CSSProperties = {
  display: 'inline-flex',
  alignItems: 'center',
  gap: 10,
  background: '#0a0a0a',
  color: '#fff',
  padding: '14px 22px',
  borderRadius: 14,
  fontWeight: 600,
  fontSize: 15,
}

export default function Download() {
  return (
    <section style={section} id="download">
      <h2 style={heading}>Get the app</h2>
      <p style={sub}>Available on iPhone and Android. Free to download.</p>
      <div style={row}>
        <a href="#" style={btn}>Download on the App Store</a>
        <a href="#" style={{ ...btn, background: '#5b5fc7' }}>Get it on Google Play</a>
      </div>
    </section>
  )
}
