import react from '@vitejs/plugin-react';
import path from "path";
import { defineConfig } from 'vite';

// The app calls the API through the relative /api path; in development and
// preview, forward it to the locally running Go server.
const proxy = { '/api': 'http://localhost:8080' };

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: { proxy },
  preview: { proxy },
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "./src"),
    },
  },
  optimizeDeps: {
    exclude: ['lucide-react'],
  },
});
