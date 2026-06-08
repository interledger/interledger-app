import type { ReactNode } from 'react'

interface PageSectionProps {
  children: ReactNode
  className?: string
  innerClassName?: string
  id?: string
  style?: React.CSSProperties
}

export function PageSection({
  children,
  className = '',
  innerClassName = '',
  id,
  style
}: PageSectionProps) {
  return (
    <section
      id={id}
      className={`page-section content-section ${className}`}
      style={style}
    >
      <div className={`page-section-inner ${innerClassName}`}>{children}</div>
    </section>
  )
}
