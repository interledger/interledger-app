import { Tab } from '@headlessui/react'
import clsx from 'clsx'
import { AnimatePresence, motion, useAnimate } from 'framer-motion'
import {
  Fragment,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState
} from 'react'
import type { ResponsiveImageType } from 'react-datocms'
import { Image, StructuredText } from 'react-datocms'
import {
  AnchorRouter,
  FynbosIcon,
  LinkedInIcon,
  TwitterIcon
} from '~/components'
import { Prose } from '~/components/Content'
import type {
  CtaContentRecord,
  FeatureBlocksContentRecord,
  FeatureContentRecord,
  HeaderContentRecord,
  HeroContentRecord,
  HomeHeroContentRecord,
  ShowcaseContentRecord,
  StoryContentRecord,
  TeamContentRecord,
  TextContentRecord
} from '~/generated/dato-cms-graphql'
import { ContentRouter } from './ContentRouter'
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
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, transition: { duration: 0.15 } }}
            exit={{ opacity: 0 }}
            className='block dark:hidden'
          />
          <motion.img
            key={content.imageDark?.url + 'imageDark'}
            src={content.imageDark?.url}
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
      />
      <img
        alt='Feature'
        className='mt-6 block lg:hidden'
        src={content.imageMobile?.url}
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
          alt='Fynbos shapes'
          className='block h-64 lg:-mr-20'
          src={content.shapes?.url}
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
      <AnimatePresence mode='wait'>
        <motion.img
          alt='Hero image'
          key={content.imageMobile?.url + 'imageMobile'}
          src={content.imageMobile?.url}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='-mt-20 block dark:hidden lg:hidden'
        />
        <motion.img
          alt='Hero image'
          key={content.imageDarkMobile?.url + 'imageDarkMobile'}
          src={content.imageDarkMobile?.url}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='-mt-20 hidden dark:block lg:hidden'
        />
      </AnimatePresence>
      <div className='flex w-full flex-col justify-center space-y-8 px-4 text-center lg:px-0 lg:text-start'>
        <h1 className='mt-10 font-display text-5xl font-medium lg:mt-0 lg:text-6xl'>
          {content.title}
        </h1>
        <p className='text text-lg text-medium lg:text-2xl'>{content.body}</p>
      </div>
      <AnimatePresence mode='wait'>
        <motion.img
          alt='Hero image'
          key={content.image?.url + 'image'}
          src={content.image?.url}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='-mr-[10.5rem] hidden lg:block lg:dark:hidden'
        />
        <motion.img
          alt='Hero image'
          key={content.imageDark?.url + 'imageDark'}
          src={content.imageDark?.url}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='-mr-[10.5rem] hidden lg:dark:block'
        />
      </AnimatePresence>
    </div>
  )
}

// type Segment = {
//   val: string
//   active: boolean
// }
//
// function getSegments(str: string, search: string): Segment[] {
//   const index = str.indexOf(search)
//   const length = search.length
//   return [
//     { val: index > 0 ? str.slice(0, index) + ' ' : '', active: false },
//     { val: str.slice(index, index + length), active: true },
//     { val: ' ' + str.slice(index + length), active: false }
//   ]
// }

type Segment = {
  val: string
  animated: boolean
}

function getSegments(str: string, iterations: HeroContentRecord[]): Segment[] {
  console.log('content', JSON.stringify(iterations))

  // TODO pickup here
  const segments: { start: number; end: number; text: string }[] = []
  iterations?.forEach((iteration, i) => {
    const start = str?.indexOf(iteration.title as string) ?? 0
    segments.push({
      start: start,
      end: start + (iteration.title as string).length,
      text: iteration.title as string
    })
  })
  console.log('segments', segments)
  return [{ val: 'test', animated: false }]
}

export function HomeHeroTitle({
  content,
  isDark
}: {
  content: HomeHeroContentRecord
  isDark: boolean
}) {
  console.log('content', JSON.stringify(content))

  const segments: { start: number; end: number; text: string }[] = []
  content.iterations?.forEach((iteration, i) => {
    const start = content.title?.indexOf(iteration.title as string) ?? 0
    segments.push({
      start: start,
      end: start + (iteration.title as string).length,
      text: iteration.title as string
    })
  })
  console.log('segments', segments)

  /**
   * content {
   *   id: '122272899',
   *   title: 'Connect. Verify. Transact with certainty.',
   *   iterations: [
   *     {
   *       id: '122272895',
   *       title: 'Connect.'
   *     },
   *     {
   *       id: '122272896',
   *       title: 'Verify.'
   *     },
   *     {
   *       id: '122272897',
   *       title: 'Transact'
   *     },
   *     {
   *       id: '122272898',
   *       title: 'certainty.'
   *     }
   *   ],
   * }
   */
}

function TitleSegment({
  active,
  index,
  children
}: {
  active: number
  index: number
  children: string
}) {
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
      { duration: 0.6 }
    )
  }, [active, activeColor, animate, index, scope])

  return <motion.span ref={scope}>{children}</motion.span>
}

export function HomeHeroContentRecordComponent({
  content
}: {
  content: HomeHeroContentRecord
}) {
  console.log('content', content)
  const [active, setActive] = useState<number>(0)

  const titleSegments = useMemo(() => {
    return getSegments(content.title as string, content.iterations)
  }, [active, content.iterations, content.title])

  const bodySegments = useMemo(() => {
    return getSegments(content.body as string, content.iterations)
  }, [active, content.body, content.iterations])

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
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='-mr-4 -mt-20 block h-44 w-44 self-end lg:hidden'
        />
      </AnimatePresence>
      <div className='flex w-full flex-col space-y-8 lg:w-1/2'>
        <div className='font-display text-5xl font-bold'>
          <TitleSegment active={active} index={2}>
            This is the first index
          </TitleSegment>

          {/*<AnimatePresence mode='wait'>*/}
          {/*  {titleSegments.map((segment, index) => {*/}
          {/*    if (segment.active) {*/}
          {/*      return (*/}
          {/*        <motion.span*/}
          {/*          key={segment.val + 'active' + index}*/}
          {/*          initial={{ color: hightlightTitleDefault[currentTheme] }}*/}
          {/*          animate={{*/}
          {/*            color: highlightTitle[active],*/}
          {/*            transition: { duration: 0.15 }*/}
          {/*          }}*/}
          {/*          exit={{ color: hightlightTitleDefault[currentTheme] }}*/}
          {/*          className='-mx-1 -my-1 rounded-lg box-decoration-clone px-1 py-1 text-strong selection:bg-transparent'*/}
          {/*        >*/}
          {/*          {segment.val}*/}
          {/*        </motion.span>*/}
          {/*      )*/}
          {/*    } else*/}
          {/*      return (*/}
          {/*        <motion.span key={segment.val + 'inactive' + index}>*/}
          {/*          {segment.val}*/}
          {/*        </motion.span>*/}
          {/*      )*/}
          {/*  })}*/}
          {/*</AnimatePresence>*/}
        </div>
        <div className='text-xl text-medium'>
          {/*<AnimatePresence mode='wait'>*/}
          {/*  {bodySegments.map((segment, index) => {*/}
          {/*    if (segment.active) {*/}
          {/*      return (*/}
          {/*        <motion.span*/}
          {/*          key={segment.val + 'active' + index}*/}
          {/*          initial={{*/}
          {/*            backgroundColor: hightlightBodyDefault[currentTheme]*/}
          {/*          }}*/}
          {/*          animate={{*/}
          {/*            backgroundColor: highlightBody[active][currentTheme],*/}
          {/*            transition: { duration: 0.15 }*/}
          {/*          }}*/}
          {/*          exit={{*/}
          {/*            backgroundColor: hightlightBodyDefault[currentTheme]*/}
          {/*          }}*/}
          {/*          className='-mx-1 -my-1 rounded-lg box-decoration-clone px-1 py-1 selection:bg-transparent'*/}
          {/*        >*/}
          {/*          {segment.val}*/}
          {/*        </motion.span>*/}
          {/*      )*/}
          {/*    } else*/}
          {/*      return (*/}
          {/*        <motion.span key={segment.val + 'inactive' + index}>*/}
          {/*          {segment.val}*/}
          {/*        </motion.span>*/}
          {/*      )*/}
          {/*  })}*/}
          {/*</AnimatePresence>*/}
        </div>
        {content.button.length > 0 && (
          <ContentRouter shrink to={content.button[0]} />
        )}
      </div>
      <AnimatePresence mode='wait'>
        <motion.img
          alt='Hero image'
          key={content.iterations[active]?.image?.url + 'image'}
          src={content.iterations[active]?.image?.url}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='absolute -right-[10.5rem] top-0 hidden lg:block lg:dark:hidden'
        />
        <motion.img
          alt='Hero image'
          key={content.iterations[active]?.imageDark?.url + 'imageDark'}
          src={content.iterations[active]?.imageDark?.url}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
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
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1, transition: { duration: 0.15 } }}
                      exit={{ opacity: 0 }}
                      className='block dark:hidden'
                    />
                    <motion.img
                      alt={feature.title as string}
                      key={feature.imageDark?.url + 'imageDark'}
                      src={feature.imageDark?.url}
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
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1, transition: { duration: 0.15 } }}
                  exit={{ opacity: 0 }}
                  className='block w-full dark:hidden'
                />
                <motion.img
                  alt={feature.title as string}
                  key={feature.imageDark?.url + 'imageDark'}
                  src={feature.imageDark?.url}
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
  return (
    <div
      key={content.id}
      className='grid w-full grid-cols-12 gap-x-6 gap-y-10 px-4 lg:gap-y-14 lg:px-0'
    >
      <div className='col-span-full flex items-start lg:col-span-4'>
        <AnimatePresence mode='wait'>
          <motion.img
            alt='Fynbos shapes'
            key={content.image?.url + 'image'}
            src={content.image?.url}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, transition: { duration: 0.15 } }}
            exit={{ opacity: 0 }}
            className='block dark:hidden'
          />
          <motion.img
            alt='Fynbos shapes'
            key={content.imageDark?.url + 'imageDark'}
            src={content.imageDark?.url}
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
      className='grid w-full grid-cols-4 gap-y-20 px-4  sm:grid-cols-8 lg:grid-cols-12 lg:gap-x-10 lg:gap-y-14 lg:px-0'
    >
      <div className='col-span-full flex flex-col items-center space-y-10 text-center'>
        <AnimatePresence mode='wait'>
          <motion.img
            alt='Fynbos shapes'
            key={content.image?.url + 'image'}
            src={content.image?.url}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, transition: { duration: 0.15 } }}
            exit={{ opacity: 0 }}
            className='block dark:hidden'
          />
          <motion.img
            alt='Fynbos shapes'
            key={content.imageDark?.url + 'imageDark'}
            src={content.imageDark?.url}
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
        'flex flex-col items-center space-y-6 px-4 py-20 lg:px-0'
      )}
    >
      {content.title && (
        <h2 className='font-display text-4xl font-medium'>{content.title}</h2>
      )}
      {content.bodyText && (
        <Prose
          className={clsx(
            'max-w-none',
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
        <ContentRouter shrink to={content.button[0]} />
      )}
    </div>
  )
}
