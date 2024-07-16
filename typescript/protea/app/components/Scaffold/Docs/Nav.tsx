import type { SerializeFrom } from '@remix-run/node'
import {
  Link,
  useLocation,
  useResolvedPath,
  useRouteLoaderData
} from '@remix-run/react'
import clsx from 'clsx'
import { AnimatePresence, motion, useIsPresent } from 'framer-motion'
import type { FC, ReactNode } from 'react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import { InterledgerLogo, IconButton, Router } from '~/components'
import type { DocsNav } from '~/components/Scaffold/Docs/useDocsStore'
import { useDocsStore } from '~/components/Scaffold/Docs/useDocsStore'
import { NavDrawer } from '~/components/Scaffold/NavDrawer'
import type { loader as docsLoader } from '~/routes/docs'

type DocItemProps = {
  children?: ReactNode
  to: string
  type?: 'doc' | 'section'
}

// This is a custom NavLink that includes anchor tags
export const DocItem: FC<DocItemProps> = ({ children, to, type }) => {
  let path = useResolvedPath(to)
  let location = useLocation()

  let toPathname = (path.pathname + path.hash).toLowerCase()
  let locationPathname = (location.pathname + location.hash).toLowerCase()

  let isActive = locationPathname === toPathname

  if (typeof document === 'undefined') return null

  return (
    <Link
      prefetch='intent'
      className={clsx(
        isActive && 'bg-nav-active',
        'w-full rounded-xl hover:bg-nav-hover focus-visible:outline focus-visible:outline-2 focus-visible:outline-focus'
      )}
      to={to}
    >
      <li className='relative flex w-56 items-center rounded-xl p-4'>
        <span className='z-10 truncate'>{children}</span>
      </li>
    </Link>
  )
}

DocItem.displayName = 'ListItem'

type NavGroupProps = {
  children?: ReactNode
  doc: DocsNav
}

export const NavGroup: FC<NavGroupProps> = ({ children, doc }) => {
  let location = useLocation()
  let isActiveGroup = location.pathname?.includes(doc.slug || ' ')
  return (
    <li className='relative mt-4'>
      <ul className='flex w-full flex-col gap-y-4'>
        <DocItem
          key={doc.id + 'item'}
          to={route('/docs/:slug', { slug: doc.slug as string })}
        >
          {doc.title}
        </DocItem>
        {/* TODO animate these in */}
        <div className='hidden lg:contents'>
          {isActiveGroup &&
            doc.sections?.map((section) => {
              if (!section.slug || !section.title) return null
              return (
                <DocItem
                  key={section.id + 'section'}
                  to={route('/docs/:slug', {
                    slug: `${doc.slug}#${section?.slug}`
                  })}
                >
                  <span className='ml-4'>{section?.title}</span>
                </DocItem>
              )
            })}
        </div>
      </ul>
      <AnimatePresence>
        {isActiveGroup && <VisibleSectionHighlight doc={doc} />}
      </AnimatePresence>
    </li>
  )
}

NavGroup.displayName = 'NavGroup'

type ListProps = {
  children?: ReactNode
}

const List: FC<ListProps> = ({ children }) => {
  return <ul className='flex w-full flex-col'>{children}</ul>
}

List.displayName = 'List'

type NavDrawerRootProps = {
  onClick?: () => void
}

// Client-only components - stops the shapes' classname from changing on initial load
// https://remix.run/docs/en/1.18.0/guides/migrating-react-router-app#client-only-components
let isHydrating = true

export const DocsNavDrawer: FC<NavDrawerRootProps> = ({ onClick }) => {
  const docsLoaderData = useRouteLoaderData('routes/docs') as SerializeFrom<
    typeof docsLoader
  >

  const [isHydrated, setIsHydrated] = useState(!isHydrating)
  const [navItems, setNavSections] = useDocsStore((state) => [
    state.sections,
    state.setSections
  ])

  // We do this because the server doesn't know about location.hash
  useEffect(() => {
    isHydrating = false
    setIsHydrated(true)
  }, [])

  useEffect(() => {
    if (docsLoaderData && docsLoaderData.allDocs && navItems.length == 0) {
      setNavSections(docsLoaderData.allDocs)
    }
  })

  useVisibleSections()

  if (!isHydrated) {
    return null
  }

  return (
    <NavDrawer.List>
      <div className='relative mb-8 ml-1 flex items-center space-x-4'>
        <IconButton
          className='lg:hidden'
          onClick={onClick}
          aria-label='Close menu'
        >
          menu_open
        </IconButton>
        <Router to={route('/')} aria-label='Fynbos logo'>
          <InterledgerLogo className='h-12' />
        </Router>
      </div>

      {navItems &&
        navItems.map((item) => <NavGroup key={item.id} doc={item} />)}
    </NavDrawer.List>
  )
}

DocsNavDrawer.displayName = 'DocsNavDrawer'

type VisibleSectionHighlightProps = {
  doc: DocsNav
}

const VisibleSectionHighlight: FC<VisibleSectionHighlightProps> = ({ doc }) => {
  const [visibleSections] = useDocsStore((state) => [state.visibleSections])

  let isPresent = useIsPresent()
  let firstVisibleSectionIndex = Math.max(
    0,
    [{ id: '_top' }, ...doc.sections].findIndex(
      (section) => section.id === visibleSections[0]
    )
  )
  let itemHeight = remToPx('3.5')
  let gapHeight = remToPx('1')
  let height = isPresent
    ? Math.max(1, visibleSections.length) * itemHeight +
      gapHeight * (visibleSections.length - 1)
    : itemHeight

  let top = firstVisibleSectionIndex
    ? firstVisibleSectionIndex * itemHeight +
      firstVisibleSectionIndex * gapHeight
    : 0

  return (
    <motion.div
      layout
      initial={{ opacity: 0 }}
      animate={{ opacity: 1, transition: { delay: 0.2 } }}
      exit={{ opacity: 0 }}
      className='absolute inset-x-0 top-0 -z-50 bg-nav will-change-transform dark:bg-slate-900'
      style={{ borderRadius: 12, height, top }}
    />
  )
}

function useVisibleSections() {
  let location = useLocation() // Use location to determine which doc is visible
  // let setVisibleSections = useStore(sectionStore, (s) => s.setVisibleSections)
  // let sections = useStore(sectionStore, (s) => s.sections)

  const [sections, setVisibleSections] = useDocsStore((state) => [
    state.sections,
    state.setVisibleSections
  ])

  useEffect(() => {
    function checkVisibleSections() {
      let { innerHeight, scrollY } = window
      let newVisibleSections = []

      const doc = sections.find((doc) =>
        location.pathname.includes(doc.slug || ' ')
      )

      if (!doc) return

      for (
        let sectionIndex = 0;
        sectionIndex < doc.sections.length;
        sectionIndex++
      ) {
        let { id, ref } = doc.sections[sectionIndex]
        let offset = remToPx('4')
        if (!ref || !ref.current) continue
        let top = ref.current.getBoundingClientRect().top + scrollY

        if (sectionIndex === 0 && top - offset > scrollY) {
          newVisibleSections.push('_top')
        }

        let nextSection = doc.sections[sectionIndex + 1]?.ref
        let bottom =
          ((nextSection &&
            nextSection.current &&
            nextSection.current.getBoundingClientRect().top) ??
            Infinity) +
          scrollY -
          remToPx('4')

        if (
          (top > scrollY && top < scrollY + innerHeight) ||
          (bottom > scrollY && bottom < scrollY + innerHeight) ||
          (top <= scrollY && bottom >= scrollY + innerHeight)
        ) {
          newVisibleSections.push(id)
        }
      }

      setVisibleSections(newVisibleSections)
    }

    let raf = window.requestAnimationFrame(() => checkVisibleSections())
    window.addEventListener('scroll', checkVisibleSections, { passive: true })
    window.addEventListener('resize', checkVisibleSections)

    return () => {
      window.cancelAnimationFrame(raf)
      window.removeEventListener('scroll', checkVisibleSections)
      window.removeEventListener('resize', checkVisibleSections)
    }
  }, [setVisibleSections, sections, location.pathname])
}

export function remToPx(remValue: string) {
  let rootFontSize =
    typeof window === 'undefined'
      ? 16
      : parseFloat(window.getComputedStyle(document.documentElement).fontSize)

  return parseFloat(remValue) * rootFontSize
}
