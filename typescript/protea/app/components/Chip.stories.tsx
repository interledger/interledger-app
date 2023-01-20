import type { Meta, StoryFn } from '@storybook/react'
import { unstable_createRemixStub as createRemixStub } from '@remix-run/testing'
import { Chip, ChipColor } from '~/components'

const meta: Meta<typeof Chip> = {
  title: 'components/Chip',
  component: Chip,
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

const Template: StoryFn<typeof Chip> = (_args) => (
  <Chip {..._args}>{_args.children}</Chip>
)

export const DefaultChip = Template.bind({})
DefaultChip.storyName = 'Chip'
DefaultChip.args = {
  color: ChipColor.green,
  children: "I'm green."
}
