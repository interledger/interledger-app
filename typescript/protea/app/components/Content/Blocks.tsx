import { Tab } from '@headlessui/react'
import clsx from 'clsx'
import type { MotionProps } from 'framer-motion'
import { AnimatePresence, motion, useAnimate } from 'framer-motion'
import { Fragment, useCallback, useEffect, useRef, useState } from 'react'
import type { ResponsiveImageType } from 'react-datocms'
import { Image, StructuredText } from 'react-datocms'
import { useParams } from 'react-router'
import type { SelectOptions } from '~/components'
import {
  AnchorRouter,
  Card,
  CardContent,
  FynbosIcon,
  LinkedInIcon,
  Select,
  TextArea,
  TextField,
  TwitterIcon
} from '~/components'
import { ContentRouter, Prose } from '~/components/Content'
import type {
  CtaContentRecord,
  FeatureBlocksContentRecord,
  FeatureContentRecord,
  FormLandingRecord,
  FormLandingStepRecord,
  FormSectionRecord,
  FormSelectRecord,
  FormTextRecord,
  HeaderContentRecord,
  HeroContentRecord,
  HomeHeroContentRecord,
  ShowcaseContentRecord,
  StoryContentRecord,
  TeamContentRecord,
  TextContentRecord
} from '~/generated/dato-cms-graphql'
import { renderLinkNodeRule } from './renderNodeRules'

export function CtaContentRecordComponent({
  content
}: {
  content: CtaContentRecord
}) {
  return (
    <div
      key={content.id}
      className='flex w-full flex-col space-y-16 px-4 py-10 lg:flex-row lg:space-x-16 lg:space-y-0 lg:px-0 lg:py-16'
    >
      <div className='flex items-start'>
        <AnimatePresence mode='wait'>
          <motion.img
            key={content.image?.url + 'image'}
            src={content.image?.url}
            height={content.image?.height}
            width={content.image?.width}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, transition: { duration: 0.15 } }}
            exit={{ opacity: 0 }}
            className='block dark:hidden'
          />
          <motion.img
            key={content.imageDark?.url + 'imageDark'}
            src={content.imageDark?.url}
            height={content.imageDark?.height}
            width={content.imageDark?.width}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, transition: { duration: 0.15 } }}
            exit={{ opacity: 0 }}
            className='hidden dark:block'
          />
        </AnimatePresence>
      </div>
      <div className='flex flex-col space-y-6'>
        <h2 className='text-4xl font-medium'>{content.title}</h2>
        <p className='text-2xl text-medium'>{content.body}</p>
        {content.button.length > 0 && (
          <ContentRouter shrink to={content.button[0]} />
        )}
      </div>
    </div>
  )
}

export function FeatureBlocksContentRecordComponent({
  content
}: {
  content: FeatureBlocksContentRecord
}) {
  return (
    <div
      key={content.id}
      className='w-full columns-1 gap-10 space-y-10 px-4 lg:columns-2 lg:px-0'
    >
      <AnimatePresence>
        {content.blocks.map((block) => (
          <motion.img
            key={block.image?.url}
            src={block.image?.url}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, transition: { duration: 0.15 } }}
            exit={{ opacity: 0 }}
            className='w-full'
          />
        ))}
      </AnimatePresence>
    </div>
  )
}

export function FeatureContentRecordComponent({
  content
}: {
  content: FeatureContentRecord
}) {
  return (
    <div
      key={content.id}
      className={clsx(
        'mt-20 flex w-full flex-col justify-between px-4 first-of-type:mt-0 lg:mt-40 lg:flex-row lg:gap-x-20 lg:px-0',
        content.rowReverse && 'lg:flex-row-reverse'
      )}
    >
      <div className='flex w-full flex-col justify-center gap-y-6'>
        <h2 className='font-display text-3xl font-medium'>{content.title}</h2>
        <p className='text-xl text-medium'>{content.body}</p>
      </div>
      <img
        alt='Feature'
        className={clsx(
          'hidden lg:-mr-[10.5rem] lg:block',
          content.rowReverse && 'lg:-ml-[10.5rem] lg:mr-0'
        )}
        src={content.image?.url}
        height={content.image?.height}
        width={content.image?.width}
      />
      <img
        alt='Feature'
        className='mt-6 block lg:hidden'
        src={content.imageMobile?.url}
        height={content.imageMobile?.height}
        width={content.imageMobile?.width}
      />
    </div>
  )
}

export function HeaderContentRecordComponent({
  content
}: {
  content: HeaderContentRecord
}) {
  return (
    <div
      key={content.id}
      className='flex w-full flex-col items-end space-y-20 px-4 lg:flex-row-reverse lg:items-center lg:justify-between lg:space-y-0 lg:px-0'
    >
      {content.shapes && (
        <img
          alt='shapes'
          className='block h-64 lg:-mr-20'
          src={content.shapes?.url}
          height={content.shapes?.height}
          width={content.shapes?.width}
        />
      )}
      {content.title && (
        <h2 className='font-display text-4xl font-medium lg:mr-6 lg:text-5xl'>
          {content.title}
        </h2>
      )}
    </div>
  )
}

export function HeroContentRecordComponent({
  content
}: {
  content: HeroContentRecord
}) {
  return (
    <div
      key={content.id}
      className='flex w-full flex-col lg:-my-20 lg:flex-row lg:gap-x-12'
    >
      {/* <AnimatePresence mode='wait'>
        <motion.img
          alt='Hero image'
          key={content.imageMobile?.url + 'imageMobile'}
          src={content.imageMobile?.url}
          height={content.imageMobile?.height}
          width={content.imageMobile?.width}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='-mt-20 block dark:hidden lg:hidden'
        />
        <motion.img
          alt='Hero image'
          key={content.imageDarkMobile?.url + 'imageDarkMobile'}
          src={content.imageDarkMobile?.url}
          height={content.imageDarkMobile?.height}
          width={content.imageDarkMobile?.width}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='-mt-20 hidden dark:block lg:hidden'
        />
      </AnimatePresence> */}
      <div className='flex w-full flex-col justify-center space-y-8 px-4 text-center lg:px-0 lg:text-start'>
        <h1 className='mt-10 font-display text-5xl font-medium lg:mt-0 lg:text-6xl'>
          {content.title}
        </h1>
        <p className='text text-lg text-medium lg:text-2xl'>{content.body}</p>
        {content.button.length > 0 && (
          <ContentRouter shrink to={content.button[0]} />
        )}
      </div>
      {/* <AnimatePresence mode='wait'>
        <motion.img
          alt='Hero image'
          key={content.image?.url + 'image'}
          src={content.image?.url}
          height={content.imageDark?.height}
          width={content.imageDark?.width}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='-mr-[10.5rem] hidden lg:block lg:dark:hidden'
        />
        <motion.img
          alt='Hero image'
          key={content.imageDark?.url + 'imageDark'}
          src={content.imageDark?.url}
          height={content.imageDark?.height}
          width={content.imageDark?.width}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='-mr-[10.5rem] hidden lg:dark:block'
        />
      </AnimatePresence> */}
    </div>
  )
}

/**
 * Start of HomeHeroContentRecordComponent
 */

type Segment = {
  text: string
  animated: boolean
  index: number
}

function getSegments(
  title: string,
  iterations: HeroContentRecord[],
  type: 'title' | 'body'
): Segment[] {
  const segments: Segment[] = []

  let lastIndex = 0,
    count = 0
  for (const iteration of iterations) {
    const start = title?.indexOf(iteration[type] as string)
    const end = start + (iteration[type] as string).length

    if (start == -1) continue

    if (start > lastIndex) {
      segments.push({
        animated: false,
        index: -1,
        text: title.slice(lastIndex, start)
      })
    }

    segments.push({
      animated: true,
      index: count,
      text: iteration[type] as string
    })
    count++
    lastIndex = end
  }

  return segments
}

interface SegmentProps extends MotionProps {
  active: number
  index: number
}

function TitleSegment({ active, index, children, ...rest }: SegmentProps) {
  const [scope, animate] = useAnimate()

  const activeColors = [
    'var(--title-one)',
    'var(--title-two)',
    'var(--title-three)',
    'var(--title-four)'
  ]
  const activeColor = activeColors[index]

  useEffect(() => {
    animate(
      scope.current,
      { color: active == index ? activeColor : 'var(--title-default)' },
      { duration: 0.5 }
    )
  }, [active, activeColor, animate, index, scope])

  return (
    <motion.span ref={scope} {...rest}>
      {children}
    </motion.span>
  )
}

function BodySegment({ active, index, children, ...rest }: SegmentProps) {
  const [scope, animate] = useAnimate()

  const activeColors = [
    'var(--background-one)',
    'var(--background-two)',
    'var(--background-three)',
    'var(--background-four)'
  ]
  const activeColor = activeColors[index]

  useEffect(() => {
    animate(
      scope.current,
      {
        backgroundColor:
          active == index ? activeColor : 'var(--background-default)'
      },
      { duration: 0.5 }
    )
  }, [active, activeColor, animate, index, scope])

  return (
    <motion.span
      ref={scope}
      {...rest}
      className='-mx-1 -my-1 rounded-lg box-decoration-clone px-1 py-1 selection:bg-transparent'
    >
      {children}
    </motion.span>
  )
}

export function HomeHeroContentRecordComponent({
  content
}: {
  content: HomeHeroContentRecord
}) {
  const [active, setActive] = useState<number>(0)

  const titleSegments = getSegments(
    content.title as string,
    content.iterations,
    'title'
  )
  const bodySegments = getSegments(
    content.body as string,
    content.iterations,
    'body'
  )

  useEffect(() => {
    const interval: NodeJS.Timeout = setInterval(() => {
      if (document.visibilityState === 'visible') {
        setActive((active + 1) % content.iterations.length)
      }
    }, 3000)
    return () => clearInterval(interval)
  })

  return (
    <div
      key={content.id}
      className='flex flex-col space-y-10 px-4 pb-10 text-center lg:space-y-0 lg:py-16 lg:text-left'
    >
      <AnimatePresence mode='wait'>
        <motion.img
          alt='Hero image'
          key={content.iterations[active]?.mobileShape?.url + 'shapeMobile'}
          src={content.iterations[active]?.mobileShape?.url}
          height={content.iterations[active]?.mobileShape?.height}
          width={content.iterations[active]?.mobileShape?.width}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.5 } }}
          exit={{ opacity: 0 }}
          className='-mr-4 -mt-20 block h-44 w-44 self-end lg:hidden'
        />
      </AnimatePresence>
      <div className='flex w-full flex-col space-y-8 lg:w-1/2'>
        <div className='font-display text-5xl font-medium'>
          {titleSegments.map((segment, index) => {
            if (segment.animated) {
              return (
                <TitleSegment
                  key={segment.text + 'active' + index}
                  active={active}
                  index={segment.index}
                >
                  {segment.text}
                </TitleSegment>
              )
            }
            return (
              <span key={segment.text + 'inactive' + index}>
                {segment.text}
              </span>
            )
          })}
        </div>
        <div className='text-xl text-medium'>
          {bodySegments.map((segment, index) => {
            if (segment.animated) {
              return (
                <BodySegment
                  key={segment.text + 'active' + index}
                  active={active}
                  index={segment.index}
                >
                  {segment.text}
                </BodySegment>
              )
            }
            return (
              <span key={segment.text + 'inactive' + index}>
                {segment.text}
              </span>
            )
          })}
        </div>
        {content.button.length > 0 && (
          <ContentRouter shrink to={content.button[0]} />
        )}
      </div>
      <AnimatePresence>
        <motion.img
          alt='Hero image'
          key={content.iterations[active]?.image?.url + 'image'}
          src={content.iterations[active]?.image?.url}
          height={content.iterations[active]?.image?.height}
          width={content.iterations[active]?.image?.width}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.5 } }}
          exit={{ opacity: 0 }}
          className='absolute -right-[10.5rem] top-0 hidden lg:block lg:dark:hidden'
        />
        <motion.img
          alt='Hero image'
          key={content.iterations[active]?.imageDark?.url + 'imageDark'}
          src={content.iterations[active]?.imageDark?.url}
          height={content.iterations[active]?.imageDark?.height}
          width={content.iterations[active]?.imageDark?.width}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.5 } }}
          exit={{ opacity: 0 }}
          className='absolute -right-[10.5rem] top-0 hidden lg:dark:block'
        />
      </AnimatePresence>
    </div>
  )
}

export function ShowCaseDesktop({
  content
}: {
  content: ShowcaseContentRecord
}) {
  let [selectedIndex, setSelectedIndex] = useState(0)

  let onChange = useCallback(
    (selectedIndex: number) => {
      setSelectedIndex(selectedIndex)
    },
    [setSelectedIndex]
  )

  return (
    <Tab.Group
      as='div'
      className='hidden grid-cols-12 items-center gap-6 sm:grid'
      selectedIndex={selectedIndex}
      onChange={onChange}
      vertical
    >
      <Tab.List
        className={clsx(
          !content.rowReverse && 'order-last',
          'relative z-10 col-span-7 space-y-6'
        )}
      >
        {content.cases.map((feature, featureIndex) => (
          <div
            key={feature.id}
            className='relative rounded-2xl transition-colors hover:bg-mk-section-hover/50'
          >
            {featureIndex === selectedIndex && (
              <motion.div
                layoutId={'activeBackground' + content.id}
                className='absolute inset-0 rounded-2xl bg-mk-section-hover'
              />
            )}
            <Tab className='relative z-10 rounded-2xl p-6 text-left focus:outline-none'>
              <h3 className='font-display text-2xl font-medium'>
                {feature.title}
              </h3>
              <p className='mt-4 text-lg text-medium'>{feature.body}</p>
            </Tab>
          </div>
        ))}
      </Tab.List>
      <div className='relative col-span-5 h-[42.1875rem]'>
        <div className='z-10 mx-auto w-full'>
          <Tab.Panels as={Fragment}>
            <AnimatePresence initial={false} mode='wait'>
              {content.cases.map((feature, featureIndex) =>
                selectedIndex === featureIndex ? (
                  <Tab.Panel
                    static
                    key={feature.image?.url}
                    className='flex focus:outline-offset-[32px] [&:not(:focus-visible)]:focus:outline-none'
                  >
                    <motion.img
                      alt={feature.title as string}
                      key={feature.image?.url + 'image'}
                      src={feature.image?.url}
                      height={feature.image?.height}
                      width={feature.image?.width}
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1, transition: { duration: 0.15 } }}
                      exit={{ opacity: 0 }}
                      className='block dark:hidden'
                    />
                    <motion.img
                      alt={feature.title as string}
                      key={feature.imageDark?.url + 'imageDark'}
                      src={feature.imageDark?.url}
                      height={feature.imageDark?.height}
                      width={feature.imageDark?.width}
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1, transition: { duration: 0.15 } }}
                      exit={{ opacity: 0 }}
                      className='hidden dark:block'
                    />
                  </Tab.Panel>
                ) : null
              )}
            </AnimatePresence>
          </Tab.Panels>
        </div>
      </div>
    </Tab.Group>
  )
}

export function ShowCaseMobile({
  content
}: {
  content: ShowcaseContentRecord
}) {
  let [activeIndex, setActiveIndex] = useState(0)
  let slideContainerRef = useRef<HTMLDivElement>(null)
  let slideRefs = useRef<(HTMLDivElement | null)[]>([])

  useEffect(() => {
    let observer = new window.IntersectionObserver(
      (entries) => {
        for (let entry of entries) {
          if (entry.isIntersecting) {
            setActiveIndex(
              slideRefs.current.indexOf(entry.target as HTMLDivElement)
            )
            break
          }
        }
      },
      {
        root: slideContainerRef.current,
        threshold: 0.6
      }
    )

    for (let slide of slideRefs.current) {
      if (slide) {
        observer.observe(slide)
      }
    }

    return () => {
      observer.disconnect()
    }
  }, [slideContainerRef, slideRefs])

  return (
    <>
      <div
        ref={slideContainerRef}
        className='flex snap-x snap-mandatory -space-x-4 overflow-x-auto overscroll-x-contain scroll-smooth pb-4 [scrollbar-width:none] sm:hidden sm:-space-x-6 [&::-webkit-scrollbar]:hidden'
      >
        {content.cases.map((feature, featureIndex) => (
          <div
            key={featureIndex}
            ref={(ref) => (slideRefs.current[featureIndex] = ref)}
            className='w-full flex-none snap-center px-4 sm:px-6'
          >
            <div className='relative transform'>
              <div className='relative mx-auto w-full'>
                <motion.img
                  alt={feature.title as string}
                  key={feature.image?.url + 'image'}
                  src={feature.image?.url}
                  height={feature.image?.height}
                  width={feature.image?.width}
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1, transition: { duration: 0.15 } }}
                  exit={{ opacity: 0 }}
                  className='block w-full dark:hidden'
                />
                <motion.img
                  alt={feature.title as string}
                  key={feature.imageDark?.url + 'imageDark'}
                  src={feature.imageDark?.url}
                  height={feature.imageDark?.height}
                  width={feature.imageDark?.width}
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1, transition: { duration: 0.15 } }}
                  exit={{ opacity: 0 }}
                  className='hidden w-full dark:block'
                />
              </div>
              <div className='absolute inset-x-0 bottom-0 rounded-b-2xl bg-mk-section-hover p-6 backdrop-blur sm:p-10'>
                <h3 className='font-display text-2xl font-medium'>
                  {feature.title}
                </h3>
                <p className='mt-2 text-lg text-medium'>{feature.body}</p>
              </div>
            </div>
          </div>
        ))}
      </div>
      <div className='mt-6 flex justify-center gap-3 sm:hidden'>
        {content.cases.map((_, featureIndex) => (
          <button
            type='button'
            key={featureIndex}
            className={clsx(
              'relative h-0.5 w-8 rounded-full',
              featureIndex === activeIndex
                ? 'bg-slate-950 dark:bg-slate-300'
                : 'bg-slate-300 dark:bg-slate-700'
            )}
            aria-label={`Go to slide ${featureIndex + 1}`}
            onClick={() => {
              slideRefs.current[featureIndex]?.scrollIntoView({
                block: 'nearest',
                inline: 'nearest'
              })
            }}
          >
            <span className='absolute -inset-x-1.5 -inset-y-3' />
          </button>
        ))}
      </div>
    </>
  )
}

export function ShowcaseContentRecordComponent({
  content
}: {
  content: ShowcaseContentRecord
}) {
  return (
    <div key={content.id} className='mt-40 last-of-type:mt-0'>
      <ShowCaseDesktop content={content} />
      <ShowCaseMobile content={content} />
    </div>
  )
}

export function StoryContentRecordComponent({
  content
}: {
  content: StoryContentRecord
}) {
  return content.blurb?.value.document.children[0].children[0].value != '' ? (
    <div
      key={content.id}
      className='grid w-full grid-cols-12 gap-x-6 gap-y-10 px-4 lg:gap-y-14 lg:px-0'
    >
      <div className='col-span-full flex items-start lg:col-span-4'>
        <AnimatePresence mode='wait'>
          <motion.img
            alt='shapes'
            key={content.image?.url + 'image'}
            src={content.image?.url}
            height={content.image?.height}
            width={content.image?.width}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, transition: { duration: 0.15 } }}
            exit={{ opacity: 0 }}
            className='block dark:hidden'
          />
          <motion.img
            alt='shapes'
            key={content.imageDark?.url + 'imageDark'}
            src={content.imageDark?.url}
            height={content.imageDark?.height}
            width={content.imageDark?.width}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, transition: { duration: 0.15 } }}
            exit={{ opacity: 0 }}
            className='hidden dark:block'
          />
        </AnimatePresence>
      </div>
      <div className='col-span-full flex w-full items-center lg:col-span-8'>
        <h2 className='text-4xl font-medium'>{content.title}</h2>
      </div>
      <div className='col-span-full w-full items-start lg:col-span-4'>
        <div className='text-2xl text-medium'>
          <Prose className='max-w-none prose-p:text-lg prose-p:font-medium prose-p:lg:text-2xl'>
            <StructuredText
              data={content.blurb?.value}
              customNodeRules={[renderLinkNodeRule]}
            />
          </Prose>
        </div>
      </div>
      <div className='col-span-full lg:col-span-8'>
        <div className='text-2xl text-medium'>
          <Prose className='max-w-none prose-p:text-lg prose-p:lg:text-2xl'>
            <StructuredText
              data={content.bodyText?.value}
              customNodeRules={[renderLinkNodeRule]}
            />
          </Prose>
        </div>
      </div>
    </div>
  ) : (
    <div
      key={content.id}
      className='grid w-full grid-cols-12 gap-x-6 gap-y-10 px-4 lg:gap-y-14 lg:px-0'
    >
      <div className='col-span-full flex items-start lg:col-span-4'>
        <AnimatePresence mode='wait'>
          <motion.img
            alt='shapes'
            key={content.image?.url + 'image'}
            src={content.image?.url}
            height={content.image?.height}
            width={content.image?.width}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, transition: { duration: 0.15 } }}
            exit={{ opacity: 0 }}
            className='block dark:hidden'
          />
          <motion.img
            alt='shapes'
            key={content.imageDark?.url + 'imageDark'}
            src={content.imageDark?.url}
            height={content.imageDark?.height}
            width={content.imageDark?.width}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, transition: { duration: 0.15 } }}
            exit={{ opacity: 0 }}
            className='hidden dark:block'
          />
        </AnimatePresence>
      </div>
      <div className='col-span-full flex w-full flex-col items-center space-y-14 lg:col-span-8'>
        <div className='flex w-full items-center'>
          <h2 className='text-4xl font-medium'>{content.title}</h2>
        </div>
        <div>
          <div className='text-2xl text-medium'>
            <Prose className='max-w-none prose-p:text-lg prose-p:lg:text-2xl'>
              <StructuredText
                data={content.bodyText?.value}
                customNodeRules={[renderLinkNodeRule]}
              />
            </Prose>
          </div>
        </div>
      </div>
    </div>
  )
}

export function TeamContentRecordComponent({
  content
}: {
  content: TeamContentRecord
}) {
  return (
    <div
      key={content.id}
      className='grid w-full grid-cols-4 gap-y-20 px-4 sm:grid-cols-8 lg:grid-cols-12 lg:gap-x-10 lg:gap-y-14 lg:px-0'
    >
      <div className='col-span-full flex flex-col items-center space-y-10 text-center'>
        <AnimatePresence mode='wait'>
          <motion.img
            alt='shapes'
            key={content.image?.url + 'image'}
            src={content.image?.url}
            height={content.image?.height}
            width={content.image?.width}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, transition: { duration: 0.15 } }}
            exit={{ opacity: 0 }}
            className='block dark:hidden'
          />
          <motion.img
            alt='shapes'
            key={content.imageDark?.url + 'imageDark'}
            src={content.imageDark?.url}
            height={content.imageDark?.height}
            width={content.imageDark?.width}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, transition: { duration: 0.15 } }}
            exit={{ opacity: 0 }}
            className='hidden dark:block'
          />
        </AnimatePresence>
        {content.title && (
          <h2 className='font-display text-4xl font-medium'>{content.title}</h2>
        )}
      </div>
      {content.people.map((member) => (
        <div
          key={member.id}
          className='col-span-full flex flex-col items-center sm:col-span-4'
        >
          <Image
            pictureClassName='m-0'
            className='block w-full rounded-2xl'
            data={{
              ...(member.person?.avatar?.responsiveImage as ResponsiveImageType)
            }}
          />
          <h3 className='mt-10 font-display text-2xl font-medium'>
            {member.person?.name}
          </h3>
          <h3 className='mt-3 text-lg text-medium'>{member.person?.role}</h3>
          <div className='mt-3 flex space-x-4'>
            {member.person?.twitterUrl && (
              <AnchorRouter to={member.person.twitterUrl}>
                <TwitterIcon />
              </AnchorRouter>
            )}
            {member.person?.linkedinUrl && (
              <AnchorRouter to={member.person.linkedinUrl}>
                <LinkedInIcon />
              </AnchorRouter>
            )}
            {member.person?.fynbosUrl && (
              <AnchorRouter to={member.person.fynbosUrl}>
                <FynbosIcon />
              </AnchorRouter>
            )}
          </div>
        </div>
      ))}
    </div>
  )
}

export function TextContentRecordComponent({
  content
}: {
  content: TextContentRecord
}) {
  return (
    <div
      key={content.id}
      className={clsx(
        content.textCentered && 'text-center',
        'flex w-full flex-col items-center space-y-6 px-4 lg:px-0'
      )}
    >
      {content.image && (
        <AnimatePresence>
          <motion.img
            alt='Image'
            key={content.image?.url + 'image'}
            src={content.image?.url}
            height={content.image?.height}
            width={content.image?.width}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, transition: { duration: 0.15 } }}
            exit={{ opacity: 0 }}
          />
        </AnimatePresence>
      )}
      {content.title && (
        <h2 className='pt-4 font-display text-4xl font-medium'>
          {content.title}
        </h2>
      )}
      {content.bodyText && (
        <Prose
          className={clsx(
            'w-full max-w-none',
            !content.textStandard && 'prose-p:text-lg prose-p:lg:text-2xl'
          )}
        >
          <StructuredText
            data={content.bodyText.value}
            customNodeRules={[renderLinkNodeRule]}
          />
        </Prose>
      )}
      {content.button.length > 0 && (
        <div className='flex flex-col items-center space-y-4 md:flex-row md:space-x-4 md:space-y-0'>
          {content.button.map((button, index) => (
            <ContentRouter
              key={button.id}
              shrink
              outline={index == 0} // Max two buttons :shrug:
              to={button}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export function FormSectionRecordComponent({
  content
}: {
  content: FormSectionRecord
}) {
  return (
    <Card key={content.id} className={clsx('')}>
      {content.description && (
        <CardContent>
          <p className='text-medium'>{content.description}</p>
        </CardContent>
      )}
      {content.content.map((content, index) => {
        switch (content.__typename) {
          case 'FormLandingRecord':
            return (
              <FormLandingRecordComponent
                key={content.__typename + index}
                content={content}
              />
            )
          case 'FormTextRecord':
            return (
              <FormTextRecordComponent
                key={content.__typename + index}
                content={content}
              />
            )
          case 'FormSelectRecord':
            return (
              <FormSelectRecordComponent
                key={content.__typename + index}
                content={content}
              />
            )
          default:
            return (
              <Fragment key={'children' + index}>
                Tried rendering a form field that doesn't exist
              </Fragment>
            )
        }
      })}
    </Card>
  )
}

export function FormLandingRecordComponent({
  content
}: {
  content: FormLandingRecord
}) {
  return (
    <CardContent key={content.id}>
      {content.steps.map((step, index) => {
        switch (step.__typename) {
          case 'FormLandingStepRecord':
            return (
              <FormLandingStepRecordComponent
                key={step.__typename + index}
                content={step}
              />
            )
          default:
            return (
              <Fragment key={'children' + index}>
                Tried rendering a form step that doesn't exist
              </Fragment>
            )
        }
      })}
    </CardContent>
  )
}

export function FormLandingStepRecordComponent({
  content
}: {
  content: FormLandingStepRecord
}) {
  return (
    <div key={content.id} className='mt-10 flex items-start first-of-type:mt-2'>
      <img
        alt='Decorative shapes'
        className='h-8 w-16 flex-none'
        src={content.shapes?.url}
      />
      <div className='ml-5'>
        <h3 className='mb-1 font-medium text-strong'>{content.title}</h3>
        <p className='text-xs text-medium'>{content.body}</p>
      </div>
    </div>
  )
}

export function FormTextRecordComponent({
  content
}: {
  content: FormTextRecord
}) {
  const { slug } = useParams()
  return content.fieldType == 'area' ? (
    <TextArea
      key={content.id}
      id={content.id}
      form={`dynamic-${slug}`}
      label={content.label || ''}
      name={content.fieldName || content.id}
      required={content.required}
      className='mt-4'
    />
  ) : (
    <TextField
      key={content.id}
      id={content.id}
      form={`dynamic-${slug}`}
      label={content.label || ''}
      name={content.fieldName || content.id}
      required={content.required}
      type={content.fieldType as 'text' | 'number'}
      className='mt-4'
    />
  )
}

export function FormSelectRecordComponent({
  content
}: {
  content: FormSelectRecord
}) {
  const options =
    content.options?.split(',').map((option) => ({
      id: option,
      name: option
    })) || []

  const { slug } = useParams()
  const [value, setValue] = useState<SelectOptions>(options[0])

  const _onChangeSelect = useCallback((event: SelectOptions) => {
    setValue(event)
  }, [])

  return (
    <>
      <Select
        id={content.id}
        label={content.label || ''}
        key={content.id}
        value={value}
        options={options}
        onChange={_onChangeSelect}
      />
      <input
        type='hidden'
        name={content.fieldName || content.id}
        value={value?.name}
        form={`dynamic-${slug}`}
      />
    </>
  )
}
