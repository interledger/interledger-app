import clsx from 'clsx'
import type { LabelHTMLAttributes } from 'react'
import { forwardRef } from 'react'

const Label = forwardRef<
  HTMLLabelElement,
  LabelHTMLAttributes<HTMLLabelElement>
>(({ className, ...props }, ref) => {
  return (
    // eslint-disable-next-line jsx-a11y/label-has-associated-control -- this is a primitive label; the control is associated by the consumer via htmlFor/{...props}
    <label
      ref={ref}
      className={clsx(
        'ml-2 block text-sm font-medium text-medium first:mt-2',
        className
      )}
      {...props}
    />
  )
})
Label.displayName = 'Label'

export { Label }
