import { FC } from 'react'
import { Switch as HeadlessSwitch } from '@headlessui/react'

type SwitchProps = {
  className?: string
  enabled: boolean
  onChange: any
}

export const Switch: FC<SwitchProps> = ({
  children,
  className,
  enabled,
  onChange
}) => {
  return (
    <HeadlessSwitch
      checked={enabled}
      onChange={onChange}
      className={`${enabled ? 'bg-primary' : 'bg-strong'}
          inline-flex flex-shrink-0 h-[24px] w-[46px] p-[2px] rounded-full cursor-pointer transition-colors ease-in-out duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-black`}
    >
      <span className='sr-only'>Test data</span>
      <span
        aria-hidden='true'
        className={`${enabled ? 'translate-x-[22px]' : 'translate-x-0'}
            pointer-events-none inline-block h-[20px] w-[20px] rounded-full bg-white shadow-lg transform transition ease-in-out duration-200`}
      />
    </HeadlessSwitch>
  )
}
