import { createRoutesStub } from 'react-router';
import type { Meta, StoryFn } from '@storybook/react'
import { Error } from '~/components'

// TODO Put over layout so scrim can show.
const meta: Meta<typeof Error> = {
  title: 'components/Error',
  component: Error,
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

const Template: StoryFn<typeof Error> = (_args) => <Error {..._args} />

export const ErrorStory = Template.bind({})
ErrorStory.storyName = 'Default Error'
ErrorStory.args = {}
