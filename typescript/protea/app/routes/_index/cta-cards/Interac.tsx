import { href, useLoaderData } from 'react-router'
import { Card, CardContent, Icon, Router } from '~/components'
import type { AppLoaderData, loader } from '../route'


export const Interac = () => {
    const { features } = useLoaderData<
        typeof loader
    >() as AppLoaderData

    return <>
        {features.interacEnabled && (
            <Card>
                <CardContent>
                    <div className='flex items-start space-x-4'>
                        <div className='flex items-center justify-between rounded-full bg-nav p-5 text-medium'>
                            <Icon>account_balance</Icon>
                        </div>
                        <div className='flex flex-col space-y-4'>
                            <p className='text-sm text-medium'>
                                Connect an Interac account to easily withdraw from your
                                balance.
                            </p>
                            <Router
                                className='text-sm font-medium text-primary'
                                to={href('/connect/interac')}
                            >
                                Connect an Interac account
                            </Router>
                        </div>
                    </div>
                </CardContent>
            </Card>
        )}
    </>
}