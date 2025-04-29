import { test } from '@playwright/test'
import { PayPage } from '../pages/pay'

test.describe('Payment', () => {
  let page: PayPage
  const term = 'playwright1'

  test.beforeEach(async ({ page: defaultPage }) => {
    page = new PayPage(defaultPage)
    await page.goto()
  })

  test('Successfully sends money to another user', async () => {
    await test.step('Search for user', async () => {
      await page.search(term)
    })

    await test.step('Create payment', async () => {
      await page.createPayment(term)
    })

    await test.step('Fill amount', async () => {
      await page.setAmount()
    })

    await test.step('Send payment', async () => {
      return await page.confirm()
    })

    await test.step('Validate payment', async () => {
      await page.validatePayment()
    })
  })
})
