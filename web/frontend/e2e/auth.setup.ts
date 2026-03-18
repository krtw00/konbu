import path from 'path'
import fs from 'fs'
import { fileURLToPath } from 'url'
import { test } from '@playwright/test'
import { bootstrapAccount } from './helpers'

const dirname = path.dirname(fileURLToPath(import.meta.url))
const storageState = path.resolve(dirname, '../playwright/.auth/user.json')

test('bootstrap account', async ({ page }) => {
  fs.mkdirSync(path.dirname(storageState), { recursive: true })
  await bootstrapAccount(page)
  await page.context().storageState({ path: storageState })
})
