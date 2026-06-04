import type { Meta, StoryFn } from '@storybook/react'
import { createRoutesStub } from 'react-router'
import { RadioGroup } from '~/components'

// TODO Put over layout so scrim can show.
const meta: Meta<typeof RadioGroup> = {
  title: 'components/RadioGroup',
  component: RadioGroup,
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

const Template: StoryFn<typeof RadioGroup> = (_args) => (
  <RadioGroup {..._args} />
)

export const RadioGroupStory = Template.bind({})
RadioGroupStory.storyName = 'Default RadioGroup'
RadioGroupStory.args = {
  label: 'Group',
  options: [
    { id: 'something', name: 'Hello', icon: 'waving_hand' },
    { id: 'somethingelse', name: 'Good bye', icon: 'potted_plant' }
  ]
}
