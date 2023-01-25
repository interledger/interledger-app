import type { StoryFn, Meta } from '@storybook/react'
import { unstable_createRemixStub as createRemixStub } from '@remix-run/testing'
import { Error } from '~/components'

// TODO Put over layout so scrim can show.
const meta: Meta<typeof Error> = {
  title: 'components/Error',
  component: Error,
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

const Template: StoryFn<typeof Error> = (_args) => <Error {..._args} />

export const ErrorStory = Template.bind({})
ErrorStory.storyName = 'Default Error'
ErrorStory.args = {}
