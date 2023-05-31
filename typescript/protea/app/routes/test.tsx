import { gql } from '@apollo/client'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import clsx from 'clsx'
import type { ApplicationProps } from '~/components'
import { ButtonRouter, Layouts, Shape } from '~/components'
import {
  CtaContentRecord,
  FeatureBlocksContentRecord,
  FeatureContentRecord,
  HeaderContentRecord,
  HeroContentRecord,
  HomeHeroContentRecord,
  Query,
  SectionModelContentField,
  ShowcaseContentRecord,
  StoryContentRecord,
  TeamContentRecord,
  TextContentRecord
} from '~/generated/dato-cms-graphql'
import { apolloClient } from '~/lib/apollo.server'

export const handle: ApplicationProps = {
  title: 'About',
  layout: Layouts.Marketing,
  scaffold: {
    header: {}
  }
}

type PageResponse = {
  homepage: Query['homepage']
  footer: Query['footer']
}

export async function loader() {
  const { homepage, footer } = await apolloClient
    .query<{
      homepage: Query['homepage']
      footer: Query['footer']
    }>({
      query: gql`
        query GetTestPageContent {
          homepage {
            id
            body {
              id
              content {
                ... on CtaContentRecord {
                  title
                  body
                  button {
                    id
                    displayText
                    url
                    button
                  }
                }
                ... on FeatureBlocksContentRecord {
                  blocks {
                    id
                    image {
                      id
                      url
                    }
                    title
                    direction
                    backgroundColour {
                      hex
                    }
                  }
                }
                ... on FeatureContentRecord {
                  title
                  body
                  image {
                    id
                    url
                  }
                  imageMobile {
                    id
                    url
                  }
                  rowReverse
                }
                ... on HeaderContentRecord {
                  id
                  title
                  shapes {
                    id
                    url
                  }
                }
                ... on HeroContentRecord {
                  id
                  title
                  body
                  button {
                    id
                    displayText
                    url
                    button
                  }
                  image {
                    id
                    url
                  }
                }
                ... on HomeHeroContentRecord {
                  id
                  iterations {
                    id
                    title
                    body
                    button {
                      id
                      displayText
                      url
                      button
                    }
                    image {
                      id
                      url
                    }
                  }
                }
                ... on ShowcaseContentRecord {
                  id
                  cases {
                    id
                    title
                    body
                    image {
                      id
                      url
                    }
                    imagedark {
                      id
                      url
                    }
                  }
                }
                ... on StoryContentRecord {
                  id
                  title
                  blurb {
                    value
                  }
                  bodyText {
                    value
                  }
                }
                ... on TeamContentRecord {
                  id
                  title
                  people {
                    id
                    person {
                      name
                      role
                      twitterUrl
                      linkedinUrl
                      fynbosUrl
                      avatar {
                        id
                        url
                        # TODO: Add responsive image
                      }
                    }
                  }
                }
                ... on TextContentRecord {
                  id
                  title
                  body
                  textCentered
                  button {
                    id
                    displayText
                    url
                    button
                  }
                }
              }
            }
            _status
            _firstPublishedAt
          }
          footer {
            id
          }
        }
      `
    })
    .then((res) => {
      console.log(res.data)
      return res.data
    })
    .catch((error) => {
      console.log(error)
      return { homepage: null, footer: null }
    })

  console.log(homepage, footer)

  return json({
    homepage,
    footer
  })
}

export default function Page() {
  const { homepage } = useLoaderData<typeof loader>()
  return homepage?.body.map((content, index) => {
    return (
      <div className='w-full overflow-hidden even:rounded-2xl even:bg-mk-section'>
        <div className='relative mx-auto flex w-full max-w-[59rem] flex-col items-center'>
          {content.content.map((innerContent) =>
            renderSectionModelContentField(
              innerContent as SectionModelContentField
            )
          )}
          {/*<div className='absolute -right-[10.5rem] top-0 block h-96 w-96 rounded-b-full rounded-tl-full bg-yellow-400' />*/}
        </div>
      </div>
    )
  })
}

function renderSectionModelContentField(content: SectionModelContentField) {
  switch (content.__typename) {
    case 'CtaContentRecord':
      return renderCtaContentRecord(content)
    case 'FeatureBlocksContentRecord':
      return renderFeatureBlocksContentRecord(content)
    case 'FeatureContentRecord':
      return renderFeatureContentRecord(content)
    case 'HeaderContentRecord':
      return renderHeaderContentRecord(content)
    case 'HeroContentRecord':
      return renderHeroContentRecord(content)
    case 'HomeHeroContentRecord':
      return renderHomeHeroContentRecord(content)
    case 'ShowcaseContentRecord':
      return renderShowcaseContentRecord(content)
    case 'StoryContentRecord':
      return renderStoryContentRecord(content)
    case 'TeamContentRecord':
      return renderTeamContentRecord(content)
    case 'TextContentRecord':
      return renderTextContentRecord(content)
    default:
      return <h1>Unknown</h1>
  }
}

// Stub out functions that return the content to replace the h1 tags in the switch cases above
// They should have typed input params and return the correct JSX
function renderCtaContentRecord(content: CtaContentRecord) {
  return (
    <div className='flex w-full flex-col space-y-16 px-4 py-10 lg:flex-row lg:space-x-16 lg:space-y-0 lg:px-0 lg:py-32'>
      <div className='flex items-start'>
        <Shape radius='rounded-full' color='bg-sky-400' width='w-32' />
        <Shape radius='rounded-l-full' color='bg-slate-200' width='w-32' />
      </div>
      <div className='flex flex-col space-y-6'>
        <h2 className='text-4xl font-medium'>{content.title}</h2>
        <p className='text-2xl text-medium'>{content.body}</p>
        <ButtonRouter
          className='h-20'
          shrink
          to={content.button[0].url as string}
        >
          {content.button[0].displayText}
        </ButtonRouter>
      </div>
    </div>
  )
}

function renderFeatureBlocksContentRecord(content: FeatureBlocksContentRecord) {
  return <h1>{content.__typename}</h1>
}

function renderFeatureContentRecord(content: FeatureContentRecord) {
  return <h1>{content.__typename}</h1>
}

function renderHeaderContentRecord(content: HeaderContentRecord) {
  // NB
  return (
    <div className='flex w-full flex-col items-end space-y-20 px-4 py-20 lg:flex-row lg:flex-row-reverse lg:items-center lg:justify-between lg:space-y-0 lg:px-0 lg:py-20'>
      <img className='block h-64 lg:-mr-20' src={content.shapes?.url} />
      <h2 className='font-display text-5xl font-medium lg:mr-6'>
        {content.title}
      </h2>
    </div>
  )
}

function renderHeroContentRecord(content: HeroContentRecord) {
  return <h1>{content.__typename}</h1>
}

function renderHomeHeroContentRecord(content: HomeHeroContentRecord) {
  return <h1>{content.__typename}</h1>
}

function renderShowcaseContentRecord(content: ShowcaseContentRecord) {
  return <h1>{content.__typename}</h1>
}

function renderStoryContentRecord(content: StoryContentRecord) {
  return <h1>{content.__typename}</h1>
}

function renderTeamContentRecord(content: TeamContentRecord) {
  return <h1>{content.__typename}</h1>
}

function renderTextContentRecord(content: TextContentRecord) {
  return (
    <div className={clsx(content.textCentered) && 'text-center'}>
      <p className='text-2xl text-medium'>{content.body}</p>
    </div>
  )
}
