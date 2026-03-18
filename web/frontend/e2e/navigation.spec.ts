import { expect, test, type Page } from '@playwright/test'

const navTargets = [
  'Home',
  'Memos',
  'Tables',
  'ToDo',
  'Calendar',
  'Tools',
  'Search',
  'AI Chat',
  'Settings',
]

async function waitForShellToSettle(page: Page) {
  const shell = page.locator('body')

  await expect(page.locator('nav')).toBeVisible()
  await expect(page.locator('main').first()).toBeVisible()
  await expect
    .poll(async () => (await shell.textContent()) ?? '', { timeout: 10_000 })
    .not.toContain('Loading...')
  await expect(shell).not.toContainText('This screen failed to load')
}

test('sidebar navigation keeps the shell visible across pages', async ({ page }) => {
  const pageErrors: string[] = []
  const nav = page.locator('nav').first()
  page.on('pageerror', (error) => {
    pageErrors.push(error.message)
  })

  await page.goto('/')
  await waitForShellToSettle(page)

  for (const name of navTargets) {
    await nav.getByRole('button', { name, exact: true }).click()
    await waitForShellToSettle(page)
  }

  expect(pageErrors).toEqual([])
})
