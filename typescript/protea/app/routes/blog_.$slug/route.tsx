import { json } from '@remix-run/node'
import type { ApplicationProps } from '~/components'
import {
  AnchorRouter,
  Chip,
  ChipColor,
  Fab,
  Icon,
  Layouts,
  Router
} from '~/components'

import type { LoaderArgs } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import type { FC, ReactNode } from 'react'
import type { ResponsiveImageType } from 'react-datocms'
import { Image, StructuredText, toRemixMeta } from 'react-datocms'
import { route } from 'routes-gen'
import type {
  InlineImageRecord,
  InlinePersonRecord,
  InlineTwitterEmbedRecord,
  InlineVideoRecord,
  PersonRecord
} from '~/generated/dato-cms-graphql'
import { getCurrentBlogPost } from '~/lib/blog.server'
import { getBlogRoute } from '~/lib/marketing.server'

export function meta({ data, params }: any) {
  return {
    ...toRemixMeta(data.post.seoMeta),
    'twitter:url': `https://fynbos.app/blog/${params.slug}`,
    'og:url': `https://fynbos.app/blog/${params.slug}`
  }
}

export async function loader({ params }: LoaderArgs) {
  const { footer } = await getBlogRoute()
  return json({
    footer,
    post: await getCurrentBlogPost({ filter: { slug: { eq: params.slug } } })
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Marketing,
  scaffold: {
    header: {},
    fab: Fab.Pay,
    footer: (match) => match.data.footer
  }
}

export default function Page() {
  const { post } = useLoaderData()

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
            {post?.title}
          </h1>
          <img src={post.shapes?.url} alt='' className='hidden w-60 lg:block' />
        </div>
      </div>
      <div className='col-span-2 mb-2 lg:hidden'>
        <img
          src={post.shapesMobile?.url}
          alt=''
          className='flex h-10 lg:hidden'
        />
      </div>
      <h1 className='col-span-full mb-4 font-display text-3xl font-medium text-medium sm:text-5xl lg:hidden'>
        {post?.title}
      </h1>
      <div className='relative col-span-full flex h-min flex-col justify-start border-t-2 border-black lg:sticky lg:top-24 lg:col-span-3 lg:border-b-0'>
        <div className='mb-2 mt-6 lg:mb-6'>{post?.date}</div>
        {post?.authors.map((author: PersonRecord) => (
          <AuthorBlock key={author.id} author={author} />
        ))}
        {post._status === 'draft' && (
          <div className='relative'>
            <Chip color={ChipColor.purple}>Draft</Chip>
          </div>
        )}
      </div>
      <article className='col-span-full max-w-full lg:col-start-5 lg:max-w-prose'>
        <Prose>
          {post && (
            <StructuredText
              data={post.content}
              renderBlock={({ record }) => {
                switch (record.__typename) {
                  case 'InlineImageRecord':
                    return (
                      <>
                        <Image
                          pictureClassName='m-0'
                          className='w-full dark:hidden lg:hidden'
                          data={{
                            ...((record as InlineImageRecord).imageMobile
                              ?.responsiveImage as ResponsiveImageType),
                            alt: (record as InlineImageRecord).altText
                          }}
                        />
                        <Image
                          pictureClassName='m-0'
                          className='hidden w-full lg:flex lg:dark:hidden'
                          data={{
                            ...((record as InlineImageRecord).image
                              ?.responsiveImage as ResponsiveImageType),
                            alt: (record as InlineImageRecord).altText
                          }}
                        />
                        <Image
                          pictureClassName='m-0'
                          className='hidden w-full dark:flex lg:hidden'
                          data={{
                            ...((record as InlineImageRecord).imageDarkMobile
                              ?.responsiveImage as ResponsiveImageType),
                            alt: (record as InlineImageRecord).altText
                          }}
                        />
                        <Image
                          pictureClassName='m-0'
                          className='hidden w-full lg:dark:flex'
                          data={{
                            ...((record as InlineImageRecord).imageDark
                              ?.responsiveImage as ResponsiveImageType),
                            alt: (record as InlineImageRecord).altText
                          }}
                        />
                      </>
                    )
                  case 'InlineTwitterEmbedRecord':
                    return (
                      <AnchorRouter to={record.url as string}>
                        <img
                          src={
                            (record as InlineTwitterEmbedRecord).imageOfTweet
                              ?.url
                          }
                          alt={
                            (record as InlineTwitterEmbedRecord).imageOfTweet
                              ?.alt || ''
                          }
                          className='w-full'
                        />
                      </AnchorRouter>
                    )
                  case 'InlinePersonRecord':
                    return (
                      <div className='flex content-start space-x-4'>
                        <Image
                          pictureClassName='m-0'
                          className='aspect-square'
                          data={
                            (record as InlinePersonRecord).avatar
                              ?.responsiveImage as ResponsiveImageType
                          }
                        />
                        <div className='flex flex-col justify-center text-medium'>
                          <span className='font-medium'>
                            {(record as InlinePersonRecord).name}
                          </span>
                          <span>{(record as InlinePersonRecord).role}</span>
                        </div>
                      </div>
                    )
                  case 'InlineVideoRecord':
                    switch ((record as InlineVideoRecord).video?.provider) {
                      case 'youtube':
                        return (
                          <iframe
                            className='w-full'
                            style={{ aspectRatio: '16 / 9' }}
                            src={`https://www.youtube-nocookie.com/embed/${
                              (record as InlineVideoRecord).video?.providerUid
                            }`}
                            title={(record as InlineVideoRecord).video?.title}
                            allow='accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture'
                            allowFullScreen
                          />
                        )
                      default:
                        return null
                    }

                  default:
                    return null
                }
              }}
            />
          )}
        </Prose>
      </article>
    </main>
  )
}

type ProseProps = {
  children?: ReactNode
}

const Prose: FC<ProseProps> = ({ children }) => {
  return (
    <div className='prose prose-slate dark:prose-invert prose-h1:font-display prose-h1:font-medium prose-h2:font-display prose-h2:font-medium prose-h3:font-display prose-h3:font-medium prose-h4:font-display prose-h4:font-medium prose-h5:font-display prose-h5:font-medium prose-h6:font-display prose-h6:font-medium prose-a:rounded prose-a:text-primary prose-a:no-underline prose-a:focus-visible:outline prose-a:focus-visible:outline-2 prose-a:focus-visible:outline-focus prose-blockquote:border-0 prose-blockquote:p-0 prose-blockquote:text-3xl prose-blockquote:font-normal prose-blockquote:not-italic prose-code:font-normal prose-code:tracking-wider prose-pre:rounded-xl prose-pre:bg-slate-800 prose-pre:p-4 prose-pre:pb-6'>
      {children}
    </div>
  )
}
//
const AuthorBlock: FC<{ author: PersonRecord }> = ({ author }) => {
  return (
    <div className='mb-6 flex'>
      <Image
        pictureClassName='m-0'
        className='mr-3 aspect-square'
        data={author.avatar?.responsiveImage as ResponsiveImageType}
      />
      <div className='flex flex-grow flex-col lg:mt-3'>
        <div className='font-medium'>{author.name}</div>
        <AnchorRouter
          to={author.twitterUrl as string}
          aria-label='Author twitter'
          className='text-primary'
        >
          @{author.twitterUrl?.split('/').pop()}
        </AnchorRouter>
      </div>
    </div>
  )
}
