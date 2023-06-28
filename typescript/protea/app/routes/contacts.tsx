import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import type { ShouldRevalidateFunction } from '@remix-run/react'
import {
  Form,
  useFetcher,
  useLoaderData,
  useSearchParams
} from '@remix-run/react'
import { useCallback, useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Avatar, Card, CardButton, CardContent, Layouts } from '~/components'
import type { ListContactsResponse } from '~/generated/protobuf-ts/backend/v1/backend'
import { getWalletContacts } from '~/lib/wallet.server'

/**
 * Allows us to change the searchParams without revalidating the pages data
 * This is useful for pagination.
 */
export const shouldRevalidate: ShouldRevalidateFunction = ({
  currentUrl,
  defaultShouldRevalidate,
  nextUrl
}) => {
  if (currentUrl.search !== nextUrl.search) return false
  return defaultShouldRevalidate
}

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)
  const pages = parseInt(url.searchParams.get('pages') || '1')

  let pageInfo = {
    // pageToken is only set by fetcher so initial page loads this should be blank
    pageToken: url.searchParams.get('pageToken') || '',
    pageSize: 30
  }

  let allContacts: ListContactsResponse['contacts'] = []

  /**
   * We can loop over pages as fetcher should omit this.
   * This allows us to pull all necessary data when navigating back to this page.
   */
  for (let i = 0; i < pages; i++) {
    const { contacts, nextPageToken } = await getWalletContacts(
      request,
      pageInfo
    )
    pageInfo.pageToken = nextPageToken
    allContacts = [...allContacts, ...contacts]
    if (nextPageToken == '') break
  }

  // const contacts = (await getWalletContacts(request, { pageSize: 3 })).contacts

  return json({ contacts: allContacts, nextPageToken: pageInfo.pageToken })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/'),
      title: 'Contacts'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Contacts'
  }
}

export default function Page() {
  const initialPage = useLoaderData<typeof loader>()
  let [, setSearchParams] = useSearchParams()
  const fetcher = useFetcher()
  const [contacts, setContacts] = useState(initialPage.contacts)
  const [nextPageToken, setNextPageToken] = useState<string>(
    initialPage.nextPageToken
  )
  const [scrollPosition, setScrollPosition] = useState(0)
  const [clientHeight, setClientHeight] = useState(0)
  const [height, setHeight] = useState(null)
  const [shouldFetch, setShouldFetch] = useState(true)

  const divHeight = useCallback(
    (node: any) => {
      if (node !== null) {
        setHeight(node.getBoundingClientRect().height)
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [contacts?.length]
  )

  // Add Listeners to scroll and client resize
  useEffect(() => {
    const scrollListener = () => {
      setClientHeight(window.innerHeight)
      setScrollPosition(window.scrollY)
    }

    // Avoid running during SSR
    if (typeof window !== 'undefined') {
      window.addEventListener('scroll', scrollListener)
    }

    // Clean up
    return () => {
      if (typeof window !== 'undefined') {
        window.removeEventListener('scroll', scrollListener)
      }
    }
  }, [clientHeight, scrollPosition])

  // Trigger fetching new data when the user scrolls near the bottom
  useEffect(() => {
    if (!shouldFetch || !height) return
    if (clientHeight + scrollPosition + 100 < height) return

    if (nextPageToken != '') {
      fetcher.load(`/contacts?pageToken=${nextPageToken}`)
      setSearchParams(
        (old) => {
          old.set('pages', `${parseInt(old.get('pages') || '1') + 1}`)
          return old
        },
        { replace: true, preventScrollReset: true }
      )
    }

    setShouldFetch(false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [clientHeight, scrollPosition])

  // Handle new contacts being fetched and add to contact list
  useEffect(() => {
    if (fetcher.data && fetcher.data.contacts.length > 0) {
      setContacts((currentContacts) => [
        ...currentContacts,
        ...fetcher.data.contacts
      ])

      setNextPageToken(fetcher.data.nextPageToken || '')
      setShouldFetch(true)
    }
  }, [fetcher.data])

  return (
    <div ref={divHeight}>
      <Card>
        <Form
          id='pay-payment-pointer'
          action={route('/pay')}
          method='post'
          className='hidden'
        />
        {contacts.length == 0 && (
          <CardContent>
            <p className='text-sm text-medium'>You haven't paid anyone yet.</p>
          </CardContent>
        )}
        {contacts &&
          contacts.map((contact, index) => (
            <CardButton
              key={contact.id}
              name='paymentPointer'
              form='pay-payment-pointer'
              value={contact.paymentPointer}
              className='items-center space-x-3'
            >
              <Avatar index={index}>{contact.name.charAt(0)}</Avatar>
              <span className='text-medium'>{contact.name}</span>
            </CardButton>
          ))}
      </Card>
    </div>
  )
}
