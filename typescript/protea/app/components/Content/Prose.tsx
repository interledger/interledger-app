import clsx from 'clsx'
import type { FC, HTMLAttributes } from 'react'

/**
 * Implements all the custom mark types
 * We use tailwind prose for this, as it then filters down to code blocks too. It also comes with sane defaults.
 */
export const Prose: FC<HTMLAttributes<HTMLDivElement>> = ({
  className,
  children
}) => {
  return (
    <div
      className={clsx(
        'prose prose-slate dark:prose-invert',
        className,
        'prose-h1:font-display prose-h1:font-medium',
        'prose-h2:font-display prose-h2:font-medium',
        'prose-h3:font-display prose-h3:font-medium',
        'prose-h4:font-display prose-h4:font-medium',
        'prose-h5:font-display prose-h5:font-medium',
        'prose-h6:font-display prose-h6:font-medium',
        'prose-a:rounded prose-a:font-normal prose-a:text-primary prose-a:no-underline prose-a:focus-visible:outline prose-a:focus-visible:outline-2 prose-a:focus-visible:outline-focus',
        'prose-blockquote:border-0 prose-blockquote:p-0 prose-blockquote:text-3xl prose-blockquote:font-normal prose-blockquote:not-italic',
        'prose-strong:font-medium',
        'prose-code:font-normal prose-code:tracking-wider'
      )}
    >
      {children}
    </div>
  )
}
