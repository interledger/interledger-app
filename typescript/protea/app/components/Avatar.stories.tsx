import { createRoutesStub } from 'react-router';
import type { Meta, StoryFn } from '@storybook/react'
import { Avatar } from '~/components'

const meta: Meta<typeof Avatar> = {
  title: 'components/Avatar',
  component: Avatar,
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

const Template: StoryFn<typeof Avatar> = (_args) => (
  <Avatar {..._args}>{_args.children}</Avatar>
)

export const DefaultAvatar = Template.bind({})
DefaultAvatar.storyName = 'Rose Avatar'
DefaultAvatar.args = {
  color: 'rose',
  children: 'R'
}

export const GreenAvatar = Template.bind({})
GreenAvatar.args = {
  color: 'green',
  children: 'G'
}

export const PurpleAvatar = Template.bind({})
PurpleAvatar.args = {
  color: 'purple',
  children: 'P'
}

export const OrangeAvatar = Template.bind({})
OrangeAvatar.args = {
  color: 'orange',
  children: 'O'
}

export const YellowAvatar = Template.bind({})
YellowAvatar.args = {
  color: 'yellow',
  children: 'Y'
}

export const SkyAvatar = Template.bind({})
SkyAvatar.args = {
  color: 'sky',
  children: 'S'
}

export const IndigoAvatar = Template.bind({})
IndigoAvatar.args = {
  color: 'indigo',
  children: 'I'
}
