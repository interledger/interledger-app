import { unstable_createRemixStub as createRemixStub } from '@remix-run/testing'
import type { Meta, StoryFn } from '@storybook/react'
import { FocusLayout, Layouts } from '~/components'

function Home() {
  return (
    <div className='flex w-full rounded-2xl bg-white p-4'>
      <h1 className='text-xl text-strong'>Dummy content</h1>
    </div>
  )
}

let story: Meta<typeof FocusLayout> = {
  title: 'layouts/FocusLayout',
  component: Home,
  parameters: {
    docs: {
      iframeHeight: 500,
      iframeWidth: 700,
      inlineStories: false
    }
  },
  decorators: [
    (Story, { args, parameters }) => {
      let remix = parameters.remix(args)
      const RemixStub = createRemixStub([
        {
          id: 'root', // NOTE id is required because we use useMatches
          element: <FocusLayout />,
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

export default story

const Template: StoryFn<typeof Home> = () => <Home />

export const FocusLayoutStory = Template.bind({})
FocusLayoutStory.parameters = {
  layout: 'fullscreen',
  remix(_args: any) {
    return {
      handle: {
        layout: Layouts.FocusLayout
      },
      loader() {
        return []
      }
    }
  }
}
