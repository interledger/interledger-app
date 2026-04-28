import { useState } from 'react'
import { href, Navigate } from 'react-router'
import type { ApplicationProps, SelectOptions } from '~/components'
import {
  Alert,
  AlertBody,
  ButtonRouter,
  Card,
  CardButton,
  CardContent,
  CardHeader,
  CardLink,
  CardTitle,
  Icon,
  Layouts,
  Select
} from '~/components'
import { mergeMeta } from '~/lib/meta'
import { useSettingsContext } from './settings'

type StatementDate = SelectOptions & { months: SelectOptions[] }

const MONTHS: SelectOptions[] = [
  { id: '1', name: 'January' },
  { id: '2', name: 'February' },
  { id: '3', name: 'March' },
  { id: '4', name: 'April' },
  { id: '5', name: 'May' },
  { id: '6', name: 'June' },
  { id: '7', name: 'July' },
  { id: '8', name: 'August' },
  { id: '9', name: 'September' },
  { id: '10', name: 'October' },
  { id: '11', name: 'November' },
  { id: '12', name: 'December' }
]

function getStatementDates(accountCreatedAt?: string): StatementDate[] {
  if (!accountCreatedAt) return []

  const now = new Date()
  // assumed GateHub only provides statements up to last month — subject to change
  const lastAvailableDate = new Date(
    Date.UTC(now.getUTCFullYear(), now.getUTCMonth() - 1)
  )
  const lastAvailableMonth = lastAvailableDate.getUTCMonth() + 1
  const lastAvailableYear = lastAvailableDate.getUTCFullYear()

  const firstAvailableYear = new Date(accountCreatedAt).getUTCFullYear()
  const firstAvailableMonth = new Date(accountCreatedAt).getUTCMonth() + 1

  return Array.from(
    { length: lastAvailableYear - firstAvailableYear + 1 },
    (_, i) => {
      const year = lastAvailableYear - i
      const monthStart = year === firstAvailableYear ? firstAvailableMonth : 1
      const monthEnd = year === lastAvailableYear ? lastAvailableMonth : 12
      return {
        id: `${year}`,
        name: `${year}`,
        months: MONTHS.slice(monthStart - 1, monthEnd)
      }
    }
  ).filter((d) => d.months.length > 0)
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: href('/settings'),
      title: 'Documents'
    },
    isNested: true
  }
}

export const meta = mergeMeta(() => [{ title: 'Documents' }])

export default function Page() {
  const { eurBalanceAccountCreatedAt } = useSettingsContext()
  const [expanded, setExpanded] = useState(false)

  const statementDates = getStatementDates(eurBalanceAccountCreatedAt)

  const [year, setYear] = useState(statementDates.at(0))
  const [month, setMonth] = useState(statementDates.at(0)?.months.at(-1))

  if (!eurBalanceAccountCreatedAt)
    return <Navigate to={href('/settings')} replace />

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Confirmations</CardTitle>
        </CardHeader>
        <CardLink
          end
          preventScrollReset
          reloadDocument
          target='_blank'
          to={href('/api/statements/accountConfirmation')}
        >
          <div className='mr-auto flex space-x-3'>
            <Icon>description</Icon>
            <span>Account confirmation</span>
          </div>
          <Icon>open_in_new</Icon>
        </CardLink>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Statements</CardTitle>
        </CardHeader>
        <CardButton onClick={() => setExpanded((v) => !v)}>
          <div className='mr-auto flex space-x-3'>
            <Icon>description</Icon>
            <span>Account statement</span>
          </div>
          <Icon>{expanded ? 'expand_less' : 'expand_more'}</Icon>
        </CardButton>
        {expanded && (
          <CardContent className='flex flex-col gap-4'>
            {statementDates.length === 0 ? (
              <Alert>
                <Icon>error</Icon>
                <AlertBody>No statements available yet.</AlertBody>
              </Alert>
            ) : (
              <>
                <div className='flex gap-3'>
                  <Select
                    className='w-1/2 min-w-0'
                    id='year'
                    label='Year'
                    options={statementDates}
                    value={year}
                    onChange={(v) => {
                      const selected = v as StatementDate
                      setYear(selected)
                      setMonth(selected.months.at(-1))
                    }}
                  />
                  <Select
                    className='w-1/2 min-w-0'
                    id='month'
                    label='Month'
                    options={year?.months ?? []}
                    value={month}
                    onChange={setMonth}
                  />
                </div>
                {year && month && (
                  <ButtonRouter
                    reloadDocument
                    target='_blank'
                    to={href('/api/statements/monthly/:year/:month', {
                      year: year.id,
                      month: month.id
                    })}
                  >
                    Generate
                  </ButtonRouter>
                )}
              </>
            )}
          </CardContent>
        )}
      </Card>
    </>
  )
}
