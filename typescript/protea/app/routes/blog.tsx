import { AnchorRouter, Icon, Router, BlogShapes } from '~/components'
import type { FC } from 'react'
import { json } from '@remix-run/node'

import { Outlet, useLoaderData, useLocation } from '@remix-run/react'
import type { Author, BlogMeta } from '~/lib/blog.server'
import { getAllPosts } from '~/lib/blog.server'
import { route } from 'routes-gen'

export async function loader() {
  return json({
    posts: await getAllPosts()
  })
}

export default function Page() {
  const { posts } = useLoaderData<typeof loader>()
  const location = useLocation()
  const isPost = location.pathname.search(/\/blog\/[A-z]+/) > -1

  if (isPost) {
    return (
      <BlogLayout
        meta={posts.find(
          (post: BlogMeta) => post.slug == location.pathname.substring(6)
        )}
      >
        <Outlet />
      </BlogLayout>
    )
  } else {
    return (
      <main className='relative mx-auto w-full flex-grow overflow-x-visible px-4 sm:max-w-lg sm:px-0 lg:max-w-3xl xl:max-w-[59rem]'>
        <div className='relative col-span-full mb-2 h-20'>
          <div className='absolute top-0 -left-4 h-20 w-20 rounded-tl-full bg-sky-50 lg:-left-40' />
          <div className='absolute top-0 -right-4 h-20 w-20 rounded-tl-full bg-sky-200 lg:-right-40' />
          <div className='absolute top-20 -right-40 hidden h-20 w-20 rounded-tl-full bg-slate-50 lg:block' />
          <div className='absolute top-[25.5rem] -right-40 hidden h-20 w-20 rounded-full bg-lime-200 lg:block' />
          <div className='absolute top-[35.5rem] -left-40 hidden h-20 w-20 rounded-bl-full bg-yellow-100 lg:block' />
        </div>
        <div className='flex flex-col space-y-3 text-center sm:space-y-6'>
          <h1 className='font-display text-3xl font-medium sm:text-4xl'>
            Blog
          </h1>
          <h2 className='sm:text-2xl'>
            All the ins-and-outs of how we do things, and the problems we solve
            along the way.
          </h2>
        </div>

        <ul className='mt-12 flex flex-col space-y-6 sm:mt-20'>
          {posts.map((post) => (
            <Router to={`/blog/${post.slug}`} key={post.slug}>
              <li className='flex cursor-pointer flex-col justify-start space-y-4 rounded-xl bg-app p-4 pb-6 hover:bg-container sm:flex-row-reverse sm:justify-between sm:space-y-0 sm:space-x-8 sm:space-x-reverse sm:p-8'>
                <BlogShapes slug={post.slug} />
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
          <div className='absolute top-0 -left-4 h-20 w-20 rounded-full bg-slate-300 lg:-left-20' />
          <div className='absolute top-0 -right-4 h-20 w-20 rounded-tl-full bg-slate-100 lg:-right-20' />
          <div className='absolute top-0 -right-40 hidden h-20 w-20 rounded-tl-full bg-slate-100 lg:block' />
        </div>
      </main>
    )
  }
}

type BlogLayoutProps = {
  meta?: BlogMeta
}

const BlogLayout: FC<BlogLayoutProps> = ({ children, meta }) => {
  return (
    <main className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
      <div className='col-span-full -mt-4 hidden w-full flex-col content-start bg-page pb-20 lg:flex'>
        <Router
          className='hidden items-center p-4 text-primary lg:flex'
          to={route('/blog')}
        >
          <Icon className='mr-2'>arrow_back</Icon>
          Back to blogs
        </Router>
        <div className='flex w-full items-center justify-between space-x-20'>
          <h1 className='font-display text-3xl font-medium text-medium sm:text-5xl'>
            {meta?.title}
          </h1>
          <BlogShapes slug={meta?.slug || ''} large />
        </div>
      </div>
      <div className='col-span-2 mb-2 lg:hidden'>
        <BlogShapes slug={meta?.slug || ''} large />
      </div>
      <h1 className='col-span-full mb-4 font-display text-3xl font-medium text-medium sm:text-5xl lg:hidden'>
        {meta?.title}
      </h1>
      <div className='relative col-span-full flex h-min flex-col justify-start border-t-2 border-black lg:sticky lg:top-24 lg:col-span-3 lg:border-b-0'>
        <div className='mt-6 mb-2 lg:mb-6'>{meta?.date}</div>
        {meta?.authors.map((author, index) => (
          <AuthorBlock
            key={index}
            name={author.name}
            avatar={author.avatar}
            twitterHandle={author.twitterHandle}
          />
        ))}
      </div>
      <article className='col-span-full max-w-full lg:col-start-5 lg:max-w-prose'>
        <Prose>{children}</Prose>
      </article>
    </main>
  )
}

const Prose: FC = ({ children }) => {
  return (
    <div className='prose prose-slate prose-h1:font-display prose-h1:font-medium prose-h2:font-display prose-h2:font-medium prose-h3:font-display prose-h3:font-medium prose-h4:font-display prose-h4:font-medium prose-h5:font-display prose-h5:font-medium prose-h6:font-display prose-h6:font-medium prose-a:text-primary prose-a:no-underline prose-blockquote:font-normal prose-code:font-medium prose-pre:rounded-xl prose-pre:bg-container prose-pre:p-4 prose-pre:pb-6 prose-pre:text-strong prose-img:rounded-xl'>
      {children}
    </div>
  )
}

const AuthorBlock: FC<Author> = ({ name, avatar, twitterHandle }) => {
  return (
    <div className='mb-6 flex'>
      <img
        src={avatar}
        className='mr-3 hidden h-20 w-20 max-w-full rounded-full lg:flex'
        loading='lazy'
        alt='Avatar'
      />
      <div className='flex flex-grow flex-col lg:mt-3'>
        <div className='font-medium'>{name}</div>
        <AnchorRouter
          to={'https://twitter.com/' + twitterHandle}
          aria-label='Author twitter'
          className='text-primary'
        >
          @{twitterHandle}
        </AnchorRouter>
      </div>
    </div>
  )
}
