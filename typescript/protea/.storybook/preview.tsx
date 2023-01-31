import '~/styles/app.css'
import '~/styles/flags.css'

export const decorators = [
  (Story) => (
    <div className='theme-blue flex font-sans text-base font-normal text-strong antialiased selection:bg-brand/50'>
      <Story />
    </div>
  )
]

export const parameters = {
  actions: { argTypesRegex: '^on.*' },
  backgrounds: {
    default: 'app',
    values: [
      { name: 'app', value: '#F8FAFC' },
      { name: 'pink', value: '#f0f' }
    ]
  }
}
