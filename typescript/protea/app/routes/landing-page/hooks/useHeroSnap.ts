import { useEffect, useRef, type RefObject } from "react"
import type { CarouselScreen } from "../context/PhoneCarouselContext"

interface UseHeroSnapArgs {
  sectionRef: RefObject<HTMLElement | null>
  activeScreen: CarouselScreen
  setActiveScreen: (s: CarouselScreen) => void
  screenCount?: number
  cooldownMs?: number
}

type Direction = 1 | -1
type KeyIntent = Direction | "first" | "last" | null

// Wheel deltas smaller than this are noise (e.g. high-precision touchpads at rest).
const WHEEL_DEAD_ZONE = 4
// Minimum vertical swipe distance to count as a gesture.
const TOUCH_THRESHOLD_PX = 30
// Wheel events closer in time than this are treated as one gesture
// (handles trackpad inertia and long flicks: only the first event advances).
const GESTURE_GAP_MS = 150
// After releasing at the bottom edge, ignore scroll-driven re-engage for this
// window so the in-flight native scroll past the hero isn't intercepted.
const REENGAGE_COOLDOWN_MS = 1000

const intentFromKey = (key: string): KeyIntent => {
  switch (key) {
    case "ArrowDown":
    case "PageDown":
    case " ":
      return 1
    case "ArrowUp":
    case "PageUp":
      return -1
    case "Home":
      return "first"
    case "End":
      return "last"
    default:
      return null
  }
}

const isEditableTarget = (t: EventTarget | null): boolean => {
  const el = t as HTMLElement | null
  return !!el && (el.tagName === "INPUT" || el.tagName === "TEXTAREA" || el.isContentEditable)
}

/**
 * Pins the hero section and turns each discrete scroll/swipe/key gesture
 * into a single screen advance. Releases the lock at the bottom edge so the
 * page can continue into the next section, and re-engages when the user
 * scrolls back up to the hero.
 */
export function useHeroSnap({
  sectionRef,
  activeScreen,
  setActiveScreen,
  screenCount = 5,
  cooldownMs = 100,
}: UseHeroSnapArgs) {
  // All mutable per-effect state lives in one ref to avoid stale closures
  // and to keep the timing/locking concerns visible together.
  const stateRef = useRef({
    // Current carousel screen (1–5). Mirrors React state but lives in the ref
    // so event handlers always read the latest value without stale closures.
    // Kept in sync by the first useEffect.
    screen: activeScreen,

    // Whether body scroll is currently locked. Guards all event handlers —
    // they bail immediately if false. Also prevents double-lock / double-unlock.
    locked: false,

    // performance.now() of the last screen advance. Enforces the cooldown:
    // gestures arriving within cooldownMs of the last advance are discarded
    // to prevent rapid-fire screen skipping.
    lastAdvanceAt: 0,

    // performance.now() of the last wheel event that passed the dead-zone check.
    // Used to compute isNewGesture = now - lastWheelAt > GESTURE_GAP_MS.
    // Updated on every non-noise event, not just advances. This is the inertia
    // filter: trackpads fire a burst of events; only the first in each gap window
    // counts as a new gesture.
    lastWheelAt: 0,

    // Same as lastWheelAt but for keyboard. Updated on every direction key event.
    // Prevents OS auto-repeat from firing multiple advances while a key is held.
    lastKeyAt: 0,

    // performance.now() of the last releasePastHero() call. The onScroll
    // re-engage handler skips if now - lastReleaseAt < REENGAGE_COOLDOWN_MS,
    // giving the in-flight native scroll past the hero time to complete before
    // the lock can re-engage.
    lastReleaseAt: 0,

    // clientY of the first touch point, recorded in onTouchStart.
    // Compared against onTouchEnd.clientY to compute swipe delta.
    // null when no touch is active or onTouchEnd fires without a preceding start.
    touchStartY: null as number | null,

    // Snapshots of original CSS values before lockBody overwrites them.
    // unlockBody restores exactly these to avoid clobbering other code that
    // may have set those properties before the hero mounted.
    savedBodyOverflow: "",
    savedHtmlOverflow: "",
    savedBodyPaddingRight: "",
  })

  useEffect(() => {
    stateRef.current.screen = activeScreen
  }, [activeScreen])

  useEffect(() => {
    const section = sectionRef.current
    if (!section) return
    const state = stateRef.current

    // ── Body scroll lock ──────────────────────────────────────────────
    const lockBody = () => {
      if (state.locked) return
      state.locked = true
      state.savedBodyOverflow = document.body.style.overflow
      state.savedHtmlOverflow = document.documentElement.style.overflow
      state.savedBodyPaddingRight = document.body.style.paddingRight
      // Compensate the disappearing scrollbar so the layout doesn't shift.
      const scrollbarWidth = window.innerWidth - document.documentElement.clientWidth
      if (scrollbarWidth > 0) {
        document.body.style.paddingRight = `${scrollbarWidth}px`
      }
      document.body.style.overflow = "hidden"
      document.documentElement.style.overflow = "hidden"
    }

    const unlockBody = () => {
      if (!state.locked) return
      state.locked = false
      document.body.style.overflow = state.savedBodyOverflow
      document.documentElement.style.overflow = state.savedHtmlOverflow
      document.body.style.paddingRight = state.savedBodyPaddingRight
    }

    // ── Navigation ────────────────────────────────────────────────────
    const goToScreen = (n: CarouselScreen) => {
      state.screen = n
      setActiveScreen(n)
      state.lastAdvanceAt = performance.now()
    }

    const releasePastHero = () => {
      const rect = section.getBoundingClientRect()
      state.lastReleaseAt = performance.now()
      unlockBody()
      window.scrollTo({ top: window.scrollY + rect.bottom + 4, behavior: "auto" })
    }

    /**
     * Apply a direction at the current screen.
     *  - bottom edge + down → release the lock (let the page continue)
     *  - top edge + up      → no-op (hero is at the top of the document)
     *  - otherwise          → advance one screen
     */
    const navigate = (dir: Direction) => {
      const at = state.screen
      if (dir === 1 && at === screenCount) return releasePastHero()
      if (dir === -1 && at === 1) return
      goToScreen((at + dir) as CarouselScreen)
    }

    // ── Event handlers ────────────────────────────────────────────────
    const onWheel = (e: WheelEvent) => {
      if (!state.locked) return
      if (Math.abs(e.deltaY) < WHEEL_DEAD_ZONE) return
      e.preventDefault()
      const now = performance.now()
      const isNewGesture = now - state.lastWheelAt > GESTURE_GAP_MS
      state.lastWheelAt = now
      if (!isNewGesture) return
      if (now - state.lastAdvanceAt < cooldownMs) return
      navigate(e.deltaY > 0 ? 1 : -1)
    }

    const onTouchStart = (e: TouchEvent) => {
      if (!state.locked) return
      state.touchStartY = e.touches[0]?.clientY ?? null
    }

    const onTouchMove = (e: TouchEvent) => {
      if (state.locked) e.preventDefault()
    }

    const onTouchEnd = (e: TouchEvent) => {
      if (!state.locked) return
      const startY = state.touchStartY
      state.touchStartY = null
      if (startY == null) return
      const endY = e.changedTouches[0]?.clientY ?? startY
      const delta = startY - endY
      if (Math.abs(delta) < TOUCH_THRESHOLD_PX) return
      if (performance.now() - state.lastAdvanceAt < cooldownMs) return
      navigate(delta > 0 ? 1 : -1)
    }

    const onKeyDown = (e: KeyboardEvent) => {
      if (!state.locked) return
      if (isEditableTarget(e.target)) return
      const intent = intentFromKey(e.key)
      if (intent === null) return
      e.preventDefault()
      if (intent === "first") return goToScreen(1)
      if (intent === "last") return goToScreen(screenCount as CarouselScreen)
      const now = performance.now()
      // Held key / OS auto-repeat: swallow.
      if (e.repeat || now - state.lastKeyAt < GESTURE_GAP_MS) {
        state.lastKeyAt = now
        return
      }
      state.lastKeyAt = now
      if (now - state.lastAdvanceAt < cooldownMs) return
      navigate(intent)
    }

    // ── Lock lifecycle ────────────────────────────────────────────────
    /** Re-engage when the user scrolls back up to the hero from below. */
    const onScroll = () => {
      if (state.locked) return
      if (performance.now() - state.lastReleaseAt < REENGAGE_COOLDOWN_MS) return
      const rect = section.getBoundingClientRect()
      const fullyAtTop = rect.top >= 0 && rect.top < window.innerHeight && rect.bottom > 0
      if (!fullyAtTop) return
      const fromBelow = rect.bottom <= window.innerHeight + 4
      const entryScreen = (fromBelow ? screenCount : 1) as CarouselScreen
      state.screen = entryScreen
      setActiveScreen(entryScreen)
      lockBody()
      window.scrollTo({ top: window.scrollY + rect.top, behavior: "auto" })
    }

    const initLockState = () => {
      const rect = section.getBoundingClientRect()
      const offscreen = rect.bottom <= 0 || rect.top >= window.innerHeight
      if (offscreen) unlockBody()
      else lockBody()
    }

    initLockState()

    window.addEventListener("wheel", onWheel, { passive: false })
    window.addEventListener("touchstart", onTouchStart, { passive: true })
    window.addEventListener("touchmove", onTouchMove, { passive: false })
    window.addEventListener("touchend", onTouchEnd, { passive: true })
    window.addEventListener("keydown", onKeyDown)
    window.addEventListener("scroll", onScroll, { passive: true })

    return () => {
      window.removeEventListener("wheel", onWheel)
      window.removeEventListener("touchstart", onTouchStart)
      window.removeEventListener("touchmove", onTouchMove)
      window.removeEventListener("touchend", onTouchEnd)
      window.removeEventListener("keydown", onKeyDown)
      window.removeEventListener("scroll", onScroll)
      unlockBody()
    }
    // sectionRef is a ref; setActiveScreen is stable from context.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [screenCount, cooldownMs])
}
