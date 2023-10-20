import type { Locator, Page } from '@playwright/test'
import { expect } from '@playwright/test'

export class SignupPage {
  readonly getStartedButton: Locator
  readonly continueButton: Locator
  readonly verifyButton: Locator
  readonly confirmButton: Locator
  readonly saveButton: Locator
  readonly loginLink: Locator

  // About
  readonly firstName: Locator
  readonly lastName: Locator
  readonly email: Locator
  readonly country: Locator

  // Phone
  readonly phone: Locator
  readonly otp: Locator

  // Password
  readonly password: Locator
  readonly terms: Locator

  // Wallet address
  readonly walletAddress: Locator

  constructor(readonly page: Page) {
    this.page = page
    this.getStartedButton = page.getByRole('button', {
      name: "Let's get started"
    })
    this.continueButton = page.getByRole('button', {
      name: 'Continue'
    })
    this.verifyButton = page.getByRole('button', { name: 'Verify' })
    this.confirmButton = page.getByRole('button', { name: 'Confirm' })
    this.saveButton = page.getByRole('button', { name: 'Save' })
    this.loginLink = page.getByRole('link', { name: 'Log in' })

    this.firstName = page.getByLabel('First name †')
    this.lastName = page.getByLabel('Last name †')
    this.email = page.getByLabel('Email address')
    this.country = page.locator('[id="headlessui-combobox-input-\\:r9\\:"]')
    this.phone = page.getByLabel('Mobile number')
    this.otp = page.getByLabel('Verification code')
    this.password = page.getByLabel('Password')
    this.terms = page.getByRole('checkbox', {
      name: 'I agree to the Fynbos Privacy Policy, Terms of Use, and E-sign Agreement.'
    })
    this.walletAddress = page.getByLabel('Wallet address')
  }

  async goto() {
    await this.page.goto('/signup')
    await expect(this.page).toHaveTitle('Sign up')
  }

  async screenshot(name: string, mask?: Locator[] | undefined) {
    await expect(this.page).toHaveScreenshot(name, {
      // For some reason certain text gets shifted
      maxDiffPixels: 1000,
      mask: [this.page.getByLabel('Loading shapes'), ...(mask ?? [])]
    })
  }

  // TODO: check meta tags

  async inputAboutFields(
    firstName: string,
    lastName: string,
    email: string,
    country: string,
    countryOption: string
  ) {
    await this.firstName.fill(firstName)
    await this.lastName.fill(lastName)
    await this.email.fill(email)
    await this.page.getByRole('button', { name: 'unfold_more' }).click()
    // await page.getByRole('option', { name: 'United States of America' }).click();
    // await this.country.fill(country)
    await this.page.getByRole('option', { name: countryOption }).click()
  }

  // TODO separate this for easier error checking?
  async inputPhoneFields(phone: string) {
    await this.phone.fill(phone)
  }

  async inputOtpFields(otp: string) {
    await this.otp.fill(otp)
  }

  async inputPasswordFields(password: string, check: boolean) {
    await this.password.fill(password)
    if (check) await this.terms.check()
  }
}
