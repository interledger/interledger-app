import type { Meta, StoryFn } from '@storybook/react'
import { createRoutesStub } from 'react-router'
import { Switch } from '~/components'

// TODO Put over layout so scrim can show.
const meta: Meta<typeof Switch> = {
  title: 'components/Switch',
  component: Switch,
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

const Template: StoryFn<typeof Switch> = (_args) => <Switch {..._args} />

export const SwitchStory = Template.bind({})
SwitchStory.storyName = 'Default Switch'
SwitchStory.args = {}
