import { expect, test } from '@playwright/test'
import { openCalendar, selectCalendar } from './helpers'

test.describe('calendar critical flows', () => {
  test('month view side panel stays usable on shorter desktop screens', async ({ page }) => {
    const eventTitle = `Month panel ${Date.now()}`

    await page.setViewportSize({ width: 1280, height: 640 })
    await openCalendar(page)

    await page.getByTestId('calendar-month-day').nth(14).click()
    await expect(page.getByTestId('calendar-day-panel')).toBeVisible()

    await page.locator('#new-ev-title').fill(eventTitle)
    await page.locator('#new-ev-desc').fill('layout regression check')
    await page.getByRole('button', { name: 'Add' }).click()

    await expect(page.getByTestId('calendar-day-panel')).not.toBeVisible()
    await expect(page.getByText(eventTitle).first()).toBeVisible()
    await expect.poll(async () => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBeTruthy()
  })

  test('creates an all-day event from the calendar UI', async ({ page }) => {
    const eventTitle = `E2E all-day ${Date.now()}`

    await openCalendar(page)
    await page.getByRole('button', { name: 'List' }).click()
    await page.getByRole('button', { name: 'New Event' }).click()
    await page.locator('#new-ev-title-list').fill(eventTitle)
    await page.getByLabel('All day').check()
    await page.getByRole('button', { name: 'Add' }).click()

    await expect(page.getByText(eventTitle)).toBeVisible()
    await expect(page.getByText('All day').first()).toBeVisible()
  })

  test('renames and publishes a calendar, then switches public views', async ({ page, context }) => {
    const calendarName = `E2E Calendar ${Date.now()}`

    await openCalendar(page)
    await selectCalendar(page, 'My Calendar')
    await page.getByRole('button', { name: /manage calendars/i }).click()

    const nameInput = page.getByLabel('Calendar name')
    await nameInput.fill(calendarName)
    await page.getByRole('button', { name: 'Save' }).click()
    await expect(page.getByText('Saved')).toBeVisible()

    await expect(page.getByText(calendarName).first()).toBeVisible()

    await page.getByRole('button', { name: 'Share' }).click()
    await page.getByRole('dialog').getByRole('button', { name: 'Create share link' }).click()

    const publicUrlText = page.getByRole('dialog').locator('text=/http:\\/\\/127\\.0\\.0\\.1:8091\\/public\\//').first()
    await expect(publicUrlText).toBeVisible()
    const publicUrl = await publicUrlText.textContent()
    if (!publicUrl) throw new Error('public URL was not created')

    const publicPage = await context.newPage()
    await publicPage.goto(publicUrl)

    await expect(publicPage.getByRole('heading', { name: calendarName })).toBeVisible()
    await publicPage.getByRole('tab', { name: 'Week' }).click()
    await expect(publicPage.getByRole('tab', { name: 'Week', selected: true })).toBeVisible()
    await publicPage.getByRole('tab', { name: 'List' }).click()
    await expect(publicPage.getByRole('tab', { name: 'List', selected: true })).toBeVisible()
    await publicPage.getByRole('tab', { name: 'Month' }).click()
    await expect(publicPage.getByRole('tab', { name: 'Month', selected: true })).toBeVisible()
  })

  test('deleting a calendar also removes its events from the visible schedule', async ({ page }) => {
    const calendarName = `Delete Calendar ${Date.now()}`
    const eventTitle = `Delete Event ${Date.now()}`

    await openCalendar(page)
    await page.getByRole('button', { name: /all calendars|my calendar|e2e calendar/i }).click()
    await page.getByRole('button', { name: 'New calendar' }).click()
    await page.getByPlaceholder('Calendar name').fill(calendarName)
    await page.getByRole('button', { name: 'Create' }).click()

    await selectCalendar(page, calendarName)
    await expect(page.getByRole('button', { name: new RegExp(calendarName) }).first()).toBeVisible()
    await page.getByRole('button', { name: 'List' }).click()
    await page.getByRole('button', { name: 'New Event' }).click()
    await page.locator('#new-ev-title-list').fill(eventTitle)
    await page.getByRole('button', { name: 'Add' }).click()
    await expect(page.getByText(eventTitle)).toBeVisible()

    await page.getByRole('button', { name: /manage calendars/i }).click()
    page.once('dialog', (dialog) => dialog.accept())
    await page.getByRole('dialog').getByRole('button', { name: 'Delete', exact: true }).click()

    await expect(page.getByText(eventTitle)).not.toBeVisible()
    await expect(page.getByRole('button', { name: /all calendars/i })).toBeVisible()
  })
})
