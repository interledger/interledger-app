import type { Meta, StoryFn } from '@storybook/react'
import { createRoutesStub } from 'react-router'
import { Card } from '~/components'

const meta: Meta<typeof Card> = {
  title: 'components/Card',
  component: Card,
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

const Template: StoryFn<typeof Card> = (_args) => (
  <Card {..._args}>{_args.children}</Card>
)

export const CardStory = Template.bind({})
CardStory.storyName = 'Default Card'
CardStory.args = {
  children: 'Basic Card'
}
