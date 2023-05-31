import type { MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Chip, ChipColor, Layouts, Router } from '~/components'

import { useLoaderData } from '@remix-run/react'
import { BlogPostModelOrderBy } from '~/generated/dato-cms-graphql'
import { getAllBlogPosts } from '~/lib/blog.server'

export const meta: MetaFunction<typeof loader> = () => {
  const metaContent = {
    title: 'Fynbos blog',
    description:
      'All the ins-and-outs of how we do things, and the problems we solve along the way.'
  }

  return {
    title: metaContent.title,
    description: metaContent.description,
    'og:title': metaContent.title,
    'og:url': 'https://fynbos.app/blog',
    'og:description': metaContent.description,
    'twitter:url': 'https://fynbos.app/blog',
    'twitter:title': metaContent.title,
    'twitter:description': metaContent.description
  }
}

export async function loader() {
  const posts = await getAllBlogPosts({
    orderBy: [BlogPostModelOrderBy.DateDesc]
  })
  return json({
    posts: posts
  })
}

export const handle = {
  layout: Layouts.LandingLayout
}

export default function Page() {
  const { posts } = useLoaderData<typeof loader>()
  return (
    <main className='relative mx-auto w-full flex-grow overflow-x-visible px-4 sm:max-w-lg sm:px-0 lg:max-w-3xl xl:max-w-[59rem]'>
      <div className='relative col-span-full mb-2 h-20'>
        <div className='absolute -left-4 top-0 h-20 w-20 rounded-tl-full bg-sky-50 lg:-left-40' />
        <div className='absolute -right-4 top-0 h-20 w-20 rounded-tl-full bg-sky-200 lg:-right-40' />
        <div className='absolute -right-40 top-20 hidden h-20 w-20 rounded-tl-full bg-slate-50 lg:block' />
        <div className='absolute -right-40 top-[25.5rem] hidden h-20 w-20 rounded-full bg-lime-200 lg:block' />
        <div className='absolute -left-40 top-[35.5rem] hidden h-20 w-20 rounded-bl-full bg-yellow-100 lg:block' />
      </div>
      <div className='flex flex-col space-y-3 text-center sm:space-y-6'>
        <h1 className='font-display text-3xl font-medium sm:text-4xl'>Blog</h1>
        <h2 className='sm:text-2xl'>
          All the ins-and-outs of how we do things, and the problems we solve
          along the way.
        </h2>
      </div>

      <ul className='mt-12 flex flex-col space-y-6 sm:mt-20'>
        {posts &&
          posts.map((post) => (
            <Router to={`/blog/${post.slug}`} key={post.slug}>
              <li className='flex cursor-pointer flex-col items-start justify-start space-y-4 rounded-xl bg-mk-section p-4 pb-6 hover:bg-mk-section-hover sm:flex-row-reverse sm:justify-between sm:space-x-8 sm:space-y-0 sm:space-x-reverse sm:p-8'>
                <div className='min-w-max'>
                  <img
                    src={post.shapes?.url}
                    alt=''
                    className='hidden w-[7.5rem] lg:flex'
                  />
                  <img
                    src={post.shapesMobile?.url}
                    alt=''
                    className='flex h-10 lg:hidden'
                  />
                </div>
                {post._status === 'draft' && (
                  <div className='sticky -right-10'>
                    <Chip color={ChipColor.purple}>Draft</Chip>
                  </div>
                )}
                <div className='flex flex-col space-y-4 rounded-xl'>
                  <span className='font-display text-2xl font-medium'>
                    {post.title}
                  </span>
                  <span className='text-sm text-medium'>
                    {post.authors.map((author, index, array) => {
                      return (
                        author.name + (index == array.length - 1 ? '' : ', ')
                      )
                    })}
                    {' | '}
                    {post.date}
                  </span>
                  <span className='text-medium'>{post.description}</span>
                </div>
              </li>
            </Router>
          ))}
      </ul>
      <div className='relative col-span-full mt-20 h-20'>
        <div className='absolute -left-4 top-0 h-20 w-20 rounded-full bg-slate-300 lg:-left-20' />
        <div className='absolute -right-4 top-0 h-20 w-20 rounded-tl-full bg-slate-100 lg:-right-20' />
        <div className='absolute -right-40 top-0 hidden h-20 w-20 rounded-tl-full bg-slate-100 lg:block' />
      </div>
    </main>
  )
}
