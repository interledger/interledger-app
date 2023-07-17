import { useFetcher } from '@remix-run/react'
import type { ChangeEventHandler } from 'react'
import { useCallback, useEffect } from 'react'
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
import { PayStep, useStore } from '~/store'

export function Search() {
  const fetcher = useFetcher()
  const [setStep, searchTerm, setSearchTerm, setAddress, results, setResults] =
    useStore((state) => [
      state.setStep,
      state.searchTerm,
      state.setSearchTerm,
      state.setAddress,
      state.results,
      state.setResults
    ])

  const _onChangeInput = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      let term = event.target.value
      setSearchTerm(term)
      fetcher.load(`/pay?term=${term}`)
    },
    [fetcher, setSearchTerm]
  )

  useEffect(() => {
    if (searchTerm.length >= 3) {
      setResults(fetcher.data?.results || [])
    }
  }, [fetcher.data, searchTerm.length, setResults])

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
    <>
      {/*<fetcher.Form*/}
      {/*  id='search-form'*/}
      {/*  action={route('/pay')}*/}
      {/*  method='post'*/}
      {/*  className='hidden'*/}
      {/*/>*/}
      <Card>
        <TextField
          id='search'
          form='search-form'
          name='search'
          defaultValue={searchTerm}
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
        {results.map((result: SearchResult) => {
          return (
            <CardButton
              key={result.walletID}
              onClick={() => _onClickResult(result)}
              name='address'
              type='button'
              className='items-center space-x-3'
            >
              {result.identifierType == 'wallet' && <FynbosIcon />}
              {result.identifierType == 'twitter' && <TwitterIcon />}
              <span className='text-medium'>{result.identifier}</span>
            </CardButton>
          )
        })}
      </Card>
    </>
  )
}
