import type { Author, BlogMeta } from '~/lib/blog.server'
import type { FC, ReactNode } from 'react'
import { AnchorRouter, BlogShapes, Icon, Router } from '~/components'
import { route } from 'routes-gen'

type BlogLayoutProps = {
  children?: ReactNode
  meta?: BlogMeta
}

export const BlogLayout: FC<BlogLayoutProps> = ({ children, meta }) => {
  return (
    <main className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
      <div className='col-span-full -mt-4 hidden w-full flex-col content-start bg-mk-page pb-20 lg:flex'>
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

type ProseProps = {
  children?: ReactNode
}

const Prose: FC<ProseProps> = ({ children }) => {
  return (
    <div className='prose-h5:font-display prose-h5:font-medium prose-h6:font-display prose-h6:font-medium prose prose-slate prose-h1:font-display prose-h1:font-medium prose-h2:font-display prose-h2:font-medium prose-h3:font-display prose-h3:font-medium prose-h4:font-display prose-h4:font-medium prose-a:rounded prose-a:text-primary prose-a:no-underline prose-a:focus-visible:outline prose-a:focus-visible:outline-2 prose-a:focus-visible:outline-focus prose-blockquote:border-0 prose-blockquote:p-0 prose-blockquote:text-3xl prose-blockquote:font-normal prose-blockquote:not-italic prose-code:font-normal prose-code:tracking-wider prose-pre:rounded-xl prose-pre:bg-slate-800 prose-pre:p-4 prose-pre:pb-6 dark:prose-invert'>
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
