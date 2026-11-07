import { defineConfig } from '@playwright/test';

export const APP = process.env.E2E_APP ?? 'http://localhost:3000';
export const API = process.env.E2E_API ?? 'http://localhost:8080';
// The session of the e2e account, written by test/e2e/login.ts: every page and API call starts logged in.
export const SESSION = 'test-results/.session.json';

// Runs against the stack started by `make up` (app on :3000, API on :8080), in the installed Chrome:
// no browser to download. E2E_APP and E2E_API point it elsewhere.
export default defineConfig({
  testDir: 'test/e2e',
  globalSetup: './test/e2e/login.ts',
  fullyParallel: false,
  workers: 1,
  use: {
    baseURL: APP,
    channel: 'chrome',
    colorScheme: 'light',
    // Without the cookie of a language, the app follows the browser's: the tests read French.
    locale: 'fr-FR',
    storageState: SESSION,
  },
});
