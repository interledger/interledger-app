import type { PlainMessage } from '@bufbuild/protobuf'
import { Combobox } from '@headlessui/react'
import { Form, useFetcher, useNavigate } from 'react-router';
import clsx from 'clsx'
import type { ChangeEventHandler } from 'react'
import { useCallback, useEffect, useState } from 'react'
import { href } from 'react-router'
import {
  Card,
  CardButton,
  CardContent,
  DiscordIcon,
  Icon,
  InterledgerIcon,
  TextField,
  TwitterIcon
} from '~/components'
import { Label } from '~/components/Label'
import type { SearchResult } from '~/generated/connect/backend/v1/backend_pb'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { searchLoader } from '~/routes/pay/route';


type Action = {
  icon: string
  title: string
  kbd: string
  route: string
}

function isAction(action: unknown): action is Action {
  return (
    (action as Action).icon !== undefined &&
    (action as Action).title !== undefined &&
    (action as Action).kbd !== undefined &&
    (action as Action).route !== undefined
  )
}

// TODO: WE could filter these based on location if the list becomes too long
const defaultActions = [
  {
    icon: 'south_west',
    title: 'Deposit',
    kbd: 'D',
    route: href('/deposit')
  },
  {
    icon: 'north_east',
    title: 'Withdraw',
    kbd: 'W',
    route: href('/withdraw')
  }
]

export function CommandActions() {
  const search = useFetcher<typeof searchLoader>()

  const navigate = useNavigate()

  // We use this to submit the form so that we don't navigate to /pay
  const submit = useFetcher()

  const [term, setTerm] = useState<string>('')
  const [results, setResults] = useState<PlainMessage<SearchResult>[]>([])
  const [actions, setActions] = useState<Action[]>(defaultActions)

  const [setLoading, setCommandPaletteOpen] = useScaffoldStore((state) => [
    state.setLoading,
    state.setCommandPalletOpen
  ])

  const _onChangeInput = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      let term = event.target.value
      setTerm(term)
      setLoading(true)
      search.load(`/pay?term=${term}`)

      const newActions = defaultActions.filter((action) =>
        action.title.toLowerCase().includes(term.toLowerCase())
      )
      setActions(newActions)
    },
    [search, setLoading]
  )

  const _onChangeCombobox = useCallback<
    (value: PlainMessage<SearchResult> | Action) => void
  >(
    (event) => {
      if (isAction(event)) {
        navigate(event.route)
      } else {
        submit.submit(
          { walletUrl: event.walletUrl },
          {
            action: href('/pay'),
            method: 'POST'
          }
        )
      }
    },
    [navigate, submit]
  )
  useEffect(() => {
    return () => setCommandPaletteOpen(false)
  }, [setCommandPaletteOpen])

  useEffect(() => {
    if (search.state == 'idle') {
      setResults(search.data?.results || [])
    }
    setLoading(false)
  }, [search.data?.results, search.state, setLoading, term.length])

  return (
    <Combobox onChange={_onChangeCombobox}>
      <Form
        id='pay-search-form'
        action={href('/pay')}
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
        <Combobox.Options static className='contents w-full'>
          {term.length > 0 && <Label>Pay</Label>}
          {results.length == 0 && term.length > 0 && (
            <CardContent>Your search returned no results.</CardContent>
          )}
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
                    result.identifierType == 'wallet_url') && (
                      <InterledgerIcon />
                    )}
                  {result.identifierType == 'twitter' && <TwitterIcon />}
                  {result.identifierType == 'discord' && <DiscordIcon />}
                  {result.identifierType == 'domain' && (
                    <Icon>captive_portal</Icon>
                  )}
                  <div className='flex flex-col items-start gap-y-2'>
                    <span className='text-medium'>{result.identifier}</span>
                  </div>
                </div>
                <Icon className='ml-auto'>navigate_next</Icon>
              </Combobox.Option>
            )
          })}
          {actions.length > 0 && <Label className='mt-2'>Actions</Label>}
          {actions.map((action: Action) => {
            return (
              <Combobox.Option
                as={CardButton}
                form='pay-search-form'
                key={action.kbd}
                value={action}
                name='walletUrl'
                type='button'
                className={({ active }) =>
                  clsx('items-center gap-x-3', active ? 'bg-nav-hover' : '')
                }
              >
                <div className='flex gap-x-3'>
                  <Icon>{action.icon}</Icon>
                  <span className='text-medium'>{action.title}</span>
                </div>
                <kbd className='ml-auto flex h-6 w-6 items-center justify-center rounded border-2 border-base'>
                  <span className='font-sans text-xs text-weak'>
                    {action.kbd}
                  </span>
                </kbd>
              </Combobox.Option>
            )
          })}
        </Combobox.Options>
      </Card>
    </Combobox>
  )
}
