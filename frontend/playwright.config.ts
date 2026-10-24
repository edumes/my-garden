import { defineConfig } from '@playwright/test';

// Browser smoke tests against the local stack. PostgreSQL must be running
// (`docker compose up -d --wait`); the API and the Vite dev server are
// started here. Override both ports to keep a manual preview running.
const apiPort = process.env.E2E_API_PORT ?? '8080';
const webPort = process.env.E2E_WEB_PORT ?? '5173';
const apiOrigin = `http://127.0.0.1:${apiPort}`;
const webOrigin = `http://127.0.0.1:${webPort}`;
export default defineConfig({
  testDir: './e2e',
  workers: 2,
  use: {
    baseURL: webOrigin,
    channel: 'chrome',
    trace: 'retain-on-failure',
  },
  webServer: [
    {
      command: 'go run ./cmd/server',
      cwd: '..',
      url: `${apiOrigin}/health`,
      reuseExistingServer: false,
      env: { PORT: apiPort, CORS_ORIGIN: webOrigin, GAME_SPEED: '600', GAME_TICK_INTERVAL: '1s' },
      timeout: 120_000,
    },
    {
      command: `npm run dev -- --host 127.0.0.1 --port ${webPort} --strictPort`,
      url: webOrigin,
      env: { API_PROXY_TARGET: apiOrigin },
      reuseExistingServer: false,
    },
  ],
});
