import { unstable_createRemixStub as createRemixStub } from '@remix-run/testing'
import type { Meta, StoryFn } from '@storybook/react'
import { Layouts, WalletLayout } from '~/components'

function Home() {
  return (
    <div className='flex w-full rounded-2xl bg-white p-4'>
      <h1 className='text-xl text-strong'>Dummy content</h1>
    </div>
  )
}

let meta: Meta<typeof WalletLayout> = {
  title: 'layouts/WalletLayout',
  component: Home,
  parameters: {
    docs: {
      iframeHeight: 500,
      inlineStories: false
    }
  },
  decorators: [
    (Story, { args, parameters }) => {
      let remix = parameters.remix(args)
      const RemixStub = createRemixStub([
        {
          id: 'root', // NOTE id is required because we use useMatches
          element: <WalletLayout />,
          children: [
            {
              id: 'Nested',
              path: '/login',
              handle: remix.handle,
              // @ts-ignore
              element: <Story />
            }
          ]
        }
      ])

      return <RemixStub initialEntries={['/login']} />
    }
  ]
}

export default meta

const Template: StoryFn<typeof Home> = () => <Home />

export const WalletLayoutStory = Template.bind({})
WalletLayoutStory.parameters = {
  layout: 'fullscreen',
  remix(_args: any) {
    return {
      handle: {
        layout: Layouts.WalletLayout
      },
      loader() {
        return []
      }
    }
  }
}
