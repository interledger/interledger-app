import type { CSSProperties } from "react"

type GlowType = "hero" | "default" | "type2"

interface GlowProps {
  type?: GlowType
  x?: number | string
  y?: number | string
  className?: string
}

const GRADIENT_VAR: Record<GlowType, string> = {
  hero:    "var(--glow-hero-color)",
  default: "var(--glow-default-color)",
  type2:   "var(--glow-type2-color)",
}

/**
 * Glow effect component.
 * Always 400×300px, absolutely positioned within its nearest relative parent.
 * Implemented as a CSS gradient + blur(140px) — matches Figma annotation
 * "Linear gradient + 140 Blur" from features-and-animations.md §7.
 *
 * Gradient colors are placeholder tokens (needs-verification from Figma visual inspection).
 *
 * @param type  - "hero" | "default" | "type2" (maps to CSS token)
 * @param x     - left offset within parent (px number or CSS string)
 * @param y     - top offset within parent (px number or CSS string)
 */
export function Glow({ type = "default", x = 0, y = 0, className }: GlowProps) {
  const style: CSSProperties = {
    position: "absolute",
    left: typeof x === "number" ? `${x}px` : x,
    top:  typeof y === "number" ? `${y}px` : y,
    width:  "400px",
    height: "300px",
    background: GRADIENT_VAR[type],
    filter: `blur(var(--glow-blur, 140px))`,
    opacity: "var(--glow-opacity, 1)",
    pointerEvents: "none",
    userSelect: "none",
    zIndex: 0,
  }

  return <div aria-hidden="true" style={style} className={className} />
}
