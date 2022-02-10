import type { NextPage } from 'next'
import { getAllPosts, BlogMeta } from 'lib/blog'
import { Container, Footer, Header, Router, Routes } from 'components'
import { FC } from 'react'
import { DateTime } from 'luxon'

type BlogPageProps = {
  posts: BlogMeta[]
}

const BlogPage: NextPage<BlogPageProps> = ({ posts }) => {
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
            <Router
              href={{
                pathname: Routes.blogPost,
                query: { slug: post.slug }
              }}
              key={post.slug}
            >
              <ListItem>
                <span className='text-lg text-gray-700'>
                  {DateTime.fromISO(post.date).toFormat('dd LLLL yyyy')}
                  {' | '}
                  {post.authors.map((author, index, array) => {
                    return author.name + (index == array.length - 1 ? '' : ', ')
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
    <li className='flex cursor-pointer flex-col justify-start p-4 hover:bg-gray-50 sm:p-8'>
      {children}
    </li>
  )
}
