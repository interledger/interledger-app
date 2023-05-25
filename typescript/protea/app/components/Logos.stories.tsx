import type { StoryFn, Meta } from '@storybook/react'
import { unstable_createRemixStub as createRemixStub } from '@remix-run/testing'
import { FynbosLogo } from '~/components'

const meta: Meta<typeof FynbosLogo> = {
  title: 'components/Logo',
  component: FynbosLogo,
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

const Template: StoryFn<typeof FynbosLogo> = (_args) => (
  <FynbosLogo {..._args} />
)

export const LogoStory = Template.bind({})
LogoStory.storyName = 'Default Logo'
LogoStory.args = {}

export const SetHeight = Template.bind({})
SetHeight.args = {
  className: 'h-6'
}
