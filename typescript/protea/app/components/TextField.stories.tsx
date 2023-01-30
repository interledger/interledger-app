import type { StoryFn, Meta } from '@storybook/react'
import { unstable_createRemixStub as createRemixStub } from '@remix-run/testing'
import { Icon, TextField } from '~/components'

const meta: Meta<typeof TextField> = {
  title: 'components/TextField',
  component: TextField,
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
export const TextFieldWithPrefixIcon = Template.bind({})
TextFieldWithPrefixIcon.args = {
  name: 'something',
  type: 'text',
  label: 'This is an input',
  prefixIcon: <Icon>face</Icon>
}
