const wrap: React.CSSProperties = {
  position: 'sticky',
  top: 0,
  zIndex: 10,
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'space-between',
  padding: '18px 24px',
  maxWidth: 1100,
  margin: '0 auto',
  width: '100%',
}

const brand: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: 10,
  fontWeight: 800,
  fontSize: 19,
  letterSpacing: -0.3,
}

const dot: React.CSSProperties = {
  width: 26,
  height: 26,
  borderRadius: 8,
  background: '#5b5fc7',
  display: 'inline-block',
}

const cta: React.CSSProperties = {
  background: '#5b5fc7',
  color: '#fff',
  padding: '10px 18px',
  borderRadius: 999,
  fontWeight: 600,
  fontSize: 14,
}

export default function Nav() {
  return (
    <header style={{ borderBottom: '1px solid #f0f0f2', background: 'rgba(255,255,255,0.85)', backdropFilter: 'blur(8px)' }}>
      <nav style={wrap}>
        <span style={brand}><span style={dot} />HumanLoop</span>
        <a href="#download" style={cta}>Download</a>
      </nav>
    </header>
  )
}
