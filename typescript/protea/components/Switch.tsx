import { FC } from 'react'
import { Switch as HeadlessSwitch } from '@headlessui/react'

type SwitchProps = {
  className?: string
  // The toggled state of the switch.
  checked: boolean
  // Whether the switch is active and can be toggled.
  disabled: boolean
  onChange: any
}

export const Switch: FC<SwitchProps> = ({
  children,
  className,
  checked,
  disabled,
  onChange
}) => {
  return (
    <HeadlessSwitch
      checked={checked}
      disabled={disabled}
      onChange={onChange}
      className={`${checked ? 'bg-primary' : 'bg-strong'}
          inline-flex flex-shrink-0 h-[24px] w-[46px] p-[2px] rounded-full cursor-pointer disabled:cursor-not-allowed transition-colors ease-in-out duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-black disabled:bg-opacity-40`}
    >
      <span className='sr-only'>Test data</span>
      <span
        aria-hidden='true'
        className={`${checked ? 'translate-x-[22px]' : 'translate-x-0'}
            pointer-events-none inline-block h-[20px] w-[20px] rounded-full bg-white shadow-lg transform transition ease-in-out duration-200`}
      />
    </HeadlessSwitch>
  )
}
