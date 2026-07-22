import { href, useLoaderData } from 'react-router'
import { Card, CardContent, Icon, Router } from '~/components'
import type { AppLoaderData, loader } from '../route'


// The US Plaid flow depends on a USD balance that is provisioned asynchronously
// after KYC approval. Only offer the "Connect a bank" CTA once that balance
// actually exists — otherwise the user is silently bounced from /connect/bank.
// The legacy US manual form and the ZA form don't gate on this.
export function canConnectBank(
    country: string,
    plaidEnabled: boolean,
    balances: { countryCode: string }[]
): boolean {
    const isUSPlaidPath = country === 'US' && plaidEnabled
    const hasUSLinkedAccount = balances.some((bal) => bal.countryCode === 'US');
    return isUSPlaidPath || hasUSLinkedAccount
    // return !isUSPlaidPath || hasUSLinkedAccount
}

export const ConnectBank = () => {
    const { features, walletInfo, plaidEnabled, balances } = useLoaderData<
        typeof loader
    >() as AppLoaderData

    const bankHref =
        walletInfo.country === 'US'
            ? plaidEnabled
                ? href('/connect/bank')
                : href('/connect/bank/us')
            : href('/connect/bank/za')

    const showBankCTA = canConnectBank(
        walletInfo.country,
        plaidEnabled,
        balances
    )

    return <>
        {features.banksEnabled && showBankCTA && (
            <Card>
                <CardContent>
                    <div className='flex items-start space-x-4'>
                        <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
                            <Icon>account_balance</Icon>
                        </div>
                        <div className='flex flex-col space-y-4'>
                            <p className='text-sm text-medium'>
                                Connect bank accounts to easily add or withdraw from your
                                balance.
                            </p>
                            <Router
                                className='text-sm font-medium text-primary'
                                to={bankHref}
                            >
                                Connect a bank
                            </Router>
                        </div>
                    </div>
                </CardContent>
            </Card>
        )}
    </>
}