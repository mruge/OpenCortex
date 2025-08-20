/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    appDir: true,
  },
  async rewrites() {
    return [
      {
        source: '/api/data/:path*',
        destination: process.env.DATA_ABSTRACTOR_URL + '/:path*',
      },
      {
        source: '/api/ai/:path*',
        destination: process.env.AI_ABSTRACTOR_URL + '/:path*',
      },
      {
        source: '/api/exec/:path*',
        destination: process.env.EXEC_AGENT_URL + '/:path*',
      },
      {
        source: '/api/orchestrator/:path*',
        destination: process.env.ORCHESTRATOR_URL + '/:path*',
      },
    ];
  },
}

module.exports = nextConfig