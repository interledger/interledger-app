import { ButtonRouter, Icon, Router } from '~/components'
import { route } from 'routes-gen'

const shapes = [
  [
    'bg-transparent',
    'bg-transparent',
    'bg-slate-100 rounded-tl-full',
    'bg-green-50 rounded-br-full',
    'bg-transparent'
  ],
  [
    'bg-slate-50 rounded-full',
    'bg-green-200 rounded-full',
    'bg-green-500 rounded-full',
    'bg-green-100 rounded-tl-full',
    'bg-slate-50 rounded-tr-full'
  ],
  [
    'bg-transparent',
    'bg-green-50 rounded-l-full',
    'bg-green-50 ',
    'bg-green-200 rounded-br-full',
    'bg-transparent'
  ]
]

export default function Page() {
  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <div className='mt-2 flex flex-col'>
        {shapes.map((shapeRow, outerIndex, outerArray) => (
          <div className='flex w-full justify-center' key={shapeRow.toString()}>
            {shapeRow.map((shape, index, array) => (
              <div
                key={shape + index}
                className={`flex aspect-square h-14 w-14 w-full items-center justify-center ${shape}`}
              >
                {Math.floor(outerArray.length / 2) == outerIndex &&
                  Math.floor(array.length / 2) == index && (
                    <Icon className='text-white'>check</Icon>
                  )}
              </div>
            ))}
          </div>
        ))}
      </div>

      <span className='mt-6 font-display text-2xl font-medium'>Thank you</span>
      <span className='mt-6 text-medium'>Your message has been sent.</span>
      <span className='mt-2 text-medium'>
        One of our team members will get back to you in due course.
      </span>

      <div className='flex justify-end pt-12'>
        <ButtonRouter to={route('/')}>Close</ButtonRouter>
      </div>
    </div>
  )
}
