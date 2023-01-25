import type { StoryFn, Meta } from '@storybook/react'
import { unstable_createRemixStub as createRemixStub } from '@remix-run/testing'
import { PhoneTextField } from '~/components'

// TODO Put over layout so scrim can show.
const meta: Meta<typeof PhoneTextField> = {
  title: 'components/PhoneTextField',
  component: PhoneTextField,
  decorators: [
    (Story) => {
      const RemixStub = createRemixStub([
        {
          path: '/',
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
  children: 'Something'
}
