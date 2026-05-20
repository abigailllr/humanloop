import type { NextConfig } from 'next'

const config: NextConfig = {
  env: {
    API_URL: process.env.API_URL ?? 'http://localhost:8080',
    API_KEY: process.env.API_KEY ?? '',
  },
}

export default config
