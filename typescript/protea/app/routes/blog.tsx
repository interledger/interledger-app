import { Container, Footer, Header, Icon, Logo, Router } from '~/components'
import type { FC } from 'react'
import type { LoaderFunction } from '@remix-run/node'
import { json } from '@remix-run/node'

import { Outlet, useLoaderData, useLocation } from '@remix-run/react'
import type { Author, BlogMeta } from '~/lib/blog.server'
import { getAllPosts } from '~/lib/blog.server'
import { route } from 'routes-gen'

type LoaderData = {
  posts: BlogMeta[]
}

export const loader: LoaderFunction = async () => {
  return json({
    posts: await getAllPosts()
  })
}

export default function Page() {
  const { posts } = useLoaderData<LoaderData>()
  const location = useLocation()
  const isPost = location.pathname.search(/\/blog\/[A-z]+/) > -1

  if (isPost) {
    return (
      <BlogLayout
        meta={posts.find((post) => post.slug == location.pathname.substring(6))}
      >
        <Outlet />
      </BlogLayout>
    )
  } else {
    return (
      <Container>
        <Header />
        <main className='flex-grow'>
          <div className='mt-12 mb-12 flex w-[340px] flex-col space-y-4 px-4 leading-normal sm:mt-28 sm:mb-20 sm:p-8'>
            <span className='font-display text-3xl font-medium '>
              Follow our progress
            </span>
            <span>
              The ins-and-outs of how we do things, and the problems we solve
              along the way.
            </span>
          </div>

          <List>
            {posts.map((post) => (
              <Router to={`/blog/${post.slug}`} key={post.slug}>
                <ListItem>
                  <span className='text-lg text-gray-700'>
                    {post.date}
                    {' | '}
                    {post.authors.map((author, index, array) => {
                      return (
                        author.name + (index == array.length - 1 ? '' : ', ')
                      )
                    })}
                  </span>
                  <span className='my-2 font-display text-4xl font-medium'>
                    {post.title}
                  </span>
                  <span className='text-2xl text-gray-700'>
                    {post.description}
                  </span>
                </ListItem>
              </Router>
            ))}
          </List>
          <Footer />
        </main>
      </Container>
    )
  }
}

const List: FC = ({ children }) => {
  return <ul className='flex flex-col space-y-8'>{children}</ul>
}

const ListItem: FC = ({ children }) => {
  return (
    <li className='flex cursor-pointer flex-col justify-start p-4 hover:bg-gray-50 sm:p-8'>
      {children}
    </li>
  )
}

type BlogLayoutProps = {
  meta?: BlogMeta
}

const BlogLayout: FC<BlogLayoutProps> = ({ children, meta }) => {
  return (
    <div className='w-full'>
      {/* Header */}
      <header className='sticky top-0 z-10 mx-auto flex h-16 w-full select-none items-center justify-start bg-white p-4 text-medium sm:max-w-lg sm:px-0 lg:max-w-3xl xl:max-w-4xl'>
        <Router className='lg:hidden' to={route('/home')}>
          <div className='-ml-3 p-3 text-medium'>
            <Icon>arrow_back</Icon>
          </div>
        </Router>
        <Router to={route('/')}>
          <Logo className='h-8' />
        </Router>
      </header>
      {/* Body */}
      <main className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        <div className='col-span-full my-20 lg:mt-28'>
          <span className='font-display text-4xl font-medium text-medium'>
            {meta?.title}
          </span>
        </div>
        <div className='relative col-span-full flex h-min flex-col justify-start border-t-2 border-b-2 border-black lg:sticky lg:top-16 lg:col-span-3 lg:border-b-0'>
          <div className='mt-6 mb-12'>{meta?.date}</div>
          {meta?.authors.map((author, index) => (
            <AuthorBlock
              key={index}
              name={author.name}
              avatar={author.avatar}
              twitterHandle={author.twitterHandle}
            />
          ))}
          <Router to={route('/blog')}>
            <span className='mt-12 hidden items-center text-primary lg:flex'>
              <span className='mr-2'>
                <Icon>arrow_back</Icon>
              </span>{' '}
              Back to blogs
            </span>
          </Router>
        </div>
        <article className='prose col-span-full max-w-full lg:col-start-5 lg:max-w-prose'>
          {children}
        </article>
      </main>
      <Footer />
    </div>
  )
}

const AuthorBlock: FC<Author> = ({ name, avatar, twitterHandle }) => {
  return (
    <div className='mb-6 flex'>
      <div className='h-12 w-12 rounded-full'>
        <img src={avatar} className='max-w-full' loading='lazy' alt='Avatar' />
      </div>
      <div className='ml-3 flex flex-grow flex-col'>
        <div className='font-medium'>{name}</div>
        <Router.a
          to={'https://twitter.com/' + twitterHandle}
          aria-label='Author twitter'
        >
          <span className='text-primary'>@{twitterHandle}</span>
        </Router.a>
      </div>
    </div>
  )
}
