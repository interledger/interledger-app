import { createRoutesStub } from 'react-router';
import type { Meta, StoryFn } from '@storybook/react'
import { Chip, ChipColor } from '~/components'

const meta: Meta<typeof Chip> = {
  title: 'components/Chip',
  component: Chip,
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

const Template: StoryFn<typeof Chip> = (_args) => (
  <Chip {..._args}>{_args.children}</Chip>
)

export const DefaultChip = Template.bind({})
DefaultChip.storyName = 'Chip'
DefaultChip.args = {
  color: ChipColor.blue,
  children: "I'm the default chip"
}

export const GreenChip = Template.bind({})
GreenChip.args = {
  color: ChipColor.green,
  children: "I'm green"
}

export const PurpleChip = Template.bind({})
PurpleChip.args = {
  color: ChipColor.purple,
  children: "I'm purple"
}

export const OrangeChip = Template.bind({})
OrangeChip.args = {
  color: ChipColor.orange,
  children: "I'm orange"
}

export const YellowChip = Template.bind({})
YellowChip.args = {
  color: ChipColor.yellow,
  children: "I'm yellow"
}
