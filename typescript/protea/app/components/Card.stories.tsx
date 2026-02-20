import { createRemixStub as createRemixStub } from '@remix-run/testing'
import type { Meta, StoryFn } from '@storybook/react'
import { Card } from '~/components'

const meta: Meta<typeof Card> = {
  title: 'components/Card',
  component: Card,
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

const Template: StoryFn<typeof Card> = (_args) => (
  <Card {..._args}>{_args.children}</Card>
)

export const CardStory = Template.bind({})
CardStory.storyName = 'Default Card'
CardStory.args = {
  children: 'Basic Card'
}
