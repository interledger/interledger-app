import type { ReactNode } from "react"

interface PageSectionProps {
  children: ReactNode
  className?: string
  innerClassName?: string
  id?: string
  style?: React.CSSProperties
}

/**
 * PageSection — standard layout wrapper for landing page sections.
 * Handles responsive horizontal padding and max-width centering.
 */
export function PageSection({ children, className = "", innerClassName = "", id, style }: PageSectionProps) {
  return (
    <section id={id} className={`page-section pad-default ${className}`} style={style}>
      <div className={`page-section-inner ${innerClassName}`}>
        {children}
      </div>
    </section>
  )
}
