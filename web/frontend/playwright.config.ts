import path from 'path'
import { fileURLToPath } from 'url'
import { defineConfig, devices } from '@playwright/test'

const dirname = path.dirname(fileURLToPath(import.meta.url))
const storageState = path.resolve(dirname, 'playwright/.auth/user.json')

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  reporter: 'list',
  timeout: 30_000,
  expect: {
    timeout: 10_000,
  },
  use: {
    baseURL: 'http://127.0.0.1:8091',
    trace: 'on-first-retry',
  },
  webServer: {
    command: 'cd ../.. && ./bin/e2e-server',
    url: 'http://127.0.0.1:8091/health',
    reuseExistingServer: false,
    timeout: 120_000,
  },
  projects: [
    {
      name: 'setup',
      testMatch: /auth\.setup\.ts/,
    },
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        storageState,
      },
      dependencies: ['setup'],
      testIgnore: /auth\.setup\.ts/,
    },
  ],
})
