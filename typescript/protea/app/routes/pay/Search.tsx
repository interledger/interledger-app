import type { PlainMessage } from '@bufbuild/protobuf/dist/types/message'
import { Combobox } from '@headlessui/react'
import { Form, useFetcher, useNavigation } from '@remix-run/react'
import clsx from 'clsx'
import type { ChangeEventHandler } from 'react'
import { useCallback, useEffect, useState } from 'react'
import { route } from 'routes-gen'
import {
  Card,
  CardButton,
  CardContent,
  CardHeader,
  CardTitle,
  DiscordIcon,
  FynbosIcon,
  Icon,
  TextField,
  TwitterIcon
} from '~/components'
import type { SearchResult } from '~/generated/connect/backend/v1/backend_pb'
import { useScaffoldStore } from '~/lib/useScaffoldStore'

import type { searchLoader } from './route'

export function Search() {
  const search = useFetcher<typeof searchLoader>()

  const navigation = useNavigation()

  // We use this to submit the form so that we don't navigate to /pay
  const submit = useFetcher()

  const [term, setTerm] = useState<string>('')
  const [results, setResults] = useState<PlainMessage<SearchResult>[]>([])

  const [pushSnackbar, setLoading] = useScaffoldStore((state) => [
    state.pushSnackbar,
    state.setLoading
  ])

  const _onChangeInput = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      let term = event.target.value
      setTerm(term)
      setLoading(true)
      search.load(`/pay?term=${term}`)
    },
    [search, setLoading]
  )

  useEffect(() => {
    // TODO make this smarter
    if (term.length >= 3) {
      setResults(search.data?.results || [])
    }
    setLoading(false)
  }, [search.data, term.length, setResults, setLoading])

  return (
    <Combobox
      onChange={(result: PlainMessage<SearchResult>) => {
        console.log('Selected an option:', result)
        submit.submit(
          { walletUrl: result.walletUrl },
          {
            action: route('/pay'),
            method: 'POST'
          }
        )
      }}
    >
      <Form
        id='pay-search-form'
        action={route('/pay')}
        method='post'
        className='hidden'
      />
      <Card>
        <Combobox.Input
          as={TextField}
          id='search'
          autoFocus
          form='search-form'
          onBlur={(e) => {
            // This is to stop the form from clearing on blur
            e.preventDefault()
          }}
          name='search'
          value={term}
          placeholder='Search for someone to pay'
          onChange={_onChangeInput}
          prefixIcon={<Icon>search</Icon>}
          type='text'
        />
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Results</CardTitle>
        </CardHeader>
        {results.length == 0 && term.length >= 3 && (
          <CardContent>No results found.</CardContent>
        )}
        {results.length == 0 && term.length < 3 && (
          <CardContent>Type at least 3 characters to search.</CardContent>
        )}
        <Combobox.Options static className='contents w-full'>
          {results.map((result: PlainMessage<SearchResult>) => {
            return (
              <Combobox.Option
                as={CardButton}
                form='pay-search-form'
                key={result.walletID + result.identifier}
                value={result}
                name='walletUrl'
                type='button'
                className={({ active }) =>
                  clsx('items-center gap-x-3', active ? 'bg-nav-hover' : '')
                }
              >
                <input
                  form='pay-search-form'
                  value={result.walletUrl}
                  name='walletUrl'
                  type='hidden'
                />
                <div className='flex gap-x-3'>
                  {(result.identifierType == 'wallet' ||
                    result.identifierType == 'wallet_url') && <FynbosIcon />}
                  {result.identifierType == 'twitter' && <TwitterIcon />}
                  <div className='flex flex-col items-start gap-y-2'>
                    <span className='text-medium'>{result.identifier}</span>
                    <div className='flex flex-wrap gap-x-4'>
                      {result.subResults?.map((subResult) => {
                        return (
                          <div
                            key={subResult.walletID + subResult.identifier}
                            className='flex justify-start gap-x-1'
                          >
                            {(subResult.identifierType == 'wallet' ||
                              subResult.identifierType == 'wallet_url') && (
                              <FynbosIcon />
                            )}
                            {subResult.identifierType == 'twitter' && (
                              <TwitterIcon />
                            )}
                            {subResult.identifierType == 'discord' && (
                              <DiscordIcon />
                            )}
                            {subResult.identifierType == 'domain' && (
                              <Icon>captive_portal</Icon>
                            )}
                            <span className='text-medium'>
                              {subResult.identifier}
                            </span>
                          </div>
                        )
                      })}
                    </div>
                  </div>
                </div>

                <Icon className='ml-auto'>navigate_next</Icon>
              </Combobox.Option>
            )
          })}
        </Combobox.Options>
      </Card>
      {results.length == 0 && term.length >= 3 && (
        <Card>
          <CardContent>
            Share this link with {term} to join Fynbos to transact.
          </CardContent>
          <CardButton
            noHover
            type='button'
            className='items-center justify-between'
            onClick={async () => {
              if (typeof navigator.clipboard == 'undefined') {
                pushSnackbar({
                  id: 'copy-to-clipboard-fail',
                  message: "Couldn't copy to clipboard.",
                  icon: 'close',
                  canShow: true
                })
              } else
                navigator.clipboard.writeText('https://fynbos.app/signup').then(
                  () => {
                    pushSnackbar({
                      id: 'copy-signup-link',
                      message: 'Sign up link copied to clipboard.',
                      icon: 'close',
                      canShow: true
                    })
                  },
                  () => {
                    pushSnackbar({
                      id: 'copy-to-clipboard-fail',
                      message: "Couldn't copy to clipboard.",
                      icon: 'close',
                      canShow: true
                    })
                  }
                )
            }}
          >
            <span className='text-medium'>fynbos.app/signup</span>
            <Icon className={'text-medium'}>content_copy</Icon>
          </CardButton>
        </Card>
      )}
    </Combobox>
  )
}
