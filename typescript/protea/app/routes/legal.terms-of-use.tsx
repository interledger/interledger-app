import type { MetaFunction } from '@remix-run/node';
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { Layouts } from '~/components'

export async function loader() {
  const policy = await fetch(
    'https://www.iubenda.com/api/terms-and-conditions/22630844/no-markup'
  )
  const content = (await policy.json()).content
  return json({ content })
}

export const handle = {
  layout: Layouts.LandingLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Legal | Terms of use'
  }
}

export default function Page() {
  let { content } = useLoaderData<typeof loader>()
  return (
    <main className='w-full overflow-hidden'>
      <section className='relative mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-x-visible py-20 px-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-[59rem]'>
        <div
          style={{
            fontVariant: 'normal'
          }}
          className='prose prose-slate col-span-full prose-h1:font-display prose-h1:font-medium prose-h2:font-display prose-h2:font-medium prose-h3:font-display prose-h3:font-medium prose-h4:font-display prose-h4:font-medium prose-h5:font-display prose-h5:font-medium prose-h6:font-display prose-h6:font-medium prose-a:text-primary prose-a:no-underline prose-blockquote:font-normal prose-strong:font-medium prose-code:font-medium prose-pre:rounded-xl prose-pre:bg-container prose-pre:p-4 prose-pre:pb-6 prose-pre:text-strong prose-img:rounded-xl'
          dangerouslySetInnerHTML={{ __html: content }}
        />
      </section>
    </main>
  )
}
