import { unstable_createRemixStub as createRemixStub } from '@remix-run/testing'
import type { Meta, StoryFn } from '@storybook/react'
import { Snackbar } from '~/components'

const meta: Meta<typeof Snackbar> = {
  title: 'components/Snackbar',
  component: Snackbar,
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

const Template: StoryFn<typeof Snackbar> = (_args) => <Snackbar {..._args} />

export const WithIcon = Template.bind({})
WithIcon.args = {
  message: 'Something to say?',
  icon: 'close',
  show: true
}

export const WithAction = Template.bind({})
WithAction.args = {
  message: "I don't have an icon :(",
  action: 'close',
  show: true
}

export const HidesItself = Template.bind({})
HidesItself.args = {
  message: 'I should trigger onClose automatically.',
  icon: 'close',
  show: true,
  dismissAfter: 3000
}

export const Offset = Template.bind({})
Offset.args = {
  message: 'To the right, to the right.',
  icon: 'start',
  show: true,
  offset: true
}
