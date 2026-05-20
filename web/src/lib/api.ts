const base = process.env.API_URL!
const key = process.env.API_KEY!

const headers = { 'X-API-Key': key, 'Content-Type': 'application/json' }

export interface Challenge {
  id: string
  title: string
  description: string
  difficulty: string
  submissions: number
}

export interface Stats {
  total_submissions: number
  verified: number
  synthetic: number
  failed: number
  credits_issued: number
  challenges: number
}

export async function getChallenges(): Promise<Challenge[]> {
  const res = await fetch(`${base}/v1/challenges`, { cache: 'no-store' })
  if (!res.ok) return []
  return res.json()
}

export async function createChallenge(body: Omit<Challenge, 'id' | 'submissions'>): Promise<Challenge> {
  const res = await fetch(`${base}/v1/challenges`, { method: 'POST', headers, body: JSON.stringify(body) })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}

export async function updateChallenge(id: string, body: Partial<Omit<Challenge, 'id' | 'submissions'>>): Promise<Challenge> {
  const res = await fetch(`${base}/v1/challenges/${id}`, { method: 'PUT', headers, body: JSON.stringify(body) })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}

export async function deleteChallenge(id: string): Promise<void> {
  const res = await fetch(`${base}/v1/challenges/${id}`, { method: 'DELETE', headers })
  if (!res.ok) throw new Error(await res.text())
}

export async function getStats(): Promise<Stats | null> {
  const res = await fetch(`${base}/v1/data/stats`, { headers, cache: 'no-store' })
  if (!res.ok) return null
  return res.json()
}
