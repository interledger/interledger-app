import type { StoryFn, Meta } from '@storybook/react'
import { unstable_createRemixStub as createRemixStub } from '@remix-run/testing'
import { Card } from '~/components'

const meta: Meta<typeof Card> = {
  title: 'components/Card',
  component: Card,
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

const Template: StoryFn<typeof Card> = (_args) => (
  <Card {..._args}>{_args.children}</Card>
)

export const CardStory = Template.bind({})
CardStory.storyName = 'Default Card'
CardStory.args = {
  children: 'Basic Card'
}
