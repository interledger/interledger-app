import { Combobox } from '@headlessui/react'
import { useFetcher } from '@remix-run/react'
import clsx from 'clsx'
import type { ChangeEventHandler } from 'react'
import { useCallback, useEffect, useState } from 'react'
import {
  Card,
  CardButton,
  CardContent,
  CardHeader,
  CardTitle,
  FynbosIcon,
  Icon,
  TextField,
  TwitterIcon
} from '~/components'
import type { SearchResult } from '~/generated/protobuf-ts/backend/v1/backend'
import { PayStep, usePayStore } from '~/lib/usePayStore'
import { useScaffoldStore } from '~/lib/useScaffoldStore'

export function Search() {
  const search = useFetcher()
  const [term, setTerm] = useState<string>('')
  const [results, setResults] = useState<SearchResult[]>([])
  const [setStep, setAddress] = usePayStore((state) => [
    state.setStep,
    state.setAddress
  ])

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
    if (term.length >= 3) {
      setResults(search.data?.results || [])
    }
    setLoading(false)
  }, [search.data, term.length, setResults, setLoading])

  const _onClickResult = useCallback<{
    (result: SearchResult): void
  }>(
    (result) => {
      setAddress(result)
      setStep(PayStep.AMOUNT)
    },
    [setAddress, setStep]
  )

  return (
    <Combobox onChange={_onClickResult}>
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
          {results.map((result: SearchResult) => {
            return (
              <Combobox.Option
                as={CardButton}
                key={result.walletID + result.identifier}
                value={result}
                onClick={() => _onClickResult(result)}
                name='address'
                type='button'
                className={({ active }) =>
                  clsx('items-center gap-x-3', active ? 'bg-nav-hover' : '')
                }
              >
                <div className='flex gap-x-3'>
                  {(result.identifierType == 'wallet' ||
                    result.identifierType == 'wallet_url') && <FynbosIcon />}
                  {result.identifierType == 'twitter' && <TwitterIcon />}
                  <div className='flex flex-col items-start gap-y-2'>
                    <span className='text-medium'>{result.identifier}</span>
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
                          <span className='text-medium'>
                            {subResult.identifier}
                          </span>
                        </div>
                      )
                    })}
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
