import clsx from 'clsx'
import { useState } from 'react'
import { href, Navigate } from 'react-router'
import type { ApplicationProps, SelectOptions } from '~/components'
import {
  Alert,
  AlertBody,
  ButtonRouter,
  Card,
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

function getStatementDates(accountCreatedAt?: string) {
  const now = new Date()
  const lastMonth = new Date(
    Date.UTC(now.getUTCFullYear(), now.getUTCMonth() - 1)
  )
  const lastAvailableMonth = lastMonth.getUTCMonth() + 1
  const lastAvailableYear = lastMonth.getUTCFullYear()

  const firstAvailableYear = accountCreatedAt
    ? new Date(accountCreatedAt).getUTCFullYear()
    : lastAvailableYear
  const firstAvailableMonth = accountCreatedAt
    ? new Date(accountCreatedAt).getUTCMonth() + 1
    : 1

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
  )
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
  const { isDocumentsEnabled, accountCreatedAt } = useSettingsContext()
  const [expanded, setExpanded] = useState(false)
  const [year, setYear] = useState<StatementDate>()
  const [month, setMonth] = useState<SelectOptions>()

  const statementDates = getStatementDates(accountCreatedAt)

  if (!isDocumentsEnabled) return <Navigate to={href('/settings')} replace />

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
        <CardLink
          end
          preventScrollReset
          onClick={() => setExpanded((v) => !v)}
          to='#'
        >
          <div className='mr-auto flex space-x-3'>
            <Icon>description</Icon>
            <span>Account statement</span>
          </div>
          <Icon>{expanded ? 'expand_less' : 'expand_more'}</Icon>
        </CardLink>
        {expanded && (
          <CardContent className='flex flex-col gap-4'>
            {statementDates.every((p) => p.months.length === 0) ? (
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
                    options={statementDates.filter((p) => p.months.length > 0)}
                    value={year}
                    onChange={(v) => {
                      setYear(v as StatementDate)
                      setMonth(undefined)
                    }}
                  />
                  <div
                    className={clsx(
                      'w-1/2 min-w-0',
                      !year && 'cursor-not-allowed opacity-50'
                    )}
                  >
                    <Select
                      key={year?.id}
                      className='min-w-full'
                      id='month'
                      label='Month'
                      options={year?.months ?? []}
                      value={month}
                      disabled={!year}
                      onChange={setMonth}
                    />
                  </div>
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
