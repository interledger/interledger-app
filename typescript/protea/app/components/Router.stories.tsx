import { createRemixStub as createRemixStub } from '@remix-run/testing'
import type { Meta, StoryFn } from '@storybook/react'
import { Router } from '~/components'

// TODO Put over layout so scrim can show.
const meta: Meta<typeof Router> = {
  title: 'components/Router',
  component: Router,
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

const Template: StoryFn<typeof Router> = (_args) => <Router {..._args} />

export const RouterStory = Template.bind({})
RouterStory.storyName = 'Default Router'
RouterStory.args = {
  children: 'Something'
}
