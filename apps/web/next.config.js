/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  devIndicators: false,
  async rewrites() {
    const apiOrigin = process.env.RP_API_PROXY_URL?.replace(/\/+$/, "");
    return apiOrigin ? [{ source: "/api/v1/:path*", destination: `${apiOrigin}/api/v1/:path*` }] : [];
  },
  experimental: { cpus: 2 },
  allowedDevOrigins: ["pulse.riders.agency", "demo.example.test"],
  turbopack: {
    root: __dirname,
  },
};

module.exports = nextConfig;
