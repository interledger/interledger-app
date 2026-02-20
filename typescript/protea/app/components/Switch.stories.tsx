import { createRemixStub as createRemixStub } from '@remix-run/testing'
import type { Meta, StoryFn } from '@storybook/react'
import { Switch } from '~/components'

// TODO Put over layout so scrim can show.
const meta: Meta<typeof Switch> = {
  title: 'components/Switch',
  component: Switch,
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

const Template: StoryFn<typeof Switch> = (_args) => <Switch {..._args} />

export const SwitchStory = Template.bind({})
SwitchStory.storyName = 'Default Switch'
SwitchStory.args = {}
