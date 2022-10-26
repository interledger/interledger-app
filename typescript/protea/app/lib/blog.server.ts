import { DateTime } from 'luxon'

import * as postA from '~/routes/blog.connecting-the-internet-economy.mdx'
import * as postB from '~/routes/blog.card-payments-still-suck.mdx'
import * as postC from '~/routes/blog.our-fynbos-family-meet-don.mdx'
const modules = [postA, postB, postC]

export type Author = {
  name: string
  twitterHandle: string
  avatar: string
}

export type BlogMeta = {
  title: string
  authors: Author[]
  description: string
  date: string
  slug: string
}

const authors: any = {
  fynbos: {
    name: 'Fynbos',
    twitterHandle: 'fynbosdev',
    avatar: '/icon.png'
  },
  cairin: {
    name: 'Cairin Michie',
    twitterHandle: 'cairinbruce',
    avatar: '/Frame 9.png'
  },
  adrian: {
    name: 'Adrian Hope-Bailie',
    twitterHandle: 'ahopebailie',
    avatar: '/adrian.profile.webp'
  }
}

export async function getCurrentPost(
  request: Request,
  meta: any
): Promise<BlogMeta> {
  const url = new URL(request.url)
  const slug = url.pathname.replace('/blog/', '')
  return {
    title: meta.title,
    authors: meta.authors.map(
      // TODO: handle random authors
      (author: string) => authors[author]
    ),
    description: meta.description,
    slug: slug,
    date: DateTime.fromJSDate(meta.date).toFormat('dd LLLL yyyy')
  }
}

export async function getAllPosts(): Promise<BlogMeta[]> {
  const posts = modules.sort(
    (mod1, mod2) =>
      DateTime.fromJSDate(mod2.attributes.meta.date).toSeconds() -
      DateTime.fromJSDate(mod1.attributes.meta.date).toSeconds()
  )
  return posts.map((mod) => {
    return {
      ...mod.attributes.meta,
      slug: mod.filename.slice(5, -4),
      authors: mod.attributes.meta.authors.map(
        // TODO: handle random authors
        (author: string) => authors[author]
      ),
      date: DateTime.fromJSDate(mod.attributes.meta.date).toFormat(
        'dd LLLL yyyy'
      )
    }
  })
}
