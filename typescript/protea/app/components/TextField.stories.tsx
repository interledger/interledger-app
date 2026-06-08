import type { Meta, StoryFn } from '@storybook/react'
import { createRoutesStub } from 'react-router'
import { Icon, TextField } from '~/components'

const meta: Meta<typeof TextField> = {
  title: 'components/TextField',
  component: TextField,
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

const Template: StoryFn<typeof TextField> = (_args) => (
  <TextField {..._args}>{_args.children}</TextField>
)

export const TextFieldStory = Template.bind({})
TextFieldStory.storyName = 'Default TextField'
TextFieldStory.args = {
  name: 'something',
  type: 'text',
  label: 'This is an input'
}

/**
 * Js docs get put in storybook
 */
export const WithPrefixIcon = Template.bind({})
WithPrefixIcon.args = {
  name: 'something',
  type: 'text',
  label: 'This is an input',
  prefixIcon: <Icon>face</Icon>
}

export const WithPrefix = Template.bind({})
WithPrefix.args = {
  name: 'something',
  type: 'text',
  label: 'This is an input',
  prefix: '$'
}
