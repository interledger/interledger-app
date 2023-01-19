import type { StoryFn, Meta } from '@storybook/react'
import { unstable_createRemixStub as createRemixStub } from '@remix-run/testing'
import { TextArea } from '~/components'

// TODO Put over layout so scrim can show.
const meta: Meta<typeof TextArea> = {
  title: 'components/TextArea',
  component: TextArea,
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

const Template: StoryFn<typeof TextArea> = (_args) => <TextArea {..._args} />

export const TextAreaStory = Template.bind({})
TextAreaStory.storyName = 'Default TextArea'
TextAreaStory.args = {
  children: 'Something'
}
