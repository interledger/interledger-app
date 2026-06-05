import type { Meta, StoryFn } from '@storybook/react'
import { createRoutesStub } from 'react-router'
import { NavDrawer } from './NavDrawer'

const meta: Meta<typeof NavDrawer> = {
  title: 'components/NavDrawer',
  component: NavDrawer,
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

const Template: StoryFn<typeof NavDrawer> = (_args) => <NavDrawer {..._args} />

export const NavDrawerStory = Template.bind({})
NavDrawerStory.storyName = 'Default NavDrawer'
NavDrawerStory.args = {}
