import type { FC } from 'react'

export const VirtualCardRibbon: FC = () => {
  return (
    <div className='absolute left-4 top-0 z-10'>
      <div className='rounded-b bg-purple-600 px-1.5 py-[3px] text-white'>
        <span className='text-[7px] font-semibold uppercase tracking-wide'>
          Virtual
        </span>
      </div>
    </div>
  )
}
