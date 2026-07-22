import { href, useLoaderData } from 'react-router'
import { Card, CardContent, Icon, Router } from '~/components'
import type { AppLoaderData, loader } from '../route'


export const Cards = () => {
    const { features, walletInfo } = useLoaderData<
        typeof loader
    >() as AppLoaderData

    return <>
        {features.cardsEnabled && !walletInfo.hasCard && (
            <Card>
                <CardContent>
                    <div className='flex items-start space-x-4'>
                        <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
                            <Icon>credit_card</Icon>
                        </div>
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
    </>
}