import { createRemixStub as createRemixStub } from '@remix-run/testing'
import type { Meta, StoryFn } from '@storybook/react'
import { NavDrawer } from './NavDrawer'

const meta: Meta<typeof NavDrawer> = {
  title: 'components/NavDrawer',
  component: NavDrawer,
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

const Template: StoryFn<typeof NavDrawer> = (_args) => <NavDrawer {..._args} />

export const NavDrawerStory = Template.bind({})
NavDrawerStory.storyName = 'Default NavDrawer'
NavDrawerStory.args = {}
