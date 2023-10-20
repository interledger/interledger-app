import { test as base } from '@playwright/test'
import { LoginPage, PayPage, SignupPage } from './models'

interface TestFixtures {
  signupPage: SignupPage
  loginPage: LoginPage
}

export * from '@playwright/test'

export const test = base.extend<TestFixtures>({
  signupPage: async ({ page }, use) => {
    await use(new SignupPage(page))
  },
  loginPage: async ({ page }, use) => {
    await use(new LoginPage(page))
  }
})
