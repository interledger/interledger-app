# Honeybee

Email templates for fynbos.

### Get started

```shell
# Install dependencies
yarn

# Build prod email templates
yarn build

# Format templates
yarn format
```

Any `.html` file in `./templates` will be compiled in to the `<article>` tag in
a specified `./layouts` file. Most of the time you can just use
`./layouts/main.html`. The output is dumped into `./production` with the same
directory structure. This then needs to be manually copied to sendgrid.

An example template:

```html
<template src="main">
  <h1>Welcome</h1>
  <p>Thank you for signing up to Fynbos.</p>
</template>
```

### Helpful esources

https://www.caniemail.com/
