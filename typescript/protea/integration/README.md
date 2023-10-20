## Integration testing with playwright

### Testing strategy

- Test user flows
  - Golden path
  - Error paths
- Test accessibility
- Visual testing

### Implementation

- Prefer fixtures
- Use page object models to simplify flow tests that route through multiple pages

### Useful links
- Nice implementation https://github.com/leather-wallet/extension/blob/dev/tests/page-object-models/onboarding.page.ts
- page object models https://playwright.dev/docs/pom
- fixtures https://playwright.dev/docs/test-fixtures