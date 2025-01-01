/** @type {import('next').NextConfig} */
const { i18n } = require('./next-i18next.config')
const nextConfig = {
  i18n,
  distDir: 'build',
  output: 'standalone',
  reactStrictMode: true,
  basePath: '/admin',
  async redirects() {
    return [
      {
        source: '/',
        destination: '/login',
        permanent: true,
      },
    ]
  },
}

module.exports = nextConfig
