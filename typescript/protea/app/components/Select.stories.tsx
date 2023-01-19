import type { StoryFn, Meta } from '@storybook/react'
import { unstable_createRemixStub as createRemixStub } from '@remix-run/testing'
import { Select } from '~/components'

const meta: Meta<typeof Select> = {
  title: 'components/Select',
  component: Select,
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

const Template: StoryFn<typeof Select> = (_args) => <Select {..._args} />

export const SelectStory = Template.bind({})
SelectStory.storyName = 'Default Select'
SelectStory.args = {
  label: 'This is a label',
  options: [
    { id: '1', name: 'first' },
    { id: '2', name: 'second' },
    { id: '3', name: 'third' }
  ]
}
