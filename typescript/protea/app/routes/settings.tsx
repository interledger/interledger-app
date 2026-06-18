import {
  data,
  href,
  Outlet,
  redirect,
  useLoaderData,
  useLocation,
  useOutletContext
} from 'react-router'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardHeader,
  CardLink,
  CardTitle,
  GridColumn,
  Icon,
  Layouts,
  WalletGrid
} from '~/components'
import { getFeatures, getKycStatus } from '~/data/wallet.server'
import { AccountDeletionRequestStatus } from '~/generated/connect/backend/v1/backend_pb'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { KycStatus } from '~/lib/types'
import type { Route } from './+types/settings'

export async function loader({ request }: Route.LoaderArgs) {
  const url = new URL(request.url)

  const flowId = url.searchParams.get('flow')
  if (flowId) return redirect(`${href('/recovery/password')}?flow=${flowId}`)

  const [{ kycStatus }, linkedAccountsResponse, features] = await Promise.all([
    getKycStatus(request),
    grpc.getLinkedAccounts(request, {}),
    getFeatures(request)
  ])

  const eurBalanceAccount = isConnectError(linkedAccountsResponse)
    ? undefined
    : linkedAccountsResponse.linkedAccounts.find(
        (la) =>
          la.type === 'balance' &&
          la.receiveCurrencyCode === 'EUR' &&
          la.sendCurrencyCode === 'EUR'
      )

  const deletionStatusResponse = features.deleteAccountEnabled
    ? await grpc.getAccountDeletionStatus(request, {})
    : null

  const accountDeletionStatus =
    deletionStatusResponse && !isConnectError(deletionStatusResponse)
      ? deletionStatusResponse.status
      : undefined

  return data({
    kycStatus,
    eurBalanceAccount,
    deleteAccountEnabled: features.deleteAccountEnabled,
    accountDeletionStatus
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Settings'
    }
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Settings'
  }
])

type SettingsContext = {
  eurBalanceAccountCreatedAt?: string
}

export function useSettingsContext() {
  return useOutletContext<SettingsContext>()
}

export default function Page() {
  const {
    kycStatus,
    eurBalanceAccount,
    deleteAccountEnabled,
    accountDeletionStatus
  } = useLoaderData<typeof loader>()
  const location = useLocation()
  const pathSegments = location.pathname.split('/').filter(Boolean)

  return (
    <WalletGrid>
      <GridColumn
        hideOnMobile={pathSegments[pathSegments.length - 1] !== 'settings'}
        className='col-span-full lg:col-span-6'
      >
        <Card>
          <CardHeader>
            <CardTitle>Profile</CardTitle>
          </CardHeader>
          {kycStatus != KycStatus.Unknown && (
            <CardLink
              end
              preventScrollReset
              prefetch='intent'
              to={href('/settings/profile-personal')}
            >
              <div className='mr-auto flex space-x-3'>
                <Icon>account_circle</Icon>
                <span>Personal information</span>
              </div>
              <Icon>navigate_next</Icon>
            </CardLink>
          )}
          <CardLink
            end
            preventScrollReset
            prefetch='intent'
            to={href('/settings/profile-public')}
          >
            <div className='mr-auto flex space-x-3'>
              <Icon>contact_page</Icon>
              <span>Public information</span>
            </div>
            <Icon>navigate_next</Icon>
          </CardLink>
          <CardLink
            end
            preventScrollReset
            prefetch='intent'
            to={href('/settings/profile-contact')}
          >
            <div className='mr-auto flex space-x-3'>
              <Icon>call</Icon>
              <span>Contact information</span>
            </div>
            <Icon>navigate_next</Icon>
          </CardLink>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Account</CardTitle>
          </CardHeader>
          <CardLink
            end
            preventScrollReset
            prefetch='intent'
            to={href('/settings/grants')}
          >
            <div className='mr-auto flex space-x-3'>
              <Icon>request_quote</Icon>
              <span>Grants</span>
            </div>
            <Icon>navigate_next</Icon>
          </CardLink>
          <CardLink
            end
            preventScrollReset
            prefetch='intent'
            to={href('/settings/keys')}
          >
            <div className='mr-auto flex space-x-3'>
              <Icon>key</Icon>
              <span>Keys</span>
            </div>
            <Icon>navigate_next</Icon>
          </CardLink>
          {eurBalanceAccount && (
            <CardLink
              end
              preventScrollReset
              prefetch='intent'
              to={href('/settings/documents')}
            >
              <div className='mr-auto flex space-x-3'>
                <Icon>description</Icon>
                <span>Documents</span>
              </div>
              <Icon>navigate_next</Icon>
            </CardLink>
          )}
          {deleteAccountEnabled && (
            <AccountDeletionRow status={accountDeletionStatus} />
          )}
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Security</CardTitle>
          </CardHeader>
          <CardLink
            end
            preventScrollReset
            prefetch='intent'
            to={href('/totp/two-factor-authentication')}
          >
            <div className='mr-auto flex space-x-3'>
              <Icon>scan</Icon>
              <span>Two Factor Auth</span>
            </div>
            <Icon>navigate_next</Icon>
          </CardLink>
          <CardLink
            end
            preventScrollReset
            prefetch='intent'
            to={href('/login/challenge')}
          >
            <div className='mr-auto flex space-x-3'>
              <Icon>password</Icon>
              <span>Password</span>
            </div>
            <Icon>navigate_next</Icon>
          </CardLink>
          <CardLink
            end
            preventScrollReset
            prefetch='intent'
            to={href('/logout')}
          >
            <div className='mr-auto flex space-x-3'>
              <Icon>logout</Icon>
              <span>Log out</span>
            </div>
            {/*<Icon>open_in_new</Icon>*/}
          </CardLink>
        </Card>
      </GridColumn>
      <GridColumn className='col-span-full lg:col-span-6 lg:col-start-7'>
        <Outlet
          context={
            {
              eurBalanceAccountCreatedAt: eurBalanceAccount?.createdAt
            } satisfies SettingsContext
          }
        />
      </GridColumn>
    </WalletGrid>
  )
}

// A COMPLETED deletion means the account is gone, so the user can't be logged
// in to see this row at all. Only the in-flight states need a label.
const inFlightAccountDeletionRow: Partial<
  Record<AccountDeletionRequestStatus, { icon: string; label: string }>
> = {
  [AccountDeletionRequestStatus.PENDING]: {
    icon: 'schedule',
    label: 'Account deletion pending'
  },
  [AccountDeletionRequestStatus.IN_PROGRESS]: {
    icon: 'autorenew',
    label: 'Account deletion in progress'
  }
}

function AccountDeletionRow({
  status
}: {
  status: AccountDeletionRequestStatus | undefined
}) {
  if (status === undefined) {
    return (
      <div className='my-1 flex rounded-xl p-3 text-medium'>
        <div className='mr-auto flex space-x-3'>
          <Icon>warning</Icon>
          <span>Could not load account deletion status. Try again later.</span>
        </div>
      </div>
    )
  }
  if (status === AccountDeletionRequestStatus.UNSPECIFIED) {
    return (
      <CardLink
        end
        preventScrollReset
        prefetch='intent'
        to={href('/settings/delete-account')}
      >
        <div className='mr-auto flex space-x-3 text-error'>
          <Icon>delete</Icon>
          <span>Delete account</span>
        </div>
        <Icon>navigate_next</Icon>
      </CardLink>
    )
  }
  const row = inFlightAccountDeletionRow[status]
  if (!row) return null
  return (
    <div className='my-1 flex rounded-xl p-3 text-medium'>
      <div className='mr-auto flex space-x-3'>
        <Icon>{row.icon}</Icon>
        <span>{row.label}</span>
      </div>
    </div>
  )
}
