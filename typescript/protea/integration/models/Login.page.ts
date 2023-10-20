import type { Locator, Page } from '@playwright/test'
import { expect } from '@playwright/test'

export class LoginPage {
  readonly LogInButton: Locator
  readonly signupLink: Locator
  readonly forgotPasswordLink: Locator

  readonly email: Locator
  readonly password: Locator

  constructor(readonly page: Page) {
    this.page = page
    this.LogInButton = page.getByRole('button', {
      name: 'Log in'
    })
    this.forgotPasswordLink = page.getByRole('link', {
      name: 'Forgot password?'
    })
    this.signupLink = page.getByRole('link', { name: 'Sign up' })

    this.email = page.getByLabel('Email')
    this.password = page.getByLabel('Password')
  }

  async goto() {
    await this.page.goto('/login')
    await expect(this.page).toHaveTitle('Log in')
  }

  async screenshot(name: string, mask?: Locator[] | undefined) {
    await expect(this.page).toHaveScreenshot(name, {
      // For some reason certain text gets shifted
      maxDiffPixels: 1000,
      mask: [this.page.getByLabel('Loading shapes'), ...(mask ?? [])]
    })
  }

  // TODO: check meta tags

  async inputFields(email: string, password: string) {
    await this.email.fill(email)
    await this.password.fill(password)
  }
}
