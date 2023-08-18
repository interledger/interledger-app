import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import type { FC, ReactNode } from 'react'
import type { ResponsiveImageType } from 'react-datocms'
import { Image, StructuredText, toRemixMeta } from 'react-datocms'
import type { ApplicationProps } from '~/components'
import { AnchorRouter, Layouts } from '~/components'
import type {
  DocRecord,
  InlineImageRecord,
  InlinePersonRecord,
  InlineTwitterEmbedRecord,
  InlineVideoRecord
} from '~/generated/dato-cms-graphql'
import { getCurrentDocPage } from '~/lib/marketing.server'

export async function loader({ request, params }: LoaderArgs) {
  if (process.env.FYNBOS_ENV == 'prod')
    throw json(null, { status: 404, statusText: 'Not found' })

  const { footer, doc } = await getCurrentDocPage({
    filter: { slug: { eq: params.slug } }
  })
  return json({
    footer,
    post: doc as DocRecord
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Docs,
  scaffold: {
    header: {
      title: (match) => match.data.allDocs[0].title
    },
    nav: (match) => match.data.allDocs,
    footer: (match) => match.data.footer
  }
}

export function meta({ data, params }: any) {
  return {
    ...toRemixMeta(data.allDocs.seoMeta),
    'twitter:url': 'https://fynbos.app/docs',
    'og:url': 'https://fynbos.app/docs'
  }
}

export default function Page() {
  const { post } = useLoaderData<typeof loader>()

  return (
    <>
      <div className='h-16 w-full bg-blue-500'></div>
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
    </>
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
