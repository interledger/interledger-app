import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import clsx from 'clsx'
import { toRemixMeta } from 'react-datocms'
import type { ApplicationProps } from '~/components'
import {
  Chip,
  ChipColor,
  Fab,
  Layouts,
  MarketingPageWithSections,
  Router
} from '~/components'
import type { SectionRecord } from '~/generated/dato-cms-graphql'
import { BlogPostModelOrderBy } from '~/generated/dato-cms-graphql'
import { getAllBlogPosts } from '~/lib/blog.server'
import { getBlogPage } from '~/lib/marketing.server'

export async function loader({ request }: LoaderArgs) {
  const { blogpage, footer } = await getBlogPage()
  const posts = await getAllBlogPosts({
    orderBy: [BlogPostModelOrderBy.DateDesc]
  })
  return json({ posts, blogpage, footer })
}

export const handle: ApplicationProps = {
  layout: (match) => (match.data.isUser ? Layouts.Wallet : Layouts.Marketing),
  scaffold: {
    header: {},
    fab: Fab.Pay,
    footer: (match) => match.data.footer
  }
}

export function meta({ data, params }: any) {
  return {
    ...toRemixMeta(data.blogpage.seoMeta),
    'twitter:url': 'https://fynbos.app/blog',
    'og:url': 'https://fynbos.app/blog'
  }
}

export default function Page() {
  const { posts, blogpage } = useLoaderData<typeof loader>()
  return (
    <>
      {blogpage?.body.map((section) => (
        <MarketingPageWithSections
          key={section.id}
          section={section as SectionRecord}
        >
          <div className='mt-20 grid w-full grid-cols-12 gap-y-8 px-4 py-20 lg:gap-10 lg:px-0'>
            {posts &&
              posts.map((post, index) => (
                <Router
                  className={clsx(
                    'col-span-full',
                    index % 3 !== 0 && 'lg:col-span-6 lg:min-h-max'
                  )}
                  to={`/blog/${post.slug}`}
                  key={post.slug}
                >
                  <li className='flex h-full cursor-pointer flex-col items-start justify-start space-y-4 rounded-xl bg-mk-section p-4 pb-6 hover:bg-mk-section-hover sm:flex-row-reverse sm:justify-between sm:space-x-8 sm:space-y-0 sm:space-x-reverse sm:p-8'>
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
                            author.name +
                            (index == array.length - 1 ? '' : ', ')
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
          </div>
        </MarketingPageWithSections>
      ))}
    </>
  )
}
