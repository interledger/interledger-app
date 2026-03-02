import { createRoutesStub } from 'react-router';
import type { Meta, StoryFn } from '@storybook/react'
import { TextArea } from '~/components'

// TODO Put over layout so scrim can show.
const meta: Meta<typeof TextArea> = {
  title: 'components/TextArea',
  component: TextArea,
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

const Template: StoryFn<typeof TextArea> = (_args) => <TextArea {..._args} />

export const TextAreaStory = Template.bind({})
TextAreaStory.storyName = 'Default TextArea'
TextAreaStory.args = {
  children: 'Something'
}
