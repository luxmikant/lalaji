import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Expose the API base URL at build time
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080",
  },
};

export default nextConfig;
