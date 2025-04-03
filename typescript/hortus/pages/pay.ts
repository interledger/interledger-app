import { expect, type Page } from '@playwright/test'

export class PayPage {
  public readonly page: Page

  constructor(page: Page) {
    this.page = page
  }

  async goto() {
    await this.page.goto('/pay')
  }

  async search(term: string) {
    const searchField = this.page.getByPlaceholder('Search for someone to pay')
    await searchField.focus()
    await searchField.fill(term)
    await expect(this.page.locator(`text=${term}`)).toBeVisible()
  }

  async createPayment(term: string) {
    const button = this.page.locator('button', { hasText: term })
    await expect(button).toBeVisible()
    await button.click()
    await this.page.waitForURL('/pay/*')
    await expect(this.page.locator('text=Payment to')).toBeVisible()
  }

  async setAmount(amount: string = '0.5') {
    const amountField = this.page.locator('input#send')
    const noteField = this.page.locator('input#note')
    await amountField.fill(amount)
    await noteField.fill('Integration testing payment')
  }

  async confirm() {
    const continueBtn = this.page.locator('button', { hasText: 'continue' })
    await continueBtn.click()

    const confirmBtn = this.page.locator('button', {
      hasText: 'confirm payment'
    })
    const checkbox = this.page.locator('input#service-agreement')
    await checkbox.setChecked(true)

    await confirmBtn.click()
    const otp = await this.page
      .waitForSelector('input#otp', { timeout: 2000 })
      .catch(() => null)

    if (otp) {
      const verifyBtn = this.page.locator('button', { hasText: 'Verify' })
      otp.fill('123456')
      verifyBtn.click()
    }

    await this.page.waitForURL(process.env.BASE_URL)
  }
}
