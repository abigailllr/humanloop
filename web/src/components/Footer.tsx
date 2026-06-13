const footer: React.CSSProperties = {
  borderTop: '1px solid #f0f0f2',
  background: '#fafafb',
}

const inner: React.CSSProperties = {
  maxWidth: 1100,
  margin: '0 auto',
  padding: '40px 24px',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'space-between',
  flexWrap: 'wrap',
  gap: 16,
}

const links: React.CSSProperties = { display: 'flex', gap: 24, fontSize: 14, color: '#52525b' }
const note: React.CSSProperties = { fontSize: 13, color: '#9ca3af' }

export default function Footer() {
  return (
    <footer style={footer}>
      <div style={inner}>
        <div>
          <div style={{ fontWeight: 800, fontSize: 16, marginBottom: 4 }}>HumanLoop</div>
          <div style={note}>HumanLoop is a mobile app for iOS and Android.</div>
        </div>
        <nav style={links}>
          <a href="#">Privacy</a>
          <a href="#">Terms</a>
          <a href="mailto:support@humanloop.app">Support</a>
        </nav>
      </div>
    </footer>
  )
}
