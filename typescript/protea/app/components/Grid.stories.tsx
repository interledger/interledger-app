import { createRoutesStub } from 'react-router';
import type { Meta, StoryFn } from '@storybook/react'
import { Card, WalletGrid } from '~/components'

const meta: Meta<typeof WalletGrid> = {
  title: 'components/WalletGrid',
  component: WalletGrid,
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

const Template: StoryFn<typeof WalletGrid> = (_args) => (
  <WalletGrid {..._args} />
)

export const WalletGridStory = Template.bind({})
WalletGridStory.storyName = 'WalletGrid'
WalletGridStory.args = {
  children: (
    <>
      <Card />
      <Card />
      <Card />
      <Card />
      <Card />
      <Card />
      <Card />
      <Card />
      <Card />
      <Card />
      <Card />
      <Card />
      <Card className='col-start-2' />
      <Card className='col-span-3 row-span-2' />
      <Card className='col-span-2' />
      <Card className='col-span-3 lg:col-span-5' />
      <Card className='col-span-2 col-start-3' />
      <Card className='col-span-2' />
      <Card className='col-span-full' />
    </>
  )
}
