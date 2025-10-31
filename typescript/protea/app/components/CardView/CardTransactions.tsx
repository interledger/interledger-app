import type { SerializeFrom } from '@remix-run/node'
import { route } from 'routes-gen'
import type { Transaction } from '~/generated/connect/backend/v1/backend_pb'
import { Card, CardContent, CardHeader, CardLink, CardTitle } from '../Card'
import { Chip, ChipColor } from '../Chip'
import { Icon } from '../Icon'

export type SerializedTransaction = SerializeFrom<Transaction>

interface CardTransactionsProps {
  transactions: SerializedTransaction[]
}

const getChipColor = (status: string): ChipColor => {
  switch (status) {
    case 'Completed':
      return ChipColor.green
    case 'Pending':
      return ChipColor.orange
    case 'Failed':
      return ChipColor.red
    default:
      return ChipColor.blue
  }
}

export const CardTransactions = ({ transactions }: CardTransactionsProps) => {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Latest transactions</CardTitle>
      </CardHeader>
      <CardContent>
        {transactions &&
          transactions.map((tx) => (
            <CardLink
              key={tx.id}
              to={route('/payments/:paymentId', { paymentId: tx.id })}
              className='justify-between space-x-4'
            >
              <div className='flex w-7/12 items-center space-x-4'>
                <div className='flex w-full flex-col space-y-1'>
                  <span className='truncate text-medium'>{tx.title}</span>
                  <span className='text-xs text-weak'>{tx.formattedDate}</span>
                </div>
              </div>
              <div className='flex min-w-max flex-initial items-center space-x-2'>
                <span className='text-medium'>{tx.subtotal}</span>
                <Chip color={getChipColor(tx.state)}>{tx.state}</Chip>
                <Icon>navigate_next</Icon>
              </div>
            </CardLink>
          ))}
      </CardContent>
    </Card>
  )
}
