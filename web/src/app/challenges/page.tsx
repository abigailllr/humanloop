import Link from 'next/link'
import { getChallenges } from '@/lib/api'
import ChallengeList from '@/components/ChallengeList'

export default async function ChallengesPage() {
  const challenges = await getChallenges()

  return (
    <main style={{ maxWidth: 960, margin: '0 auto', padding: '48px 24px' }}>
      <div style={{ marginBottom: 32 }}>
        <Link href="/" style={{ color: '#5b5fc7', fontSize: 14, fontWeight: 600 }}>← Dashboard</Link>
        <h1 style={{ fontSize: 32, fontWeight: 800, letterSpacing: -0.5, marginTop: 12 }}>Challenges</h1>
      </div>
      <ChallengeList initial={challenges} />
    </main>
  )
}
