import { faker } from '@faker-js/faker'

import { parsePhoneNumberFromString } from 'libphonenumber-js'
import { expect, test } from './fixtures'

function getRandomInt(min: number, max: number): number {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

function generateValidUSPhoneNumber(): string | null {
  let attempts = 0
  const maxAttempts = 100 // Maximum attempts to find a valid number

  while (attempts < maxAttempts) {
    // Generate random area code and local number
    const areaCode = getRandomInt(200, 999).toString() // Avoiding 000-199 for safety
    const localNumber = getRandomInt(1000000, 9999999).toString()

    const phoneNumber = `+1${areaCode}${localNumber}`
    const parsedNumber = parsePhoneNumberFromString(phoneNumber, 'US')

    if (parsedNumber && parsedNumber.isValid()) {
      return parsedNumber.formatNational()
    }

    attempts++
  }

  return null // Return null if a valid number wasn't found after maxAttempts
}

test.beforeEach(async ({ signupPage }) => {
  await signupPage.goto()
})

test.describe('Signup flows', () => {
  test('Can navigate to the login page', async ({ page, signupPage }) => {
    await signupPage.loginLink.click()
    await expect(page).toHaveURL('https://fynbos.test/login')
  })

  test('Can navigate through the flow', async ({ page, signupPage }) => {
    await signupPage.screenshot('signup-flow-landing.png')

    await signupPage.getStartedButton.click()
    // TODO: use the following for error checking
    // await expect(page.getByText('Sign up')).toBeVisible()

    await signupPage.screenshot('signup-flow-about.png')
    await signupPage.inputAboutFields(
      faker.person.firstName(),
      faker.person.lastName(),
      faker.internet.email(),
      'amer',
      'United States of America'
    )
    await signupPage.continueButton.click()

    await signupPage.screenshot('signup-flow-phone.png')

    // Faker doesn't have a way to generate a valid phone number for the US
    const generatedNumber = generateValidUSPhoneNumber()
    if (!generatedNumber) {
      throw new Error('Failed to generate a valid US phone number')
    }

    await signupPage.inputPhoneFields(generatedNumber)
    await signupPage.continueButton.click()

    await signupPage.screenshot('signup-flow-otp.png', [
      page.getByText(/(phone_android\+\d+)/)
    ])
    await signupPage.inputOtpFields('123456')
    await signupPage.verifyButton.click()

    await signupPage.screenshot('signup-flow-password.png')
    await signupPage.inputPasswordFields(faker.internet.password(), true)
    await signupPage.confirmButton.click()

    await page.waitForURL('https://fynbos.test/wallet-address')
    await signupPage.screenshot('signup-flow-wallet-address.png', [
      page.getByLabel('Wallet address')
    ])

    await page.goBack()
    await expect(page).toHaveURL('https://fynbos.test/wallet-address')
    await signupPage.saveButton.click()

    await expect(page).toHaveURL('https://fynbos.test')
  })

  test('Reloading the page clears the form and routes to the beginning of the flow', async ({
    page,
    signupPage
  }) => {
    await signupPage.getStartedButton.click()
    await page.reload()
    await expect(page).toHaveURL('/signup')
    await expect(page.getByRole('heading', { name: 'Sign up' })).toBeVisible()
    await expect(signupPage.getStartedButton.isVisible()).resolves.toBe(true)
  })
})
