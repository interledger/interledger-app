import {
    data,
    type LoaderFunctionArgs,
} from 'react-router';
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { jsonWithCSRF } from '~/lib/csrf.server'
import { envBool } from '~/env.server'
import {
    getBalancesForTransfer,
    getLinkedAccountsForDeposit
} from '~/data/accounts.server'
import type {
    Balance,
    XagoDepositDetails
} from '~/generated/connect/backend/v1/backend_pb'
import type { FormattedLinkedAccount } from '~/data/accounts.server'

export async function gatehubDepositLoader({ request }: LoaderFunctionArgs) {
    const widgetResponse = await grpc.getGatehubDepositWidget(request, {})
    if (isConnectError(widgetResponse)) throw widgetResponse.error

    return jsonWithCSRF(request, {
        provider: 'gatehub',
        gatehubWidgetUrl: widgetResponse.widgetUrl
    })
}

export async function chimoneyDepositLoader({ request }: LoaderFunctionArgs) {
    return jsonWithCSRF(request, { provider: 'chimoney' })
}

export async function fynbosDepositLoader({ request }: LoaderFunctionArgs) {
    const providerResponse = await grpc.getOnOffRampProvider(request, {})
    if (isConnectError(providerResponse)) throw providerResponse.error
    const url = new URL(request.url)
    let balanceAccount: Balance | undefined
    let balance: FormattedLinkedAccount | undefined
    let depositDetails: XagoDepositDetails | undefined

    const balanceResponse = await grpc.getBalances(request, {})
    if (isConnectError(balanceResponse)) throw balanceResponse.error

    const balances = await getBalancesForTransfer(request)

    const linkedAccount = url.searchParams.get('linkedAccount')

    if (linkedAccount) {
        balanceAccount = balanceResponse.balances.find(
            (acc) => acc.linkedAccount == linkedAccount
        )
    } else {
        balanceAccount = balanceResponse.balances[0]
    }
    balance = balances.find((acc) => acc.id == balanceAccount?.linkedAccount)
    if (
        !balanceAccount ||
        typeof balanceAccount == 'undefined' ||
        !balance ||
        typeof balance == 'undefined'
    )
        throw data({}, { status: 404 })

    const linkedAccounts = await getLinkedAccountsForDeposit(
        request,
        balanceAccount.linkedAccount
    )

    if (linkedAccounts.length == 0 && balanceAccount.countryCode == 'ZA') {
        let details = await grpc.getXagoDepositDetails(request, {
            linkedAccount: balanceAccount.linkedAccount
        })
        if (isConnectError(details)) throw details.errorResponse

        let ret = details.details.filter(
            (d) => d.currency == balanceAccount?.currency
        )
        depositDetails = ret[0]
    }

    return jsonWithCSRF(request, {
        provider: providerResponse.provider,
        balanceAccount,
        balance,
        balances,
        linkedAccounts,
        depositDetails,
        showTestDeposit: envBool("FEATURE_FLAG_TESTING_DEPOSIT")
    })
}