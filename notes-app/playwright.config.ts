import { defineConfig } from '@playwright/test';

// Runs against the stack started by `make up` (app on :3000, API on :8080), in the installed Chrome:
// no browser to download.
export default defineConfig({
  testDir: 'test/e2e',
  fullyParallel: false,
  workers: 1,
  use: {
    baseURL: 'http://localhost:3000',
    channel: 'chrome',
    colorScheme: 'light',
  },
});
