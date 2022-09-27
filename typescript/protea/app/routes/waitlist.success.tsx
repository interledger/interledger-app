import { Icon, Router } from '~/components'
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
    <div className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto rounded-2xl bg-page px-4 pb-16 pt-6 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:pt-12 xl:max-w-4xl'>
      <div className='col-span-full flex flex-col pt-5 sm:col-span-6 sm:col-start-2'>
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
      <div className='col-span-full flex flex-col space-y-2 pt-11 pb-8 sm:col-span-6 sm:col-start-2'>
        <span className='font-display text-2xl font-medium'>Thank you</span>
        <span className='text-medium'>
          You have successfully joined the waitlist.
        </span>
        <span className='text-medium'>
          We will let you know via email once you are able to transact.
        </span>
      </div>

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2'>
        <Router
          to={route('/')}
          className='flex h-[50px] w-full items-center justify-center rounded-full bg-primary px-10'
        >
          <span className='font-display font-medium text-white'>Close</span>
        </Router>
      </div>
    </div>
  )
}
