import { createRemixStub as createRemixStub } from '@remix-run/testing'
import type { Meta, StoryFn } from '@storybook/react'
import { Checkbox } from '~/components'

const meta: Meta<typeof Checkbox> = {
  title: 'components/Checkbox',
  component: Checkbox,
  argTypes: { onClick: { action: 'clicked' } },
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

const Template: StoryFn<typeof Checkbox> = (_args) => (
  <Checkbox {..._args}>{_args.children}</Checkbox>
)

export const Default = Template.bind({})
Default.args = {
  className: 'flex',
  children:
    'I authorize the Interledger Wallet to debit the card indicated for the amount noted on today’s date. I will not dispute Interledger Wallet debiting my account, so long as the transaction corresponds to the terms in this online form and my agreement with the Interledger Wallet.'
}

export const WithError = Template.bind({})
WithError.args = {
  className: 'flex',
  errorMessage: 'This is an error message.',
  children:
    'I authorize the Interledger Wallet to debit the card indicated for the amount noted on today’s date. I will not dispute Interledger Wallet debiting my account, so long as the transaction corresponds to the terms in this online form and my agreement with the Interledger Wallet.'
}
