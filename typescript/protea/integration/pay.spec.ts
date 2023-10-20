import { USER_COOKIE } from './constants'
import { expect, test } from './fixtures'

test.beforeEach(async ({ page, payPage }) => {
  // const authFile = 'playwright/.auth/user.json'
  // await page.context().storageState({ path: authFile })
  // TODO Use the auth files when we need multiple users
  await page.context().addCookies([USER_COOKIE])
  await payPage.goto()
})

test.describe('Pay flows', () => {
  test('Can open the pay modal and route to the amount page with a search', async ({
    page,
    payPage
  }) => {
    // THis is currently flaky because the data returned by search updates after first render
    // I think this
    await payPage.gotoModal()
    await payPage.screenshot('pay-flow-search-modal.png')

    await payPage.inputSearch('fyn')
    await page.waitForLoadState()

    await payPage.search.press('ArrowDown')
    await payPage.search.press('Enter')

    await page.waitForURL(
      new RegExp(
        '^https://fynbos.test/pay/[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}',
        'i'
      )
    )
    await expect(page).toHaveTitle('Pay')
  })

  test('Can navigate through the flow', async ({ page, payPage }) => {
    await payPage.screenshot('pay-flow-search.png')

    await payPage.inputSearch('fyn')
    await page.waitForLoadState()

    await payPage.search.press('Enter')

    await page.waitForURL(
      new RegExp(
        '^https://fynbos.test/pay/[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}',
        'i'
      )
    )
    await expect(page).toHaveTitle('Pay')

    // TODO probably need to mask the name
    await payPage.screenshot('pay-flow-amount.png', [
      page.getByTestId('card').nth(0)
    ])

    await expect(page.getByText('Total amount to debit$ 0.00')).toBeVisible()
    await expect(page.getByText('Payment protection (+3%)$ 0.00')).toBeVisible()

    const inputSendAmount = 10
    await payPage.inputSendAmount(inputSendAmount.toString())

    await page.waitForEvent('requestfinished')

    await expect(page.getByText('Payment protection (+3%)$ 0.00')).toBeVisible()
    await expect(page.getByText(`Total amount to debit$ 10.00`)).toBeVisible()

    await payPage.togglePaymentProtection()

    await page.waitForEvent('requestfinished')

    const paymentProtection = parseInt(
      (
        await page.getByText('Payment protection (+3%)$ 0.').allTextContents()
      )[0].split('.')[1]
    )

    const totalDebit = parseInt(
      (
        await page.getByText(`Total amount to debit$ 10.`).allTextContents()
      )[0].split('.')[1]
    )

    expect(paymentProtection).toBeGreaterThanOrEqual(10)
    expect(paymentProtection).toBeLessThanOrEqual(50)
    expect(paymentProtection).toEqual(totalDebit)

    await payPage.togglePaymentProtection()
    await page.waitForEvent('requestfinished')

    await payPage.continueButton.click()
    await payPage.screenshot('pay-flow-confirm.png', [
      page.getByTestId('card').nth(0)
    ])

    await payPage.checkTerms(true)
    await payPage.confirmButton.click()

    if (await payPage.otp.isVisible({ timeout: 1000 })) {
      await payPage.inputOtpFields('123123')
      await payPage.verifyButton.click()
    }

    await page.waitForURL('/')
  })
})
