# Kratos templates

This folder allows formatting/styling of custom email templates for kratos.

## Makefile

The makefile allows you can to run `make format`. This will:

- Compile the relevant tailwind classes in the `.html` files, and will store
  this is `minified.css`.
- Copy the contents of each `.html` file into it's respective `.gotmpl` file and
  inline the css from the previous step.

> Note for the templates that have an `.html` file, only the `.html` file should
> be edited. The `.gotmpl` files that don't have a corresponding `.html` file
> can be edited directly.
