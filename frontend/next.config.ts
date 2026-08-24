import type { NextConfig } from "next";

// The /api proxy lives in app/api/[...path]/route.ts so that API_ORIGIN is read
// at runtime; a rewrites() entry would be frozen into the build manifest.
const nextConfig: NextConfig = {};

export default nextConfig;
