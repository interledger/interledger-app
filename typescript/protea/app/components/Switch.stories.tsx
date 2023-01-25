import type { StoryFn, Meta } from '@storybook/react'
import { unstable_createRemixStub as createRemixStub } from '@remix-run/testing'
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
