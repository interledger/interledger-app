import type { Meta, StoryFn } from '@storybook/react'
import { createRoutesStub } from 'react-router'
import { Button } from '~/components'

const meta: Meta<typeof Button> = {
  title: 'components/Button',
  component: Button,
  argTypes: { onClick: { action: 'clicked' } },
  decorators: [
    (Story) => {
      const RemixStub = createRoutesStub([
        {
          path: '/',
          // @ts-ignore TODO Remove once fixed
          element: <Story />
        }
      ])

      return <RemixStub />
    }
  ]
}

export default meta

const Template: StoryFn<typeof Button> = (_args) => (
  <Button {..._args}>{_args.children}</Button>
)

export const ButtonStory = Template.bind({})
ButtonStory.storyName = 'Default Button'
ButtonStory.args = {
  children: 'Basic button'
}

export const ShrunkButton = Template.bind({})
ShrunkButton.args = {
  children: 'Another',
  shrink: true
}
