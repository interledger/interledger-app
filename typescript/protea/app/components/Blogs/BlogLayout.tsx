import Image from 'next/image'
import { FC } from 'react'
import Head from 'next/head'
import { Header } from '../Header'
import { BackIcon, Container, Footer, Router, Routes } from 'components'
import type { Author, BlogMeta } from 'lib/blog'
import { DateTime } from 'luxon'
import Link from 'next/link'

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
        <div className='my-20 font-display text-5xl font-medium leading-normal sm:mt-28'>
          {meta.title}
        </div>
        <div className='flex flex-col justify-start sm:flex-row'>
          <div className='sm:mr-20 sm:w-60'>
            <div className='sticky top-[112px] border-t-2 border-b-2 border-black sm:border-b-0'>
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
                <span className='mt-12 hidden items-center text-primary sm:flex'>
                  <span className='mr-2'>
                    <BackIcon />
                  </span>{' '}
                  Back to blogs
                </span>
              </Router>
            </div>
          </div>
          <article className='prose mt-12 max-w-full sm:mt-6 sm:max-w-sm md:max-w-md lg:max-w-prose'>
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
    <div className='mb-6 flex'>
      <Image alt='Author profile image.' src={avatar} width='48' height='48' />
      <div className='ml-3 flex flex-grow flex-col'>
        <div className='font-medium'>{name}</div>
        <Link href={Routes.twitter + twitterHandle}>
          <a
            className='focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-1'
            aria-label='Author twitter'
          >
            <span className='text-primary'>@{twitterHandle}</span>
          </a>
        </Link>
      </div>
    </div>
  )
}
