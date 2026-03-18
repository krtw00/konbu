import { expect, type Page } from '@playwright/test'

export const testUser = {
  name: 'E2E Owner',
  email: 'e2e-owner@example.com',
  password: 'password123',
}

export async function bootstrapAccount(page: Page) {
  await page.goto('/')

  await expect(page.getByText('Welcome to konbu')).toBeVisible()
  await page.getByPlaceholder('Name').fill(testUser.name)
  await page.getByPlaceholder('Email').fill(testUser.email)
  await page.getByPlaceholder('Password', { exact: true }).fill(testUser.password)
  await page.getByPlaceholder('Confirm password', { exact: true }).fill(testUser.password)
  await Promise.all([
    page.waitForResponse((response) =>
      response.url().includes('/api/v1/auth/register') && response.request().method() === 'POST',
    ),
    page.getByRole('button', { name: 'Create Account' }).click(),
  ])
  await page.waitForLoadState('networkidle')

  const calendarButton = page.getByRole('button', { name: 'Calendar' })
  const authenticated = await calendarButton.isVisible().catch(() => false)

  if (!authenticated) {
    const loginResponse = await page.context().request.post(new URL('/api/v1/auth/login', page.url()).toString(), {
      data: {
        email: testUser.email,
        password: testUser.password,
      },
    })
    expect(loginResponse.ok()).toBeTruthy()
    await page.goto('/')
  }

  await expect(calendarButton).toBeVisible()
}

export async function openCalendar(page: Page) {
  await page.goto('/')
  await page.getByRole('button', { name: 'Calendar' }).click()
  await expect(page.getByRole('heading', { name: 'Calendar' })).toBeVisible()
}

export async function selectCalendar(page: Page, name: string) {
  await page.getByRole('button', { name: /all calendars|my calendar|e2e calendar/i }).click()
  await page.getByRole('button', { name }).click()
}
