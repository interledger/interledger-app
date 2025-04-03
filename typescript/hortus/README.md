# Hortus

Integration tests

## Steps

> [!NOTE]
> Node LTS/Jod is used for this package.

1. Install dependencies  
```sh
pnpm install
```

2. Install Playwright browsers
```sh
playwright install --with-deps chromium
```

3. Run the tests
```sh
pnpm test
# or if you want to see the Playwright UI and run the tests manually
pnpm test:ui
```
