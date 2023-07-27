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

  const [setLoading] = useScaffoldStore((state) => [state.setLoading])

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
        {results.length == 0 && (
          <CardContent>Your search returned no results.</CardContent>
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
                  {result.identifierType == 'wallet' && <FynbosIcon />}
                  {result.identifierType == 'twitter' && <TwitterIcon />}
                  <div className='flex flex-col items-start gap-y-2'>
                    <span className='text-medium'>{result.identifier}</span>
                    {result.subResults?.map((subResult) => {
                      return (
                        <div
                          key={subResult.walletID + subResult.identifier}
                          className='flex justify-start gap-x-1'
                        >
                          {subResult.identifierType == 'wallet' && (
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
    </Combobox>
  )
}
