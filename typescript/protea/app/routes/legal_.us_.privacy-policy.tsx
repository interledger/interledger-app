import type { MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { Layouts } from '~/components'
import { fetchAndSanitizeHTML } from '~/lib/fetchAndSanitizeHTML'

export async function loader() {
  const content = await fetchAndSanitizeHTML(
    'https://gmtsend.com/en/privacy-policy'
  )

  return json({ content })
}

export const handle = {
  layout: Layouts.LandingLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Legal | GMT privacy policy'
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
          className='prose-h5:font-display prose-h5:font-medium prose-h6:font-display prose-h6:font-medium prose prose-slate col-span-full prose-h1:font-display prose-h1:font-medium prose-h2:font-display prose-h2:font-medium prose-h3:font-display prose-h3:font-medium prose-h4:font-display prose-h4:font-medium prose-a:text-primary prose-a:no-underline prose-a:focus-visible:outline-2 prose-a:focus-visible:outline-focus prose-blockquote:border-0 prose-blockquote:p-0 prose-blockquote:text-3xl prose-blockquote:font-normal prose-blockquote:not-italic prose-code:font-normal prose-code:tracking-wider prose-pre:rounded-xl prose-pre:bg-slate-800 prose-pre:p-4 prose-pre:pb-6 dark:prose-invert'
          dangerouslySetInnerHTML={{ __html: content }}
        />
      </section>
    </main>
  )
}
