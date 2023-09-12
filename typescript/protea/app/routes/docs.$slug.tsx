import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import clsx from 'clsx'
import type { Node } from 'datocms-structured-text-utils'
import { isCode } from 'datocms-structured-text-utils'
import type { FC, ReactNode } from 'react'
import { useEffect, useRef } from 'react'
import type { ResponsiveImageType } from 'react-datocms'
import { Image, StructuredText, toRemixMeta } from 'react-datocms'
import { getHighlighter, renderToHtml } from 'shiki'
import type { ApplicationProps } from '~/components'
import {
  CardContent,
  CardHeader,
  CardTitle,
  GridColumn,
  Layouts,
  Router,
  WalletGrid
} from '~/components'
import {
  Prose,
  renderCodeNodeRule,
  renderLinkNodeRule
} from '~/components/Content'
import { useDocsStore } from '~/components/Scaffold/Docs/useDocsStore'
import { getCurrentDocPage } from '~/data/content.server'
import type {
  DocRecord,
  InlineImageRecord,
  InlineVideoRecord
} from '~/generated/dato-cms-graphql'
import { sanitizeHTML } from '~/lib/fetchAndSanitizeHTML.server'

export async function loader({ request, params }: LoaderArgs) {
  if (process.env.FYNBOS_ENV == 'prod')
    throw json(null, { status: 404, statusText: 'Not found' })

  const { footer, doc } = await getCurrentDocPage({
    filter: { slug: { eq: params.slug } }
  })

  const highlighter = await getHighlighter({ theme: 'css-variables' })

  if (!doc) throw json(null, { status: 404, statusText: 'Not found' })

  doc.sections = doc.sections.map((section) => {
    if (!section.content) return section
    section.content.value.document.children =
      section.content.value.document.children.map((child: Node) => {
        if (isCode(child)) {
          let tokens = highlighter.codeToThemedTokens(
            child.code,
            child.language
          )

          return {
            ...child,
            code: sanitizeHTML(
              renderToHtml(tokens, {
                elements: {
                  pre: ({ children }) =>
                    `<pre class='flex rounded-xl bg-nav p-1'>${children}</pre>`,
                  code: ({ children }) =>
                    `<code class='language-${child.language} flex w-full min-w-max flex-col'>${children}</code>`,
                  line: ({ children, index }) =>
                    `<span class='${clsx(
                      'w-full px-2',
                      child.highlight?.includes(index) &&
                        'rounded-lg bg-nav-active',
                      // Has highlight above
                      child.highlight?.includes(index - 1) && 'rounded-t-none',
                      // Has highlight below
                      child.highlight?.includes(index + 1) && 'rounded-b-none'
                    )}'>${children || '\n'}</span>`
                }
              })
            )
          }
        }
        return child
      })
    return section
  })

  return json({
    footer,
    doc: doc as DocRecord
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Docs,
  scaffold: {
    header: {
      title: (match) => match.data.doc.title
    },
    footer: (match) => match.data.footer
  }
}

export function meta({ data, params }: any) {
  return {
    ...toRemixMeta(data.doc.seoMeta),
    'twitter:url': 'https://fynbos.app/docs',
    'og:url': 'https://fynbos.app/docs'
  }
}

export default function Page() {
  const { doc } = useLoaderData<typeof loader>()
  return (
    <WalletGrid>
      <GridColumn className='col-span-full max-w-prose'>
        {doc &&
          doc.sections &&
          doc.sections.length > 0 &&
          doc.sections.map((section) => {
            return (
              <div key={section.id} className='flex w-full flex-col'>
                {section.title && (
                  <CardHeader>
                    <SectionTitle
                      id={section.id}
                      docId={doc.id}
                      slug={section.slug || ''}
                    >
                      {section.title}
                    </SectionTitle>
                  </CardHeader>
                )}
                <CardContent>
                  <Prose className='prose-p:leading-6'>
                    <StructuredText
                      data={section.content as any}
                      customNodeRules={[renderCodeNodeRule, renderLinkNodeRule]}
                      renderBlock={({ record }) => {
                        switch (record.__typename) {
                          case 'InlineImageRecord':
                            return (
                              <>
                                <Image
                                  pictureClassName='m-0'
                                  className='w-full dark:hidden lg:hidden'
                                  data={{
                                    ...((record as InlineImageRecord)
                                      .imageMobile
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
                                    ...((record as InlineImageRecord)
                                      .imageDarkMobile
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
                          case 'InlineVideoRecord':
                            switch (
                              (record as InlineVideoRecord).video?.provider
                            ) {
                              case 'youtube':
                                return (
                                  <iframe
                                    className='w-full'
                                    style={{ aspectRatio: '16 / 9' }}
                                    src={`https://www.youtube-nocookie.com/embed/${
                                      (record as InlineVideoRecord).video
                                        ?.providerUid
                                    }`}
                                    title={
                                      (record as InlineVideoRecord).video?.title
                                    }
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
                  </Prose>
                </CardContent>
              </div>
            )
          })}
      </GridColumn>
    </WalletGrid>
  )
}

type SectionTitleProps = {
  id: string
  docId: string
  slug: string
  children?: ReactNode
}

const SectionTitle: FC<SectionTitleProps> = ({ id, docId, slug, children }) => {
  let ref = useRef<HTMLHeadingElement>(null)

  const registerHeading = useDocsStore((state) => state.registerHeading)
  useEffect(() => {
    registerHeading(id, docId, ref)
  })

  return (
    <Router to={`#${slug}`}>
      <CardTitle className='scroll-mt-20' id={slug} ref={ref}>
        {children}
      </CardTitle>
    </Router>
  )
}
