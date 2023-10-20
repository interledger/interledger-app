import type { Locator, Page } from '@playwright/test'
import { expect } from '@playwright/test'

export class PayPage {
  readonly payButton: Locator
  readonly continueButton: Locator
  readonly confirmButton: Locator
  readonly verifyButton: Locator

  // Search
  readonly search: Locator

  // Amount
  // await page.getByRole('link', { name: 'Connect new account add' })
  // await page.getByRole('option', { name: 'credit_card Default **** 1234 check' })
  readonly sendAmount: Locator
  readonly receiveAmount: Locator
  readonly paymentProtection: Locator
  readonly note: Locator

  // Confirm
  readonly terms: Locator
  readonly otp: Locator

  constructor(readonly page: Page) {
    this.page = page
    this.payButton = page.getByRole('button', { name: 'attach_money Pay' })
    this.continueButton = page.getByRole('button', {
      name: 'Continue'
    })
    this.confirmButton = page.getByRole('button', { name: 'Confirm payment' })
    this.verifyButton = page.getByRole('button', { name: 'Verify' })

    this.search = page.getByPlaceholder('Search for someone to pay')
    this.sendAmount = page.getByLabel('Amount to send')
    this.receiveAmount = page.getByLabel('Recipient gets')
    this.paymentProtection = page.getByRole('switch', {
      name: 'Payment protection switch'
    })
    this.note = page.getByLabel('Payment note (optional)')
    this.terms = page.getByLabel(
      'I authorize Fynbos to debit the card indicated for the amount noted on today’s date. I will not dispute Fynbos debiting my account, so long as the transaction corresponds to the terms in this online form and my agreement with Fynbos.'
    )
    this.otp = page.getByLabel('Verification code')
  }

  async goto() {
    await this.page.goto('/pay')
    await expect(this.page).toHaveTitle('Pay search')
  }

  async gotoModal() {
    await this.page.goto('/')
    await this.payButton.click()
    await expect(this.search).toBeVisible()
    await expect(this.page).toHaveTitle('Fynbos')
  }

  async screenshot(name: string, mask?: Locator[] | undefined) {
    await expect(this.page).toHaveScreenshot(name, {
      // For some reason certain text gets shifted
      maxDiffPixels: 1000,
      mask: [this.page.getByLabel('Loading shapes'), ...(mask ?? [])]
    })
  }

  // TODO: check meta tags

  async inputSearch(search: string) {
    await this.search.fill(search)
  }

  async inputSendAmount(amount: string) {
    await this.sendAmount.fill(amount)
  }

  async inputReceiveAmount(amount: string) {
    await this.receiveAmount.fill(amount)
  }

  async togglePaymentProtection() {
    await this.paymentProtection.click()
  }

  async checkTerms(check: boolean) {
    if (check) await this.terms.check()
    else await this.terms.uncheck()
  }

  async inputOtpFields(otp: string) {
    await this.otp.fill(otp)
  }
}
