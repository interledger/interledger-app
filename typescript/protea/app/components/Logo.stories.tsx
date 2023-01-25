import type { StoryFn, Meta } from '@storybook/react'
import { unstable_createRemixStub as createRemixStub } from '@remix-run/testing'
import { Logo } from '~/components'

const meta: Meta<typeof Logo> = {
  title: 'components/Logo',
  component: Logo,
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

const Template: StoryFn<typeof Logo> = (_args) => <Logo {..._args} />

export const LogoStory = Template.bind({})
LogoStory.storyName = 'Default Logo'
LogoStory.args = {}

export const SetHeight = Template.bind({})
SetHeight.args = {
  className: 'h-6'
}
