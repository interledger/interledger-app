import { chromium } from '@playwright/test'
import { STATE_FILE } from 'fixtures/helpers'

export default async function () {
  const { EMAIL, PASSWORD, BASE_URL } = process.env

  const browser = await chromium.launch({ headless: true })
  const page = await browser.newPage()

  const loginUrl = new URL(BASE_URL)
  loginUrl.pathname = '/login'

  await page.goto(loginUrl.toString())
  await page.getByLabel('Email').fill(EMAIL)
  await page.getByLabel('Password').fill(PASSWORD)
  await page.getByRole('button', { name: 'log in' }).click()
  await page.waitForURL(BASE_URL)

  await page.context().storageState({ path: STATE_FILE })
}
