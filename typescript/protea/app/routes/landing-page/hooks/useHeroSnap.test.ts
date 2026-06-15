import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { CarouselScreen } from '../context/PhoneCarouselContext'
import { useHeroSnap } from './useHeroSnap'

/**
 * These tests drive the real hook through window wheel/scroll events against a
 * mocked layout. The hero is a 100vh section pinned at document top; on mobile
 * its rendered height (the large viewport) can exceed window.innerHeight once
 * the URL bar reappears on scroll-up. That mismatch is the crux of the
 * re-entry bug, so the harness models hero height and viewport height
 * independently.
 */

const HERO_DOC_TOP = 0
const COOLDOWN_MS = 600
const REENGAGE_COOLDOWN_MS = 1000

let clock = 0
let scrollY = 0
let heroHeight = 0
let section: HTMLElement

const advance = (ms: number) => {
  clock += ms
}

const setViewport = (innerHeight: number, heroPx: number) => {
  heroHeight = heroPx
  Object.defineProperty(window, 'innerHeight', {
    configurable: true,
    value: innerHeight
  })
}

const setScroll = (y: number) => {
  scrollY = y
}

const fireScroll = () => {
  act(() => {
    window.dispatchEvent(new Event('scroll'))
  })
}

/** One discrete wheel flick, spaced past the cooldown so it counts as a gesture. */
const wheelFlick = (deltaY: number) => {
  advance(COOLDOWN_MS + 100)
  act(() => {
    window.dispatchEvent(
      new WheelEvent('wheel', { deltaY, cancelable: true, bubbles: true })
    )
  })
}

const mount = () => {
  let screen: CarouselScreen = 1
  const setActiveScreen = vi.fn((s: CarouselScreen) => {
    screen = s
  })
  const ref = { current: section }
  const view = renderHook(() =>
    useHeroSnap({
      sectionRef: ref,
      activeScreen: 1,
      setActiveScreen,
      screenCount: 5,
      cooldownMs: COOLDOWN_MS
    })
  )
  return {
    get screen() {
      return screen
    },
    setActiveScreen,
    unmount: view.unmount
  }
}

beforeEach(() => {
  clock = 0
  scrollY = 0
  heroHeight = 0

  vi.spyOn(performance, 'now').mockImplementation(() => clock)

  Object.defineProperty(window, 'scrollY', {
    configurable: true,
    get: () => scrollY
  })
  window.scrollTo = ((arg: number | ScrollToOptions) => {
    scrollY = typeof arg === 'object' ? arg.top ?? scrollY : arg
  }) as typeof window.scrollTo

  section = document.createElement('section')
  document.body.appendChild(section)
  section.getBoundingClientRect = () => {
    const top = HERO_DOC_TOP - scrollY
    return {
      top,
      bottom: top + heroHeight,
      left: 0,
      right: 0,
      width: 0,
      height: heroHeight,
      x: 0,
      y: top,
      toJSON: () => ({})
    } as DOMRect
  }
})

afterEach(() => {
  vi.restoreAllMocks()
  section.remove()
})

/** Walk from screen 1 down to screen 5, then release past the hero into the page. */
const traverseDownAndRelease = (h: ReturnType<typeof mount>) => {
  wheelFlick(120) // 1 -> 2
  expect(h.screen).toBe(2) // first scroll advances one feature, never skips
  wheelFlick(120) // 2 -> 3
  wheelFlick(120) // 3 -> 4
  wheelFlick(120) // 4 -> 5
  expect(h.screen).toBe(5)

  wheelFlick(120) // 5 + down -> releasePastHero (unlock, scroll past)
  // releasePastHero scrolled us below the hero; record that as the baseline.
  fireScroll()
}

/** Simulate scrolling back up until the hero is pinned at the viewport top. */
const scrollUpToHero = () => {
  advance(REENGAGE_COOLDOWN_MS + 200)
  setScroll(scrollY - 200) // a step upward, hero still partly below
  fireScroll()
  setScroll(HERO_DOC_TOP) // hero top reaches viewport top
  fireScroll()
}

describe('useHeroSnap re-entry', () => {
  it('mobile (hero 100vh > innerHeight): scrolling up re-enters at the LAST screen, not screen 1', () => {
    setViewport(750, 812) // URL bar visible: viewport shorter than the 100vh hero
    const h = mount()

    traverseDownAndRelease(h)
    scrollUpToHero()

    expect(h.screen).toBe(5)
  })

  it('desktop (hero 100vh == innerHeight): scrolling up re-enters at the LAST screen', () => {
    setViewport(800, 800)
    const h = mount()

    traverseDownAndRelease(h)
    scrollUpToHero()

    expect(h.screen).toBe(5)
  })

  it('does not jump into the page on the first downward scroll from a fresh load', () => {
    setViewport(750, 812)
    const h = mount()

    // A stray scroll event at the very top must not advance the carousel.
    fireScroll()
    expect(h.screen).toBe(1)

    // First real flick goes one feature in, never releasing to the next section.
    wheelFlick(120)
    expect(h.screen).toBe(2)
    expect(h.setActiveScreen).not.toHaveBeenCalledWith(5)

    h.unmount()
  })
})
