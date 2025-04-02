import type { Page } from '@playwright/test'

export class PayPage {
  public readonly page: Page

  constructor(page: Page) {
    this.page = page
  }

  async goto() {
    await this.page.goto('/pay')
  }

  async search(text: string) {
    const searchBox = this.page.getByPlaceholder('Search for someone to pay')
    await searchBox.fill(text)
  }
}
