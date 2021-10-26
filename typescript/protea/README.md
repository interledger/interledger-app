# Protea

Fynbos frontend.

## Get started

We use [Kratos](https://www.ory.sh/kratos/docs/) as our user identity service.
This will need to be running before the frontend can be started.

```shell
docker-compose -f ../../services/kratos/docker-compose-dev.yaml up -d --force-recreate
```

This project uses yarn 2. We don't use
[zero installs](https://yarnpkg.com/features/zero-installs) or
[pnp](https://yarnpkg.com/features/pnp) because at the time of scaffolding the
project some of our dependencies were incompatible.

```shell
yarn # Install dependencies
yarn dev # Start a dev server
yarn lint # Run linting
yarn format # Format with prettier
```

### Learn more

- [Next.js Documentation](https://nextjs.org/docs)
- [Tailwind documentation](https://tailwindcss.com/docs)

## Deployment

Currently deploys to Vercel.

Preview branches deploy to
[protea-don-fynbosdev.vercel.app](https://protea-don-fynbosdev.vercel.app)
(i.e., not main).

## Pages

Any file in `pages` with the `page.tsx`, or `blog.mdx` extension will be
rendered as a page by Next.

## Blog

### Writing a blog post

Blog posts go in `pages/blog/`. Each post has its own directory with an
`index.blog.mdx` at its root. This is done like this so that any images/files
needed for the post can be stored with the post.

Posts can be written using [mdx](https://mdxjs.com/) and will be styled and
rendered appropriately for you. To achieve this place the following code snippet
at the top of the file:

```typescript jsx
import { BlogLayout } from 'components'
import { cairin } from '../authors'
export default ({ children }) => <BlogLayout meta={meta} children={children} />

export const meta = {
  title: 'Building the Fynbos blog',
  authors: [cairin],
  description: '',
  date: '2021-09-10',
  slug: 'building-the-fynbos-blog',
  preview: false
}
```

The default export provides a wrapper for the mdx (once it's converted to html)
and thus provides the styling and meta.

The `meta` object provides necessary info for rendering everything other than
the body of the post (in both the list and post views):

- `title` The title of the post.
- `authors` An array of authors. All employees will have an author object, but
  you can add external collaborators' inline when needed.
- `description` An excerpt that will only show on the blog list page.
- `date` The published date of the post in iso format. **Make sure you update
  this for each post.**
- `slug` The slug to route to. **NB: This must mach the name of the file, not
  the full route. This is used by the blog list when routing to the new page.**
- `preview` Not used currently.

### Gotchas

- The meta object type can be found in `lib/blog/types`.
- There must be a blank line after the meta object. We use a regex to extract
  the meta object from the file for the list view.
- The top level header in your post should be a markdown level 2 header
  (`## Header`).
