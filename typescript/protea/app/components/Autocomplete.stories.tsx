import { createRemixStub as createRemixStub } from '@remix-run/testing'
import type { Meta, StoryFn } from '@storybook/react'
import { Autocomplete, Icon } from '~/components'

const meta: Meta<typeof Autocomplete> = {
  title: 'components/Autocomplete',
  component: Autocomplete,
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

const Template: StoryFn<typeof Autocomplete> = (_args) => (
  <Autocomplete {..._args} />
)

export const AutocompleteStory = Template.bind({})
AutocompleteStory.storyName = 'Default Autocomplete'
AutocompleteStory.args = {
  label: 'This is a label',
  options: [
    { id: '1', name: 'first' },
    { id: '2', name: 'second' },
    { id: '3', name: 'third' }
  ]
}

export const WithPrefixIcon = Template.bind({})
WithPrefixIcon.args = {
  label: 'This is a label',
  prefixIcon: <Icon>search</Icon>,
  options: [
    { id: '1', name: 'first' },
    { id: '2', name: 'second' },
    { id: '3', name: 'third' }
  ]
}

export const WithPrefix = Template.bind({})
WithPrefix.args = {
  label: 'This is a label',
  prefix: '$',
  options: [
    { id: '1', name: 'first' },
    { id: '2', name: 'second' },
    { id: '3', name: 'third' }
  ]
}

export const WithoutButton = Template.bind({})
WithoutButton.args = {
  label: 'This is a label',
  button: false,
  options: [
    { id: '1', name: 'first' },
    { id: '2', name: 'second' },
    { id: '3', name: 'third' }
  ]
}
