/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: process.env.NODE_ENV === 'production' ? 'standalone' : undefined,
  i18n: {
    locales: ['en', 'th'],
    defaultLocale: 'en',
  },
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081/api',
  },
  webpack: (config, { dev, isServer }) => {
    // Add auto-version update plugin in development mode only
    if (dev && !isServer) {
      const UpdateVersionPlugin = require('./plugins/update-version-plugin');
      config.plugins.push(new UpdateVersionPlugin({
        throttleMs: 2000, // Update at most once per 2 seconds
        silent: true, // Silent operation
      }));
    }
    return config;
  },
}

module.exports = nextConfig

