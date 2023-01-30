import type { StoryFn, Meta } from '@storybook/react'
import { unstable_createRemixStub as createRemixStub } from '@remix-run/testing'
import { RadioGroup } from '~/components'

// TODO Put over layout so scrim can show.
const meta: Meta<typeof RadioGroup> = {
  title: 'components/RadioGroup',
  component: RadioGroup,
  decorators: [
    (Story) => {
      const RemixStub = createRemixStub([
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
