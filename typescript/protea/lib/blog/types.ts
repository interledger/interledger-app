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
  preview?: boolean
}
