import { test } from '@playwright/test'
import { PayPage } from '../fixtures/pages/pay'

test.describe('Payment', () => {
  test('should have the ability to search', async ({ page }) => {
    const payPage = new PayPage(page)
    await payPage.goto()
    await payPage.search('radu')
  })
})
