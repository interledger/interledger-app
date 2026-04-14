import type { CSSProperties } from "react"
import { motion } from "framer-motion"

export interface GlowScrollTransform {
  scale: number
  rotate: number
  y: number
  opacity: number
}

interface GlowProps {
  x?: number | string
  y?: number | string
  className?: string
  scrollTransform?: GlowScrollTransform
}

const DEFAULT_TRANSFORM: GlowScrollTransform = {
  scale: 1,
  rotate: 0,
  y: 0,
  opacity: 0.5,
}

/**
 * Glow effect — Figma "Glow Gradient1" component.
 * 400×300px, rounded-full, linear-gradient #FFBEBE → #48156E, blur 140px, opacity 0.5.
 *
 * Enter effect: opacity 0 → target (600ms easeOut).
 * Scroll transform: driven by scrollTransform prop (scale, rotate, y, opacity per screen).
 */
export function Glow({ x = 0, y = 0, className, scrollTransform }: GlowProps) {
  const t = scrollTransform ?? DEFAULT_TRANSFORM

  const style: CSSProperties = {
    position: "absolute",
    left: typeof x === "number" ? `${x}px` : x,
    top:  typeof y === "number" ? `${y}px` : y,
    width:  "400px",
    height: "300px",
    background: "linear-gradient(135deg, #FFBEBE 0%, #48156E 100%)",
    filter: "blur(140px)",
    borderRadius: "9999px",
    pointerEvents: "none",
    userSelect: "none",
    zIndex: 0,
  }

  return (
    <motion.div
      aria-hidden="true"
      style={style}
      className={className}
      initial={{ opacity: 0, x: "-50%", scale: DEFAULT_TRANSFORM.scale, rotate: DEFAULT_TRANSFORM.rotate, y: DEFAULT_TRANSFORM.y }}
      animate={{ opacity: t.opacity, x: "-50%", scale: t.scale, rotate: t.rotate, y: t.y }}
      transition={{ duration: 0.6, ease: "easeOut" }}
    />
  )
}
