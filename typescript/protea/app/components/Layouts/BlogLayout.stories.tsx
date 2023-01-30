import type { StoryFn, Meta } from '@storybook/react'
import { unstable_createRemixStub as createRemixStub } from '@remix-run/testing'
import { BlogLayout } from './BlogLayout'

const meta: Meta<typeof BlogLayout> = {
  title: 'layouts/BlogLayout',
  component: BlogLayout,
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

const Template: StoryFn<typeof BlogLayout> = (_args) => (
  <BlogLayout {..._args} />
)

export const BlogLayoutStory = Template.bind({})
BlogLayoutStory.storyName = 'Default BlogLayout'
BlogLayoutStory.args = {
  children: 'This is the blog article.',
  meta: {
    title: 'A cool title',
    authors: [
      {
        name: 'Fynbos',
        twitterHandle: 'fynbosdev',
        avatar: '/icon.png'
      }
    ],
    description: 'The blog description',
    date: 'The date',
    slug: 'card-payments-still-suck'
  }
}
