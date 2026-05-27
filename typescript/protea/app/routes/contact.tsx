import type { Route } from './+types/contact'
import type { UIMatch } from 'react-router';
import { RootLoaderData } from '~/root'
import { useLoaderData, useRouteLoaderData } from 'react-router'
import type { ApplicationProps } from '~/components'
import { AnchorRouter, Icon, Layouts } from '~/components'
import { MarketingPageWithSections } from '~/components/Content'
import { getContactRoute } from '~/data/content.server'
import type { SectionRecord } from '~/generated/dato-cms-graphql'
import { jsonWithCSRF } from '~/lib/csrf.server'

export async function loader({ request }: Route.LoaderArgs) {
  const { contactRoute, footer } = await getContactRoute()

  return jsonWithCSRF(request, { contactRoute, footer })
}

export const handle: ApplicationProps = {
  layout: Layouts.Marketing,
  scaffold: {
    header: {},
    footer: (match: UIMatch<Route.ComponentProps['loaderData']>) => match.loaderData?.footer ?? null
  }
}

export default function Page() {
  const { contactRoute } = useLoaderData()
  const { env } = useRouteLoaderData('root') as RootLoaderData
  const supportEmail = env.supportEmail

  return (
    <>
      {contactRoute?.body.map((section: import("~/generated/dato-cms-graphql").SectionRecord) => (
        <MarketingPageWithSections
          key={section.id}
          section={section as SectionRecord}
        >
          <div className='grid w-full grid-cols-12 gap-y-6 px-4 lg:px-0'>
            <div className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'>
              <h2 className='font-display text-2xl font-medium'>
                Send us a message
              </h2>
            </div>
            <div className='col-span-full mt-10 flex flex-col justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
              <h2 className='font-display font-medium text-strong'>Support</h2>
              <div className='mt-2 flex items-center space-x-2 text-medium'>
                <Icon>mail</Icon>
                <AnchorRouter
                  to={`mailto:${supportEmail}`}
                  className='text-sm text-primary'
                >
                  {supportEmail}
                </AnchorRouter>
              </div>
            </div>
          </div>
        </MarketingPageWithSections>
      ))}
    </>
  )
}
