import { json } from '@remix-run/node'
import type { ApplicationProps } from '~/components'
import {
  AnchorRouter,
  Chip,
  ChipColor,
  Icon,
  Layouts,
  Router
} from '~/components'

import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import type { UIMatch } from '@remix-run/react'
import { useLoaderData } from '@remix-run/react'
import type { ResponsiveImageType } from 'react-datocms'
import { Image, StructuredText } from 'react-datocms'
import { route } from 'routes-gen'
import { Prose } from '~/components/Content'
import { getCurrentBlogPost } from '~/data/content.server'
import type {
  BlogPostRecord,
  InlineImageRecord,
  InlinePersonRecord,
  InlineTwitterEmbedRecord,
  InlineVideoRecord
} from '~/generated/dato-cms-graphql'
import { datoMeta, mergeMeta } from '~/lib/meta'

export const meta: MetaFunction<typeof loader> = mergeMeta(
  ({ data }) => datoMeta(data?.post?._seoMetaTags),
  ({ location }) => [
    {
      name: 'og:url',
      content: `https://fynbos.app${location.pathname}`
    },
    {
      name: 'twitter:url',
      content: `https://fynbos.app${location.pathname}`
    }
  ]
)

export async function loader({ params }: LoaderFunctionArgs) {
  const { footer, blogPost } = await getCurrentBlogPost({
    filter: { slug: { eq: params.slug } }
  })
  return json({
    footer,
    post: blogPost as BlogPostRecord
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Marketing,
  scaffold: {
    header: {},
    footer: (match: UIMatch<typeof loader>) => match.data.footer
  }
}

export default function Page() {
  const { post } = useLoaderData<typeof loader>()

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
          <img
            src={post?.shapes?.url}
            alt=''
            className='hidden w-60 lg:block'
          />
        </div>
      </div>
      <div className='col-span-2 mb-2 lg:hidden'>
        <img
          src={post?.shapesMobile?.url}
          alt=''
          className='flex h-10 lg:hidden'
        />
      </div>
      <h1 className='col-span-full mb-4 font-display text-3xl font-medium text-medium sm:text-5xl lg:hidden'>
        {post?.title}
      </h1>
      <div className='relative col-span-full flex h-min flex-col justify-start border-t-2 border-black lg:sticky lg:top-24 lg:col-span-3 lg:border-b-0'>
        <div className='mb-2 mt-6 lg:mb-6'>{post?.date}</div>
        {post &&
          post?.authors.length > 0 &&
          post?.authors.map((author) => (
            <div key={author.id} className='mb-6 flex'>
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
          ))}
        {post?._status === 'draft' && (
          <div className='relative'>
            <Chip color={ChipColor.purple}>Draft</Chip>
          </div>
        )}
      </div>
      <article className='col-span-full max-w-full lg:col-start-5 lg:max-w-prose'>
        <Prose>
          {post && (
            <StructuredText
              data={post.content as any}
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
