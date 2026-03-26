import { createRoutesStub } from 'react-router';
import type { Meta, StoryFn } from '@storybook/react'
import { Icon } from '~/components'

// TODO Put over layout so scrim can show.
const meta: Meta<typeof Icon> = {
  title: 'components/Icon',
  component: Icon,
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

const Template: StoryFn<typeof Icon> = (_args) => <Icon {..._args} />

export const IconStory = Template.bind({})
IconStory.storyName = 'Face Icon'
IconStory.args = {
  children: 'face'
}

/**
 * Can change the colour using text colour className.
 */
export const WalletIcon = Template.bind({})
WalletIcon.args = {
  children: 'savings',
  className: 'text-rose-500'
}

export const PayIcon = Template.bind({})
PayIcon.args = {
  children: 'attach_money'
}

/**
 * If the provided text doesn't correspond with an icon.
 */
export const FallbackText = Template.bind({})
FallbackText.args = {
  children: 'not_an_icon'
}
