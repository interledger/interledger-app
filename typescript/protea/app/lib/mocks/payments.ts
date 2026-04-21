import { Timestamp } from '@bufbuild/protobuf'
import { Amount, Transaction } from '~/generated/connect/backend/v1/backend_pb'

/**
 * Mock transaction/payment data for testing and development
 */
export const mockTransactions: Transaction[] = [
  // Successful sent payment
  new Transaction({
    id: 'tx-1',
    type: 'sent',
    amount: new Amount({
      amount: BigInt(10000),
      asset: 'USD',
      assetScale: 2,
      country: 'US'
    }),
    source: 'wallet-123',
    destination: 'wallet-456',
    timestamp: Timestamp.fromDate(new Date('2024-10-25T14:30:00Z')),
    state: 'Completed',
    title: 'Payment to John Doe',
    formattedAmount: '$100.00',
    formattedTime: '2:30 PM',
    formattedDate: 'Oct 25, 2024',
    subtotal: '-$100.00',
    destinationIdentity: '@johndoe',
    destinationIdentityType: 'wallet',
    receiverAccountId: 'acc-456',
    senderAccountId: 'acc-123'
  }),

  // Successful received payment
  new Transaction({
    id: 'tx-2',
    type: 'received',
    amount: new Amount({
      amount: BigInt(5000),
      asset: 'USD',
      assetScale: 2,
      country: 'US'
    }),
    source: 'wallet-789',
    destination: 'wallet-123',
    timestamp: Timestamp.fromDate(new Date('2024-10-24T10:15:00Z')),
    state: 'Completed',
    title: 'Payment from Jane Smith',
    formattedAmount: '$50.00',
    formattedTime: '10:15 AM',
    formattedDate: 'Oct 24, 2024',
    subtotal: '+$50.00',
    destinationIdentity: '@janesmith',
    destinationIdentityType: 'wallet',
    receiverAccountId: 'acc-123',
    senderAccountId: 'acc-789'
  }),

  // Pending payment
  new Transaction({
    id: 'tx-3',
    type: 'sent',
    amount: new Amount({
      amount: BigInt(2500),
      asset: 'USD',
      assetScale: 2,
      country: 'US'
    }),
    source: 'wallet-123',
    destination: 'wallet-999',
    timestamp: Timestamp.fromDate(new Date('2024-10-26T16:45:00Z')),
    state: 'Pending',
    title: 'Payment to Bob Wilson',
    formattedAmount: '$25.00',
    formattedTime: '4:45 PM',
    formattedDate: 'Oct 26, 2024',
    subtotal: '-$25.00',
    destinationIdentity: '@bobwilson',
    destinationIdentityType: 'wallet',
    receiverAccountId: 'acc-999',
    senderAccountId: 'acc-123'
  }),

  // Failed payment
  new Transaction({
    id: 'tx-4',
    type: 'sent',
    amount: new Amount({
      amount: BigInt(15000),
      asset: 'USD',
      assetScale: 2,
      country: 'US'
    }),
    source: 'wallet-123',
    destination: 'wallet-111',
    timestamp: Timestamp.fromDate(new Date('2024-10-23T09:20:00Z')),
    state: 'Failed',
    title: 'Payment to Alice Brown',
    formattedAmount: '$150.00',
    formattedTime: '9:20 AM',
    formattedDate: 'Oct 23, 2024',
    subtotal: '-$150.00',
    destinationIdentity: '@alicebrown',
    destinationIdentityType: 'wallet',
    receiverAccountId: 'acc-111',
    senderAccountId: 'acc-123'
  }),

  // Web monetization outgoing
  new Transaction({
    id: 'tx-5',
    type: 'web_monetization_outgoing',
    amount: new Amount({
      amount: BigInt(50),
      asset: 'USD',
      assetScale: 2,
      country: 'US'
    }),
    source: 'wallet-123',
    destination: 'webmon-1',
    timestamp: Timestamp.fromDate(new Date('2024-10-25T18:00:00Z')),
    state: 'Completed',
    title: 'Web Monetization - example.com',
    formattedAmount: '$0.50',
    formattedTime: '6:00 PM',
    formattedDate: 'Oct 25, 2024',
    subtotal: '-$0.50',
    destinationIdentity: 'example.com',
    destinationIdentityType: 'wallet',
    receiverAccountId: 'acc-webmon-1',
    senderAccountId: 'acc-123'
  }),

  // Web monetization incoming
  new Transaction({
    id: 'tx-6',
    type: 'web_monetization_incoming',
    amount: new Amount({
      amount: BigInt(75),
      asset: 'USD',
      assetScale: 2,
      country: 'US'
    }),
    source: 'webmon-2',
    destination: 'wallet-123',
    timestamp: Timestamp.fromDate(new Date('2024-10-24T14:30:00Z')),
    state: 'Completed',
    title: 'Web Monetization - myblog.com',
    formattedAmount: '$0.75',
    formattedTime: '2:30 PM',
    formattedDate: 'Oct 24, 2024',
    subtotal: '+$0.75',
    destinationIdentity: 'myblog.com',
    destinationIdentityType: 'wallet',
    receiverAccountId: 'acc-123',
    senderAccountId: 'acc-webmon-2'
  }),

  // Withdrawal
  new Transaction({
    id: 'tx-7',
    type: 'withdrawal',
    amount: new Amount({
      amount: BigInt(20000),
      asset: 'USD',
      assetScale: 2,
      country: 'US'
    }),
    source: 'wallet-123',
    destination: 'bank-account-1',
    timestamp: Timestamp.fromDate(new Date('2024-10-22T11:00:00Z')),
    state: 'Completed',
    title: 'Withdrawal to Bank Account',
    formattedAmount: '$200.00',
    formattedTime: '11:00 AM',
    formattedDate: 'Oct 22, 2024',
    subtotal: '-$200.00',
    fees: '$2.50',
    destinationIdentity: 'Bank **** 1234',
    destinationIdentityType: 'wallet',
    receiverAccountId: 'bank-acc-1',
    senderAccountId: 'acc-123'
  }),

  // Deposit
  new Transaction({
    id: 'tx-8',
    type: 'deposit',
    amount: new Amount({
      amount: BigInt(30000),
      asset: 'USD',
      assetScale: 2,
      country: 'US'
    }),
    source: 'bank-account-2',
    destination: 'wallet-123',
    timestamp: Timestamp.fromDate(new Date('2024-10-21T15:45:00Z')),
    state: 'Completed',
    title: 'Deposit from Bank Account',
    formattedAmount: '$300.00',
    formattedTime: '3:45 PM',
    formattedDate: 'Oct 21, 2024',
    subtotal: '+$300.00',
    destinationIdentity: 'Bank **** 5678',
    destinationIdentityType: 'wallet',
    receiverAccountId: 'acc-123',
    senderAccountId: 'bank-acc-2'
  }),

  // Twitter payment
  new Transaction({
    id: 'tx-9',
    type: 'sent',
    amount: new Amount({
      amount: BigInt(1000),
      asset: 'USD',
      assetScale: 2,
      country: 'US'
    }),
    source: 'wallet-123',
    destination: 'twitter-user-1',
    timestamp: Timestamp.fromDate(new Date('2024-10-20T13:30:00Z')),
    state: 'Completed',
    title: 'Payment to @twitteruser',
    formattedAmount: '$10.00',
    formattedTime: '1:30 PM',
    formattedDate: 'Oct 20, 2024',
    subtotal: '-$10.00',
    destinationIdentity: '@twitteruser',
    destinationIdentityType: 'Twitter',
    receiverAccountId: 'twitter-acc-1',
    senderAccountId: 'acc-123'
  }),

  // Slack payment
  new Transaction({
    id: 'tx-11',
    type: 'received',
    amount: new Amount({
      amount: BigInt(750),
      asset: 'USD',
      assetScale: 2,
      country: 'US'
    }),
    source: 'slack-user-1',
    destination: 'wallet-123',
    timestamp: Timestamp.fromDate(new Date('2024-10-18T16:20:00Z')),
    state: 'Completed',
    title: 'Payment from SlackUser',
    formattedAmount: '$7.50',
    formattedTime: '4:20 PM',
    formattedDate: 'Oct 18, 2024',
    subtotal: '+$7.50',
    destinationIdentity: '@slackuser',
    destinationIdentityType: 'Slack',
    receiverAccountId: 'acc-123',
    senderAccountId: 'slack-acc-1'
  }),

  // Pending web monetization
  new Transaction({
    id: 'tx-12',
    type: 'web_monetization_outgoing',
    amount: new Amount({
      amount: BigInt(25),
      asset: 'USD',
      assetScale: 2,
      country: 'US'
    }),
    source: 'wallet-123',
    destination: 'webmon-3',
    timestamp: Timestamp.fromDate(new Date('2024-10-27T12:00:00Z')),
    state: 'Pending',
    title: 'Web Monetization - news.site',
    formattedAmount: '$0.25',
    formattedTime: '12:00 PM',
    formattedDate: 'Oct 27, 2024',
    subtotal: '-$0.25',
    destinationIdentity: 'news.site',
    destinationIdentityType: 'wallet',
    receiverAccountId: 'acc-webmon-3',
    senderAccountId: 'acc-123'
  })
]

export const mockDateGroupedTransactions = Object.values(
  [...mockTransactions].reduce<{
    [date: string]: Transaction[]
  }>((prev, current) => {
    prev[current.formattedDate] = prev[current.formattedDate] || []
    prev[current.formattedDate].push(current)
    return prev
  }, Object.create(null))
)
