import type { StoryFn, Meta } from '@storybook/react'
import { unstable_createRemixStub as createRemixStub } from '@remix-run/testing'
import { NavDrawer } from './NavDrawer'

const meta: Meta<typeof NavDrawer> = {
  title: 'components/NavDrawer',
  component: NavDrawer,
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

const Template: StoryFn<typeof NavDrawer> = (_args) => <NavDrawer {..._args} />

export const NavDrawerStory = Template.bind({})
NavDrawerStory.storyName = 'Default NavDrawer'
NavDrawerStory.args = {}
