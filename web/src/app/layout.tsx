import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'HumanLoop — Film challenges, train robots',
  description: 'HumanLoop is a mobile app where everyday people film short physical challenges to create training data that teaches robots real-world skills. Download on iOS and Android.',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  )
}
