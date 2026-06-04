import type { Meta, StoryFn } from '@storybook/react'
import { createRoutesStub } from 'react-router'
import { PhoneTextField } from '~/components'

// TODO Put over layout so scrim can show.
const meta: Meta<typeof PhoneTextField> = {
  title: 'components/PhoneTextField',
  component: PhoneTextField,
  decorators: [
    (Story) => {
      const RemixStub = createRoutesStub([
        {
          path: '/',
          // @ts-ignore
          element: <Story />
        }
      ])

      return <RemixStub />
    }
  ]
}

export default meta

const Template: StoryFn<typeof PhoneTextField> = (_args) => (
  <PhoneTextField {..._args} />
)

export const PhoneTextFieldStory = Template.bind({})
PhoneTextFieldStory.storyName = 'Default PhoneTextField'
PhoneTextFieldStory.args = {
  label: 'Phone',
  defaultCountry: 'US',
  options: [
    { id: 'US', name: 'United States' },
    { id: 'ZA', name: 'South Africa' }
  ]
}
