import { Switch as HeadlessSwitch } from '@headlessui/react'
import type { FC } from 'react'

type SwitchProps = {
  className?: string
  srLabel?: string
  // The toggled state of the switch.
  checked: boolean
  // Whether the switch is active and can be toggled.
  disabled: boolean
  onChange: any
}

export const Switch: FC<SwitchProps> = ({
  srLabel,
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
          inline-flex h-[24px] w-[46px] flex-shrink-0 cursor-pointer rounded-full p-[2px] transition-colors duration-200 ease-in-out focus:outline-none focus-visible:outline-2 focus-visible:outline-black disabled:cursor-not-allowed disabled:bg-opacity-40`}
    >
      {srLabel && <span className='sr-only'>{srLabel}</span>}
      <span
        aria-hidden='true'
        className={`${checked ? 'translate-x-[22px]' : 'translate-x-0'}
            pointer-events-none inline-block h-[20px] w-[20px] transform rounded-full bg-app shadow-lg transition duration-200 ease-in-out`}
      />
    </HeadlessSwitch>
  )
}
