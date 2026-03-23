import { createRoutesStub } from 'react-router';
import type { Meta, StoryFn } from '@storybook/react'
import { Select } from '~/components'

const meta: Meta<typeof Select> = {
  title: 'components/Select',
  component: Select,
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
