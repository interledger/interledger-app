import type { NextPage } from 'next'
import Link from 'next/link'
import { getAllPosts, BlogMeta } from 'lib/blog'
import { Container, Footer, Header } from '../../components'
import { FC } from 'react'
import { DateTime } from 'luxon'

// TODO: Rename file to index.page.tsx when we launch the blog

type BlogPageProps = {
  posts: BlogMeta[]
}

const BlogPage: NextPage<BlogPageProps> = ({ posts }) => {
  return (
    <Container>
      <Header />
      <main className='flex-grow'>
        <div className='flex flex-col px-4 sm:p-8 mt-12 sm:mt-28 mb-12 sm:mb-20 space-y-4 leading-normal w-[340px]'>
          <span className='text-3xl font-display font-medium '>
            Follow our progress
          </span>
          <span>
            The ins-and-outs of how we do things, and the problems we solve
            along the way.
          </span>
        </div>
        <List>
          {posts.map((post) => (
            <Link
              href={{
                pathname: '/blog/[slug]',
                query: { slug: post.slug }
              }}
              key={post.slug}
            >
              <a>
                <ListItem>
                  <span className='text-lg text-gray-700 dark:text-gray-100'>
                    {DateTime.fromISO(post.date).toFormat('dd LLLL yyyy')}
                    {' | '}
                    {post.authors.map((author, index, array) => {
                      return (
                        author.name + (index == array.length - 1 ? '' : ', ')
                      )
                    })}
                  </span>
                  <span className='text-4xl font-display font-medium my-2'>
                    {post.title}
                  </span>
                  <span className='text-2xl text-gray-700 dark:text-gray-100'>
                    {post.description}
                  </span>
                </ListItem>
              </a>
            </Link>
          ))}
        </List>
      </main>
      <Footer />
    </Container>
  )
}

export default BlogPage

export const getStaticProps = async () => {
  const content = await getAllPosts()

  return {
    props: { posts: content }
  }
}

const List: FC = ({ children }) => {
  return <ul className='flex flex-col space-y-8'>{children}</ul>
}

const ListItem: FC = ({ children }) => {
  return (
    <li className='flex flex-col justify-start p-4 sm:p-8 hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer'>
      {children}
    </li>
  )
}
