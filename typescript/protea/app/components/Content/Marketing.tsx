import type { ReactNode } from 'react'
import { Fragment } from 'react'
import { Shape } from '~/components'
import type { SectionRecord } from '~/generated/dato-cms-graphql'
import {
  CtaContentRecordComponent,
  FeatureBlocksContentRecordComponent,
  FeatureContentRecordComponent,
  HeaderContentRecordComponent,
  HeroContentRecordComponent,
  HomeHeroContentRecordComponent,
  ShowcaseContentRecordComponent,
  StoryContentRecordComponent,
  TeamContentRecordComponent,
  TextContentRecordComponent
} from './Blocks'

type MarketingPageWithSectionsProps = {
  section: SectionRecord
  children?: ReactNode
}

export function MarketingPageWithSections({
  section,
  children
}: MarketingPageWithSectionsProps) {
  return (
    <div
      key={section.id + 'MarketingPageWithSections'}
      className='group w-full overflow-hidden even:rounded-2xl even:bg-mk-section'
    >
      <div className='flex w-full group-first:hidden group-even:hidden'>
        <Shape radius='rounded-l-full' color='bg-mk-section' width='w-20' />
        <Shape radius='rounded-full' color='bg-mk-section' width='w-20' />
        <Shape radius='rounded-full' color='bg-transparent' width='w-20' />
        <Shape radius='rounded-full' color='bg-mk-section' width='w-20' />
      </div>
      <div
        className={`relative mx-auto flex w-full max-w-[59rem] flex-col items-center py-20 ${
          !children ? 'pb-32 pt-32' : ''
        }`}
      >
        {section.content.map((content, index) => {
          switch (content.__typename) {
            case 'CtaContentRecord':
              return (
                <CtaContentRecordComponent
                  key={content.__typename + index}
                  content={content}
                />
              )
            case 'FeatureBlocksContentRecord':
              return (
                <FeatureBlocksContentRecordComponent
                  key={content.__typename + index}
                  content={content}
                />
              )
            case 'FeatureContentRecord':
              return (
                <FeatureContentRecordComponent
                  key={content.__typename + index}
                  content={content}
                />
              )
            case 'HeaderContentRecord':
              return (
                <HeaderContentRecordComponent
                  key={content.__typename + index}
                  content={content}
                />
              )
            case 'HeroContentRecord':
              return (
                <HeroContentRecordComponent
                  key={content.__typename + index}
                  content={content}
                />
              )
            case 'HomeHeroContentRecord':
              return (
                <HomeHeroContentRecordComponent
                  key={content.__typename + index}
                  content={content}
                />
              )
            case 'ShowcaseContentRecord':
              return (
                <ShowcaseContentRecordComponent
                  key={content.__typename + index}
                  content={content}
                />
              )
            case 'StoryContentRecord':
              return (
                <StoryContentRecordComponent
                  key={content.__typename + index}
                  content={content}
                />
              )
            case 'TeamContentRecord':
              return (
                <TeamContentRecordComponent
                  key={content.__typename + index}
                  content={content}
                />
              )
            case 'TextContentRecord':
              return (
                <TextContentRecordComponent
                  key={content.__typename + index}
                  content={content}
                />
              )
            default:
              return <Fragment key={'children' + index}>{children}</Fragment>
          }
        })}
      </div>
      <div className='flex w-full justify-end group-first:hidden group-even:hidden'>
        <Shape radius='rounded-tl-full' color='bg-mk-section' width='w-20' />
        <Shape radius='rounded-tr-full' color='bg-mk-section' width='w-20' />
        <Shape radius='rounded-br-full' color='bg-mk-section' width='w-20' />
        <Shape radius='rounded-tr-full' color='bg-mk-section' width='w-20' />
      </div>
    </div>
  )
}
