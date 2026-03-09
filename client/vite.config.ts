import path from "path";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
  server: {
    proxy: {
      "/api/sso": {
        target: "http://localhost:8080",
        changeOrigin: true,
        rewrite: (pathValue) => pathValue.replace(/^\/api\/sso/, ""),
      },
      "/api/event": {
        target: "http://localhost:6660",
        changeOrigin: true,
        rewrite: (pathValue) => pathValue.replace(/^\/api\/event/, ""),
      },
      "/api/booking": {
        target: "http://localhost:3434",
        changeOrigin: true,
        rewrite: (pathValue) => pathValue.replace(/^\/api\/booking/, ""),
      },
      "/api/searching": {
        target: "http://localhost:4541",
        changeOrigin: true,
        rewrite: (pathValue) => pathValue.replace(/^\/api\/searching/, ""),
      },
      "/api/wallet": {
        target: "http://localhost:2436",
        changeOrigin: true,
        rewrite: (pathValue) => pathValue.replace(/^\/api\/wallet/, ""),
      },
    },
  },
});
