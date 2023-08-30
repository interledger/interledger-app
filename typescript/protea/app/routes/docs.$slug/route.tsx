import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import clsx from 'clsx'
import type { Node } from 'datocms-structured-text-utils'
import { isCode } from 'datocms-structured-text-utils'
import type { FC, ReactNode } from 'react'
import { useEffect, useRef } from 'react'
import type { ResponsiveImageType } from 'react-datocms'
import {
  Image,
  StructuredText,
  renderNodeRule,
  toRemixMeta
} from 'react-datocms'
import { getHighlighter, renderToHtml } from 'shiki'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  GridColumn,
  Layouts,
  Router,
  WalletGrid
} from '~/components'
import { useDocsStore } from '~/components/Scaffold/Docs/useDocsStore'
import type {
  DocRecord,
  InlineImageRecord,
  InlineVideoRecord
} from '~/generated/dato-cms-graphql'
import { sanitizeHTML } from '~/lib/fetchAndSanitizeHTML'
import { getCurrentDocPage } from '~/lib/marketing.server'

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
                    `<pre class='-mx-2 flex rounded-xl bg-nav p-1 last:-mb-2'>${children}</pre>`,
                  code: ({ children }) =>
                    `<code class='language-${child.language} flex w-full min-w-max flex-col'>${children}</code>`,
                  line: ({ children, index }) =>
                    `<span class='${clsx(
                      'w-full px-3',
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
      <GridColumn className='col-span-full'>
        {doc &&
          doc.sections &&
          doc.sections.length > 0 &&
          doc.sections.map((section) => {
            return (
              <Card key={section.id}>
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
                  <Prose>
                    <StructuredText
                      data={section.content as any}
                      renderInlineRecord={({ record }) => {
                        // TODO figure out how to get link to the record
                        // console.log('renderInlineRecord', record)
                        return <>renderInlineRecord</>
                      }}
                      renderLinkToRecord={({ record }) => {
                        // console.log('renderLinkToRecord', record)
                        return <>renderLinkToRecord</>
                      }}
                      customNodeRules={[renderCodeNodeRule]}
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
              </Card>
            )
          })}
      </GridColumn>
    </WalletGrid>
  )
}

const renderCodeNodeRule = renderNodeRule(isCode, ({ node, key }) => {
  return <div key={key} dangerouslySetInnerHTML={{ __html: node.code }} />
})

type SectionTitleProps = {
  id: string
  docId: string
  slug: string
  children?: ReactNode
}

const SectionTitle: FC<SectionTitleProps> = ({ id, docId, slug, children }) => {
  let ref = useRef<HTMLHeadingElement>(null)

  const registerHeading = useDocsStore((state) => state.registerHeading)
  console.log('register heading for side nav', slug)
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

type ProseProps = {
  children?: ReactNode
}

const Prose: FC<ProseProps> = ({ children }) => {
  return (
    <div className='prose-p prose prose-slate max-w-none dark:prose-invert prose-h1:font-display prose-h1:font-medium prose-h2:font-display prose-h2:font-medium prose-h3:font-display prose-h3:font-medium prose-h4:font-display prose-h4:font-medium prose-h5:font-display prose-h5:font-medium prose-h6:font-display prose-h6:font-medium prose-a:rounded prose-a:text-primary prose-a:no-underline prose-a:focus-visible:outline prose-a:focus-visible:outline-2 prose-a:focus-visible:outline-focus prose-blockquote:border-0 prose-blockquote:p-0 prose-blockquote:text-3xl prose-blockquote:font-normal prose-blockquote:not-italic prose-code:font-normal prose-code:tracking-wider'>
      {children}
    </div>
  )
}
