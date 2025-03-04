import { expect, test } from '@playwright/test'
import { v4 } from 'uuid'

const FLOW_URL_PREFIX =
  '^http://interledger.test/flows/[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}/signup'

test.beforeEach(async ({ page }) => {
  await page.goto('http://interledger.test/signup')
})

test.describe('Signup', () => {
  test('Can run through the whole flow', async ({ page }) => {
    await expect(page).toHaveURL(new RegExp(FLOW_URL_PREFIX + '/about', 'i'))

    await page.locator('input[name="firstName"]').fill('Cairin')
    await page.locator('input[name="lastName"]').fill('Michie')
    await page.locator('input[role="combobox"]').fill('amer')
    await page.locator('input[role="combobox"]').press('ArrowDown')
    await page.locator('input[role="combobox"]').press('ArrowDown')
    await page.locator('input[role="combobox"]').press('Enter')
    await page.locator('input[name="email"]').fill(v4() + '@test.com')
    await page.locator('text=Continue').click()

    await expect(page).toHaveURL(new RegExp(FLOW_URL_PREFIX + '/phone', 'i'))

    await page.locator('input[name="phone"]').fill('5555555555')
    await page.locator('text=Send SMS').click()

    await expect(page).toHaveURL(new RegExp(FLOW_URL_PREFIX + '/sms', 'i'))

    await page.locator('input[name="code"]').fill('123456')
    await page.locator('text=Continue').click()

    await expect(page).toHaveURL(
      new RegExp(
        FLOW_URL_PREFIX +
          '/password\\?flow=[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}',
        'i'
      )
    )

    await page.locator('input[name="password"]').fill('alsdujkjfhasljfkhba3245')
    await page.locator('input[name="service-agreement"]').check()
    await page.locator('text=Confirm').click()

    await expect(page).toHaveURL('http://interledger.test/onboarding/unit')
  })

  test('The form values are successfully stored and refilled on navigation.', async ({
    page
  }) => {
    await expect(page).toHaveURL(new RegExp(FLOW_URL_PREFIX + '/about', 'i'))

    await page.locator('input[name="firstName"]').fill('Cairin')
    await page.locator('input[name="lastName"]').fill('Michie')
    await page.locator('input[role="combobox"]').fill('amer')
    await page.locator('input[role="combobox"]').press('ArrowDown')
    await page.locator('input[role="combobox"]').press('ArrowDown')
    await page.locator('input[role="combobox"]').press('Enter')
    await page.locator('input[name="email"]').fill(v4() + '@test.com')
    await page.locator('text=Continue').click()

    await expect(page).toHaveURL(new RegExp(FLOW_URL_PREFIX + '/phone', 'i'))

    await page.locator('input[name="phone"]').fill('5555555555')
    await page.locator('text=Send SMS').click()

    await expect(page).toHaveURL(new RegExp(FLOW_URL_PREFIX + '/sms', 'i'))

    await page.locator('input[name="code"]').fill('123456')
    await page.locator('text=Continue').click()

    await expect(page).toHaveURL(
      new RegExp(
        FLOW_URL_PREFIX +
          '/password\\?flow=[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}',
        'i'
      )
    )

    // TODO: also check default browser navigation, which is currently broken.
    await page.locator('button:has-text("arrow_back")').click()
    await expect(page).toHaveURL(new RegExp(FLOW_URL_PREFIX + '/sms', 'i'))

    await page.locator('button:has-text("arrow_back")').click()
    await expect(page).toHaveURL(new RegExp(FLOW_URL_PREFIX + '/phone', 'i'))

    await page.locator('button:has-text("arrow_back")').click()
    await expect(page).toHaveURL(new RegExp(FLOW_URL_PREFIX + '/about', 'i'))

    await expect(page.locator('input[name="firstName"]')).toHaveValue('Cairin')
    await page.locator('text=Continue').click()

    await expect(page).toHaveURL(new RegExp(FLOW_URL_PREFIX + '/phone', 'i'))
    await expect(page.locator('input[name="phone"]')).toHaveValue(
      '+15555555555'
    )
    await page.locator('text=Send SMS').click()

    await expect(page).toHaveURL(new RegExp(FLOW_URL_PREFIX + '/sms', 'i'))
    await expect(page.locator('input[name="code"]')).toHaveValue('123456')
    await page.locator('text=Continue').click()

    await page.locator('input[name="password"]').fill('alsdujkjfhasljfkhba3245')
    await page.locator('input[name="service-agreement"]').check()
    await page.locator('text=Confirm').click()

    await expect(page).toHaveURL('http://interledger.test/onboarding/unit')
  })
})
