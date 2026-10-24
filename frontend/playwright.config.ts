import { defineConfig } from '@playwright/test';

// Browser smoke tests against the local stack. PostgreSQL must be running
// (`docker compose up -d --wait`); the API and the Vite dev server are
// started here unless they already are.
export default defineConfig({
  testDir: './e2e',
  use: {
    baseURL: 'http://localhost:5173',
    channel: 'chrome',
    trace: 'retain-on-failure',
  },
  webServer: [
    {
      command: 'go run ./cmd/server',
      cwd: '..',
      url: 'http://localhost:8080/health',
      reuseExistingServer: false,
      env: { GAME_SPEED: '600', GAME_TICK_INTERVAL: '1s' },
      timeout: 120_000,
    },
    {
      command: 'npm run dev',
      url: 'http://localhost:5173',
      reuseExistingServer: true,
    },
  ],
});
