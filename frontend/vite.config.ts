import react from '@vitejs/plugin-react';
import path from "path";
import { defineConfig } from 'vite';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  optimizeDeps: {
    exclude: ['lucide-react'],
  },
  define: {
    'process.env.API_BASE_URL': JSON.stringify(process.env.API_BASE_URL || 'http://localhost:8080/api/v1'),
  },
});
