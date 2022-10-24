import { DateTime } from 'luxon'

import * as postA from '~/routes/blog/connecting-the-internet-economy.mdx'
import * as postB from '~/routes/blog/card-payments-still-suck.mdx'
const modules = [postA, postB]

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

export async function getAllPosts(): Promise<BlogMeta[]> {
  const posts = modules.sort(
    (mod1, mod2) =>
      DateTime.fromJSDate(mod2.attributes.meta.date).toSeconds() -
      DateTime.fromJSDate(mod1.attributes.meta.date).toSeconds()
  )
  return posts.map((mod) => {
    return {
      ...mod.attributes.meta,
      slug: mod.filename.replace(/\.mdx?$/, ''),
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
