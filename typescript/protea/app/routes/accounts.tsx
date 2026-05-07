import type { Route } from './+types/accounts'
import { data, redirect } from 'react-router';
import { Outlet, useLoaderData, useLocation } from 'react-router';
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardContent,
  CardHeader,
  CardIcon,
  CardLink,
  CardTitle,
  Chip,
  ChipColor,
  Fab,
  GridColumn,
  Icon,
  Layouts,
  Router,
  WalletGrid
} from '~/components'
import { getLinkedAccounts } from '~/data/accounts.server'
import { getFeatures, getKycStatus } from '~/data/wallet.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { KycStatus } from '~/lib/types'
import styles from '~/styles/flags.css?url'

export async function loader({ request }: Route.LoaderArgs) {
  const { bankAccounts, cardAccounts, interacAccounts } =
    await getLinkedAccounts(request)
  const kycStatus = await getKycStatus(request)

  const balancesResponse = await grpc.getBalances(request, {})
  if (isConnectError(balancesResponse)) throw balancesResponse.error

  const features = await getFeatures(request)

  if (!features.accountsTabEnabled) {
    throw redirect('/')
  }

  return data({
    kycStatus: kycStatus.kycStatus,
    features,
    bankAccounts,
    cardAccounts,
    interacAccounts,
    hasCard: cardAccounts.length > 0,
    hasBank: bankAccounts.length > 0,
    hasInterac: interacAccounts.length > 0,
    hasZABalance:
      balancesResponse.balances.filter((bal) => bal.countryCode == 'ZA')
        .length > 0
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Accounts'
    },
    fab: Fab.Pay
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Accounts'
  }
])

export function links() {
  return [{ rel: 'stylesheet', href: styles }]
}

export default function Page() {
  const {
    features,
    bankAccounts,
    cardAccounts,
    interacAccounts,
    hasInterac,
    hasCard,
    hasBank,
    kycStatus,
    hasZABalance
  } = useLoaderData<typeof loader>()

  const location = useLocation()
  const pathSegments = location.pathname.split('/').filter(Boolean)

  return (
    <WalletGrid>
      <GridColumn
        hideOnMobile={pathSegments[pathSegments.length - 1] !== 'accounts'}
        className='col-span-full lg:col-span-6'
      >
        {kycStatus == KycStatus.Unknown && (
          <Card>
            <CardHeader>
              <CardTitle>Wallet</CardTitle>
              <Chip color={ChipColor.orange}>Reserved</Chip>
            </CardHeader>
            <CardContent>
              <div className='flex items-start space-x-4'>
                <CardIcon>
                  <Icon>account_balance_wallet</Icon>
                </CardIcon>
                <div className='flex flex-col space-y-4'>
                  <p className='text-sm text-medium'>
                    Your wallet is reserved, we just need a few more details to
                    activate it.
                  </p>
                  <Router
                    prefetch='render'
                    className='text-sm font-medium text-primary'
                    to={href('/personal-details')}
                  >
                    Activate wallet
                  </Router>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
        {cardAccounts && hasCard && (
          <Card>
            <CardHeader>
              <CardTitle>Connected cards</CardTitle>
            </CardHeader>
            {cardAccounts.map((method) => (
              <CardLink
                key={method.id}
                to={href('/accounts/:accountId', {
                  accountId: method.id
                })}
                className='items-center justify-between'
              >
                <div className='flex space-x-3'>
                  <Icon>credit_card</Icon>
                  <span>{method.name}</span>
                </div>
                <div className='flex items-center space-x-2'>
                  <Icon>navigate_next</Icon>
                </div>
              </CardLink>
            ))}
            <CardContent>
              <Router
                className='mt-4 text-sm font-medium text-primary'
                to={href('/connect/card')}
              >
                Connect another card
              </Router>
            </CardContent>
          </Card>
        )}
        {kycStatus == KycStatus.Approved &&
          features.cardsEnabled &&
          !hasCard && (
            <Card>
              <CardContent>
                <div className='flex items-start space-x-4'>
                  <CardIcon>
                    <Icon>credit_card</Icon>
                  </CardIcon>
                  <div className='flex flex-col space-y-4'>
                    <p className='text-sm text-medium'>
                      Connect cards to easily send and receive payments.
                    </p>
                    <Router
                      className='text-sm font-medium text-primary'
                      to={href('/connect/card')}
                    >
                      Connect a card
                    </Router>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        {bankAccounts && hasBank && (
          <Card>
            <CardHeader>
              <CardTitle>Connected bank accounts</CardTitle>
            </CardHeader>
            {bankAccounts.map((method) => (
              <CardLink
                key={method.id}
                to={href('/accounts/:accountId', {
                  accountId: method.id
                })}
                className='items-center justify-between'
              >
                <div className='flex space-x-3'>
                  <Icon>account_balance</Icon>
                  <span>{method.name}</span>
                </div>
                <div className='flex items-center space-x-2'>
                  <Icon>navigate_next</Icon>
                </div>
              </CardLink>
            ))}
            {/*<CardContent>*/}
            {/*  <Router*/}
            {/*    className='mt-4 text-sm font-medium text-primary'*/}
            {/*    to={href('/connect/bank')}*/}
            {/*  >*/}
            {/*    Connect another bank account*/}
            {/*  </Router>*/}
            {/*</CardContent>*/}
          </Card>
        )}
        {interacAccounts && hasInterac && (
          <Card>
            <CardHeader>
              <CardTitle>Connected interac accounts</CardTitle>
            </CardHeader>
            {interacAccounts.map((method) => (
              <CardLink
                key={method.id}
                to={href('/accounts/:accountId', {
                  accountId: method.id
                })}
                className='items-center justify-between'
              >
                <div className='flex space-x-3'>
                  <Icon>account_balance</Icon>
                  <span>{method.name}</span>
                </div>
                <div className='flex items-center space-x-2'>
                  <Icon>navigate_next</Icon>
                </div>
              </CardLink>
            ))}
          </Card>
        )}
        {kycStatus == KycStatus.Approved && !hasBank && hasZABalance && (
          <Card>
            <CardContent>
              <div className='flex items-start space-x-4'>
                <CardIcon>
                  <Icon>account_balance</Icon>
                </CardIcon>
                <div className='flex flex-col space-y-4'>
                  <p className='text-sm text-medium'>
                    Connect bank accounts to easily send and receive payments.
                  </p>
                  <Router
                    className='text-sm font-medium text-primary'
                    to={href('/connect/bank/za')}
                  >
                    Connect a bank account
                  </Router>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      </GridColumn>
      <GridColumn className='col-span-full lg:col-span-6 lg:col-start-7'>
        <Outlet />
      </GridColumn>
    </WalletGrid>
  )
}
