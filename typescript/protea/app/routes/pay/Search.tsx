import { useFetcher } from '@remix-run/react'
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
import { PayStep, usePayStore } from '~/store'

export function Search() {
  const search = useFetcher()
  const [term, setTerm] = useState<string>('')
  const [results, setResults] = useState<SearchResult[]>([])
  const [setStep, setAddress] = usePayStore((state) => [
    state.setStep,
    state.setAddress
  ])

  const _onChangeInput = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      let term = event.target.value
      setTerm(term)
      search.load(`/pay?term=${term}`)
    },
    [search]
  )

  useEffect(() => {
    if (term.length >= 3) {
      setResults(search.data?.results || [])
    }
  }, [search.data, term.length, setResults])

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
      <Card>
        <TextField
          id='search'
          autoFocus
          form='search-form'
          name='search'
          defaultValue={term}
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
              key={result.walletID + result.identifier}
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
