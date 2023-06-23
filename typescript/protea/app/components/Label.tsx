import clsx from 'clsx'
import type { LabelHTMLAttributes } from 'react'
import { forwardRef } from 'react'

const Label = forwardRef<
  HTMLLabelElement,
  LabelHTMLAttributes<HTMLLabelElement>
>(({ className, ...props }, ref) => {
  return (
    <label
      ref={ref}
      className={clsx('ml-2 block text-sm font-medium text-medium', className)}
      {...props}
    />
  )
})
Label.displayName = 'Label'

export { Label }
