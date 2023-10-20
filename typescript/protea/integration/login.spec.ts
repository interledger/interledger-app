import { expect, test } from './fixtures'

test.beforeEach(async ({ loginPage }) => {
  await loginPage.goto()
})

test.describe('Login flows', () => {
  test('Can navigate to the signup page', async ({ page, loginPage }) => {
    await loginPage.signupLink.click()
    await expect(page).toHaveURL('https://fynbos.test/signup')
  })

  test('Can navigate to the forgot password page', async ({
    page,
    loginPage
  }) => {
    await loginPage.forgotPasswordLink.click()

    await expect(page).toHaveTitle('Recover account')
    await expect(page).toHaveURL(
      new RegExp('^https://fynbos.test/recovery', 'i')
    )
  })

  test('A valid user can sign in', async ({ page, loginPage }) => {
    await loginPage.screenshot('login.png')

    await loginPage.inputFields('LfSGlVD@nnKdPjh.info', 'fynboslocal')
    await loginPage.LogInButton.click()

    await expect(page).toHaveURL('https://fynbos.test')
  })

  test('Having incorrect password throws a snackbar', async ({
    page,
    loginPage
  }) => {
    await loginPage.inputFields('LfSGlVD@nnKdPjh.info', 'something else')
    await loginPage.LogInButton.click()
    await expect(
      page.getByText(
        'The provided credentials are invalid.Contact supportclose'
      )
    ).toBeVisible()
  })

  test('Having incorrect email throws a snackbar', async ({
    page,
    loginPage
  }) => {
    await loginPage.inputFields('notauser@nnKdPjh.info', 'fynboslocal')
    await loginPage.LogInButton.click()
    await expect(
      page.getByText(
        'The provided credentials are invalid.Contact supportclose'
      )
    ).toBeVisible()
  })
})
