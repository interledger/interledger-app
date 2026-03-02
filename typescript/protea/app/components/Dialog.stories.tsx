import { createRoutesStub } from 'react-router';
import type { Meta, StoryFn } from '@storybook/react'
import { Dialog } from '~/components'

// TODO Put over layout so scrim can show.
const meta: Meta<typeof Dialog> = {
  title: 'components/Dialog',
  component: Dialog,
  parameters: {
    docs: {
      iframeHeight: 500,
      iframeWidth: 700,
      inlineStories: false
    }
  },
  decorators: [
    (Story) => {
      const RemixStub = createRoutesStub([
        {
          id: '',
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

const Template: StoryFn<typeof Dialog> = (_args) => <Dialog {..._args} />

export const DialogStory = Template.bind({})
DialogStory.storyName = 'Default Dialog'
DialogStory.args = {
  children: 'Something',
  open: true,
  setOpen: (value) => {}
}
