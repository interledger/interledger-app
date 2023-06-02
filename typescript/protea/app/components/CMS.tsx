import clsx from 'clsx'
import { AnimatePresence, motion } from 'framer-motion'
import { useEffect, useMemo, useState } from 'react'
import { ButtonRouter, Shape } from '~/components'

import { StructuredText } from 'react-datocms'
import type {
  CtaContentRecord,
  FeatureBlocksContentRecord,
  FeatureContentRecord,
  HeaderContentRecord,
  HeroContentRecord,
  HomeHeroContentRecord,
  SectionModelContentField,
  SectionRecord,
  ShowcaseContentRecord,
  StoryContentRecord,
  TeamContentRecord,
  TextContentRecord
} from '~/generated/dato-cms-graphql'

export function renderMarketingPageWithSections(section: SectionRecord) {
  return (
    <div
      key={section.id}
      className='group w-full overflow-hidden even:rounded-2xl even:bg-mk-section'
    >
      <div className='flex w-full group-first:hidden group-even:hidden'>
        <Shape radius='rounded-l-full' color='bg-mk-section' width='w-20' />
        <Shape radius='rounded-full' color='bg-mk-section' width='w-20' />
        <Shape radius='rounded-full' color='bg-transparent' width='w-20' />
        <Shape radius='rounded-full' color='bg-mk-section' width='w-20' />
      </div>
      <div className='relative mx-auto flex w-full max-w-[59rem] flex-col items-center py-20'>
        {section.content.map((innerContent) =>
          renderSectionModelContentField(
            innerContent as SectionModelContentField
          )
        )}
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

function renderSectionModelContentField(content: SectionModelContentField) {
  switch (content.__typename) {
    case 'CtaContentRecord':
      return CtaContentRecordComponent(content)
    case 'FeatureBlocksContentRecord':
      return FeatureBlocksContentRecordComponent(content)
    case 'FeatureContentRecord':
      return FeatureContentRecordComponent(content)
    case 'HeaderContentRecord':
      return HeaderContentRecordComponent(content)
    case 'HeroContentRecord':
      return HeroContentRecordComponent(content)
    case 'HomeHeroContentRecord':
      return HomeHeroContentRecordComponent(content)
    case 'ShowcaseContentRecord':
      return ShowcaseContentRecordComponent(content)
    case 'StoryContentRecord':
      return StoryContentRecordComponent(content)
    case 'TeamContentRecord':
      return TeamContentRecordComponent(content)
    case 'TextContentRecord':
      return TextContentRecordComponent(content)
    default:
      return <h1 key={content.id}>Unknown</h1>
  }
}

// Stub out functions that return the content to replace the h1 tags in the switch cases above
// They should have typed input params and return the correct JSX
function CtaContentRecordComponent(content: CtaContentRecord) {
  return (
    <div
      key={content.id}
      className='flex w-full flex-col space-y-16 px-4 py-10 lg:flex-row lg:space-x-16 lg:space-y-0 lg:px-0 lg:py-16'
    >
      <div className='flex items-start'>
        <Shape radius='rounded-full' color='bg-sky-400' width='w-32' />
        <Shape radius='rounded-l-full' color='bg-slate-200' width='w-32' />
      </div>
      <div className='flex flex-col space-y-6'>
        <h2 className='text-4xl font-medium'>{content.title}</h2>
        <p className='text-2xl text-medium'>{content.body}</p>
        {content.button.length > 0 && (
          <ButtonRouter
            className='h-20 px-20'
            shrink
            to={content.button[0].url as string}
          >
            {content.button[0].displayText}
          </ButtonRouter>
        )}
      </div>
    </div>
  )
}

function FeatureBlocksContentRecordComponent(
  content: FeatureBlocksContentRecord
) {
  return <h1 key={content.id}>{content.__typename}</h1>
}

function FeatureContentRecordComponent(content: FeatureContentRecord) {
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
        className={clsx(
          'hidden lg:-mr-[10.5rem] lg:block',
          content.rowReverse && 'lg:-ml-[10.5rem] lg:mr-0'
        )}
        loading='lazy'
        src={content.image?.url}
      />
      <img
        className='mt-6 block lg:hidden'
        loading='lazy'
        src={content.imageMobile?.url}
      />
    </div>
  )
}

function HeaderContentRecordComponent(content: HeaderContentRecord) {
  return (
    <div
      key={content.id}
      className='flex w-full flex-col items-end space-y-20 px-4 lg:flex-row lg:flex-row-reverse lg:items-center lg:justify-between lg:space-y-0 lg:px-0'
    >
      {content.shapes && (
        <img
          className='block h-64 lg:-mr-20'
          src={content.shapes?.url}
          loading='lazy'
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

function HeroContentRecordComponent(content: HeroContentRecord) {
  return (
    <div
      key={content.id}
      className='flex w-full flex-col lg:-my-20 lg:flex-row lg:gap-x-12'
    >
      <AnimatePresence mode='wait'>
        <motion.img
          key={content.imageMobile?.url + 'imageMobile'}
          src={content.imageMobile?.url}
          loading='lazy'
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='-mt-20 block dark:hidden lg:hidden'
        />
        <motion.img
          key={content.imageDarkMobile?.url + 'imageDarkMobile'}
          src={content.imageDarkMobile?.url}
          loading='lazy'
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
          key={content.image?.url + 'image'}
          src={content.image?.url}
          loading='lazy'
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='-mr-[10.5rem] hidden lg:block lg:dark:hidden'
        />
        <motion.img
          key={content.imageDark?.url + 'imageDark'}
          src={content.imageDark?.url}
          loading='lazy'
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='-mr-[10.5rem] hidden lg:dark:block'
        />
      </AnimatePresence>
    </div>
  )
}

type Segment = {
  val: string
  active: boolean
}
function getSegments(str: string, search: string): Segment[] {
  const index = str.indexOf(search)
  const length = search.length
  return [
    { val: index > 0 ? str.slice(0, index) + ' ' : '', active: false },
    { val: str.slice(index, index + length), active: true },
    { val: ' ' + str.slice(index + length), active: false }
  ]
}

function HomeHeroContentRecordComponent(content: HomeHeroContentRecord) {
  // TODO Upgrade to pull theme once thats done
  const theme = 'system'
  // const { theme } = useLoaderData<typeof loader>()
  const [active, setActive] = useState<number>(0)

  // We do this because framer doesn't animate values from classes
  const currentTheme: 'dark' | 'light' = useMemo(() => {
    // if (theme === 'dark' || theme === 'light') return theme
    if (typeof window !== 'undefined' && theme === 'system') {
      const prefersDark = window.matchMedia(
        '(prefers-color-scheme: dark)'
      ).matches
      return prefersDark ? 'dark' : 'light'
    }
    return 'light'
  }, [theme])

  const hightlightTitleDefault = {
    dark: 'rgba(255, 255, 255, 1)',
    light: 'rgba(15, 23, 42, 1)'
  }

  const highlightTitle = [
    'rgba(99, 102, 241, 1)',
    'rgba(249, 115, 22, 1)',
    'rgba(168, 85, 247, 1)',
    'rgba(250, 204, 21, 1)'
  ]

  const hightlightBodyDefault = {
    dark: 'rgba(2, 6, 23, 1)',
    light: 'rgba(255, 255, 255, 1)'
  }

  const highlightBody = [
    {
      dark: 'rgba(49, 46, 129, 1)',
      light: 'rgba(224, 231, 255, 1)'
    },
    {
      dark: 'rgba(124, 45, 18, 1)',
      light: 'rgba(255, 237, 213, 1)'
    },
    {
      dark: 'rgba(88, 28, 135, 1)',
      light: 'rgba(243, 232, 255, 1)'
    },
    {
      dark: 'rgba(113, 63, 18, 1)',
      light: 'rgba(254, 249, 195, 1)'
    }
  ]

  const titleSegments = useMemo(() => {
    return getSegments(
      content.title as string,
      content.iterations[active].title as string
    )
  }, [active, content.body, content.iterations])

  const bodySegments = useMemo(() => {
    return getSegments(
      content.body as string,
      content.iterations[active].body as string
    )
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
      className='flex flex-col space-y-10 px-4 pb-10 lg:space-y-0 lg:py-16'
    >
      <AnimatePresence mode='wait'>
        <motion.img
          key={content.iterations[active]?.mobileShape?.url + 'shapeMobile'}
          src={content.iterations[active]?.mobileShape?.url}
          loading='lazy'
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='-mr-4 -mt-20 block h-44 w-44 self-end lg:hidden'
        />
      </AnimatePresence>
      <div className='flex w-full flex-col space-y-8 lg:w-1/2'>
        <div className='font-display text-5xl font-bold'>
          <AnimatePresence mode='wait'>
            {titleSegments.map((segment, index) => {
              if (segment.active) {
                return (
                  <motion.span
                    key={segment.val + 'active' + index}
                    initial={{ color: hightlightTitleDefault[currentTheme] }}
                    animate={{
                      color: highlightTitle[active],
                      transition: { duration: 0.15 }
                    }}
                    exit={{ color: hightlightTitleDefault[currentTheme] }}
                    className='-mx-1 -my-1 rounded-lg box-decoration-clone px-1 py-1 text-strong selection:bg-transparent'
                  >
                    {segment.val}
                  </motion.span>
                )
              } else
                return (
                  <motion.span key={segment.val + 'inactive' + index}>
                    {segment.val}
                  </motion.span>
                )
            })}
          </AnimatePresence>
        </div>
        <div className='text-xl text-medium'>
          <AnimatePresence mode='wait'>
            {bodySegments.map((segment, index) => {
              if (segment.active) {
                return (
                  <motion.span
                    key={segment.val + 'active' + index}
                    initial={{
                      backgroundColor: hightlightBodyDefault[currentTheme]
                    }}
                    animate={{
                      backgroundColor: highlightBody[active][currentTheme],
                      transition: { duration: 0.15 }
                    }}
                    exit={{
                      backgroundColor: hightlightBodyDefault[currentTheme]
                    }}
                    className='-mx-1 -my-1 rounded-lg box-decoration-clone px-1 py-1 selection:bg-transparent'
                  >
                    {segment.val}
                  </motion.span>
                )
              } else
                return (
                  <motion.span key={segment.val + 'inactive' + index}>
                    {segment.val}
                  </motion.span>
                )
            })}
          </AnimatePresence>
        </div>
        {content.button.length > 0 && (
          <ButtonRouter
            className='h-20 px-20'
            shrink
            to={content.button[0]?.url as string}
          >
            {content.button[0].displayText}
          </ButtonRouter>
        )}
      </div>
      <AnimatePresence mode='wait'>
        <motion.img
          key={content.iterations[active]?.imageMobile?.url + 'imageMobile'}
          src={content.iterations[active]?.imageMobile?.url}
          loading='lazy'
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='block w-full dark:hidden lg:hidden'
        />
        <motion.img
          key={
            content.iterations[active]?.imageDarkMobile?.url + 'imageDarkMobile'
          }
          src={content.iterations[active]?.imageDarkMobile?.url}
          loading='lazy'
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='hidden w-full dark:block lg:hidden'
        />
        <motion.img
          key={content.iterations[active]?.image?.url + 'image'}
          src={content.iterations[active]?.image?.url}
          loading='lazy'
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='absolute -right-[10.5rem] top-0 hidden lg:block lg:dark:hidden'
        />
        <motion.img
          key={content.iterations[active]?.imageDark?.url + 'imageDark'}
          src={content.iterations[active]?.imageDark?.url}
          loading='lazy'
          initial={{ opacity: 0 }}
          animate={{ opacity: 1, transition: { duration: 0.15 } }}
          exit={{ opacity: 0 }}
          className='absolute -right-[10.5rem] top-0 hidden lg:dark:block'
        />
      </AnimatePresence>
    </div>
  )
}

function ShowcaseContentRecordComponent(content: ShowcaseContentRecord) {
  return (
    <div key={content.id} className='mt-40 last-of-type:mt-0'>
      {content.__typename}
    </div>
  )
}

function StoryContentRecordComponent(content: StoryContentRecord) {
  return <h1 key={content.id}>{content.__typename}</h1>
}

function TeamContentRecordComponent(content: TeamContentRecord) {
  return <h1 key={content.id}>{content.__typename}</h1>
}

function TextContentRecordComponent(content: TextContentRecord) {
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
        <p className='text-lg text-medium lg:text-2xl'>
          <StructuredText data={content.bodyText.value} />
        </p>
      )}
      {content.button.length > 0 && (
        <ButtonRouter
          className='h-20 px-20'
          shrink
          to={content.button[0].url as string}
        >
          {content.button[0].displayText}
        </ButtonRouter>
      )}
    </div>
  )
}

// function FeaturesDesktop(content: ShowcaseContentRecord) {
//   let [changeCount, setChangeCount] = useState(0)
//   let [selectedIndex, setSelectedIndex] = useState(0)
//   let prevIndex = usePrevious(selectedIndex)
//   let isForwards = prevIndex === undefined ? true : selectedIndex > prevIndex
//
//   let onChange = useCallback(
//     (selectedIndex: number) => {
//       setSelectedIndex(selectedIndex)
//       setChangeCount((changeCount) => changeCount + 1)
//     },
//     [setSelectedIndex]
//   )
//
//   return (
//     <Tab.Group
//       as='div'
//       className='grid grid-cols-12 items-center gap-8 lg:gap-16 xl:gap-24'
//       selectedIndex={selectedIndex}
//       onChange={onChange}
//       vertical
//     >
//       <Tab.List className='relative z-10 order-last col-span-6 space-y-6'>
//         {content.cases.map((feature, featureIndex) => (
//           <div
//             key={feature.id}
//             className='relative rounded-2xl transition-colors hover:bg-nav-hover'
//           >
//             {featureIndex === selectedIndex && (
//               <motion.div
//                 layoutId='activeBackground'
//                 className='absolute inset-0 bg-nav'
//                 initial={{ borderRadius: 20 }}
//               />
//             )}
//             <div className='relative z-10 p-8'>
//               <h3 className='mt-6 text-lg font-semibold text-white'>
//                 <Tab className='text-left [&:not(:focus-visible)]:focus:outline-none'>
//                   <span className='absolute inset-0 rounded-2xl' />
//                   {feature.title}
//                 </Tab>
//               </h3>
//               <p className='mt-2 text-sm text-gray-400'>{feature.body}</p>
//             </div>
//           </div>
//         ))}
//       </Tab.List>
//       <div className='relative col-span-6'>
//         <div className='absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2'>
//           <CircleBackground color='#13B5C8' className='animate-spin-slower' />
//         </div>
//         <PhoneFrame className='z-10 mx-auto w-full max-w-[366px]'>
//           <Tab.Panels as={Fragment}>
//             <AnimatePresence
//               initial={false}
//               custom={{ isForwards, changeCount }}
//             >
//               {features.map((feature, featureIndex) =>
//                 selectedIndex === featureIndex ? (
//                   <Tab.Panel
//                     static
//                     key={feature.name + changeCount}
//                     className='col-start-1 row-start-1 flex focus:outline-offset-[32px] [&:not(:focus-visible)]:focus:outline-none'
//                   >
//                     <feature.screen
//                       animated
//                       custom={{ isForwards, changeCount }}
//                     />
//                   </Tab.Panel>
//                 ) : null
//               )}
//             </AnimatePresence>
//           </Tab.Panels>
//         </PhoneFrame>
//       </div>
//     </Tab.Group>
//   )
// }
//
// function FeaturesMobile() {
//   let [activeIndex, setActiveIndex] = useState(0)
//   let slideContainerRef = useRef()
//   let slideRefs = useRef([])
//
//   useEffect(() => {
//     let observer = new window.IntersectionObserver(
//       (entries) => {
//         for (let entry of entries) {
//           if (entry.isIntersecting) {
//             setActiveIndex(slideRefs.current.indexOf(entry.target))
//             break
//           }
//         }
//       },
//       {
//         root: slideContainerRef.current,
//         threshold: 0.6
//       }
//     )
//
//     for (let slide of slideRefs.current) {
//       if (slide) {
//         observer.observe(slide)
//       }
//     }
//
//     return () => {
//       observer.disconnect()
//     }
//   }, [slideContainerRef, slideRefs])
//
//   return (
//     <>
//       <div
//         ref={slideContainerRef}
//         className='-mb-4 flex snap-x snap-mandatory -space-x-4 overflow-x-auto overscroll-x-contain scroll-smooth pb-4 [scrollbar-width:none] sm:-space-x-6 [&::-webkit-scrollbar]:hidden'
//       >
//         {features.map((feature, featureIndex) => (
//           <div
//             key={featureIndex}
//             ref={(ref) => (slideRefs.current[featureIndex] = ref)}
//             className='w-full flex-none snap-center px-4 sm:px-6'
//           >
//             <div className='relative transform overflow-hidden rounded-2xl bg-gray-800 px-5 py-6'>
//               <div className='absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2'>
//                 <CircleBackground
//                   color='#13B5C8'
//                   className={featureIndex % 2 === 1 ? 'rotate-180' : undefined}
//                 />
//               </div>
//               <PhoneFrame className='relative mx-auto w-full max-w-[366px]'>
//                 <feature.screen />
//               </PhoneFrame>
//               <div className='absolute inset-x-0 bottom-0 bg-gray-800/95 p-6 backdrop-blur sm:p-10'>
//                 <feature.icon className='h-8 w-8' />
//                 <h3 className='mt-6 text-sm font-semibold text-white sm:text-lg'>
//                   {feature.name}
//                 </h3>
//                 <p className='mt-2 text-sm text-gray-400'>
//                   {feature.description}
//                 </p>
//               </div>
//             </div>
//           </div>
//         ))}
//       </div>
//       <div className='mt-6 flex justify-center gap-3'>
//         {features.map((_, featureIndex) => (
//           <button
//             type='button'
//             key={featureIndex}
//             className={clsx(
//               'relative h-0.5 w-4 rounded-full',
//               featureIndex === activeIndex ? 'bg-gray-300' : 'bg-gray-500'
//             )}
//             aria-label={`Go to slide ${featureIndex + 1}`}
//             onClick={() => {
//               slideRefs.current[featureIndex].scrollIntoView({
//                 block: 'nearest',
//                 inline: 'nearest'
//               })
//             }}
//           >
//             <span className='absolute -inset-x-1.5 -inset-y-3' />
//           </button>
//         ))}
//       </div>
//     </>
//   )
// }
