import { test, expect } from '@playwright/test'

test('submits feedback from the public feedback page', async ({ page }) => {
  await page.goto('/feedback')

  await page.getByLabel(/email|メールアドレス/i).fill('feedback@example.com')
  await page.getByRole('button', { name: /feature request|機能要望/i }).click()
  await page.getByLabel(/summary|概要/i).fill('Feedback page smoke test')
  await page.getByLabel(/current problem|困っていること/i).fill('Need a clearer input format for requests')
  await page.getByLabel(/proposed change|提案内容/i).fill('Use structured fields instead of one free-form message')
  await page.getByRole('button', { name: /send feedback|送信/i }).click()

  await expect(page.getByText(/thanks|ありがとうございます/i)).toBeVisible()
})
