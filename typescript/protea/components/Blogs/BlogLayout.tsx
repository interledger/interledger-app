import Image from 'next/image'
import { FC } from 'react'
import Head from 'next/head'
import { Header } from '../Header'
import { Container, Footer, Router, Routes } from 'components'
import type { Author, BlogMeta } from 'lib/blog'
import { DateTime } from 'luxon'

type BlogLayoutProps = {
  meta: BlogMeta
}

export const BlogLayout: FC<BlogLayoutProps> = ({ children, meta }) => {
  return (
    <Container>
      <Head>
        <title>{meta.title} | Fynbos</title>
      </Head>
      <Header />
      <main className='px-4 sm:px-8'>
        <div className='my-20 sm:mt-28 text-5xl font-medium font-display leading-normal'>
          {meta.title}
        </div>
        <div className='flex flex-col sm:flex-row justify-start'>
          <div className='sm:mr-20 sm:w-60'>
            <div className='sticky top-[112px] border-t-2 border-b-2 sm:border-b-0 border-black dark:border-white'>
              <div className='mt-6 mb-12'>
                {DateTime.fromISO(meta.date).toFormat('dd LLLL yyyy')}
              </div>
              {meta.authors.map((author, index) => (
                <AuthorBlock
                  key={index}
                  name={author.name}
                  avatar={author.avatar}
                  twitterHandle={author.twitterHandle}
                />
              ))}
              <Router href={Routes.blog}>
                <span className='text-primary dark:text-secondary hidden sm:flex items-center mt-12'>
                  <span className='material-icons-sharp mr-2'>arrow_back</span>{' '}
                  Back to blogs
                </span>
              </Router>
            </div>
          </div>
          <article className='prose dark:prose-dark max-w-full sm:max-w-sm md:max-w-md lg:max-w-prose mt-12 sm:mt-6'>
            {children}
          </article>
        </div>
      </main>
      <Footer />
    </Container>
  )
}

const AuthorBlock: FC<Author> = ({ name, avatar, twitterHandle }) => {
  return (
    <div className='flex mb-6'>
      <Image alt='Author profile image.' src={avatar} width='48' height='48' />
      <div className='flex flex-col flex-grow ml-3'>
        <div className='font-medium'>{name}</div>
        <Router
          href={{
            pathname: Routes.twitter,
            query: { handle: twitterHandle }
          }}
          aria-label='Author twitter'
        >
          <span className='text-primary dark:text-secondary'>
            @{twitterHandle}
          </span>
        </Router>
      </div>
    </div>
  )
}
