import { useFetcher, useLoaderData, data as rrData, LoaderFunctionArgs, ActionFunctionArgs, useRouteError, isRouteErrorResponse } from 'react-router';
import {
    Button,
    Card,
    CardHeader,
    CardTitle,
    Layouts,
    Select,
    TextField
} from '~/components'
import { useState } from 'react';
import type { ApplicationProps } from '~/components'
import { createDummyClient } from '~/lib/error-handling/client';
import { FailedServerResponse, SuccessfulServerResponse } from '~/lib/error-handling/types';

export async function loader({ request, params }: LoaderFunctionArgs) {
    const client = createDummyClient(request)
    const clientResp = client.getTransactions()

    if (client.isError(clientResp)) {
        const errrrr = clientResp
        return rrData<FailedServerResponse>({ success: false, message: clientResp.message }, { status: 500 })
    }

    return rrData<SuccessfulServerResponse>({ success: true, data: clientResp })
}

export const handle: ApplicationProps = {
    layout: Layouts.Wallet,
    scaffold: {
        header: {
            title: 'Error Handling Test'
        }
    }
}

export async function action({ request, params }: ActionFunctionArgs) {
    const client = createDummyClient(request)
    const formData = await request.formData();
    const data = Object.fromEntries(formData);
    const subType = data.subType as string;

    let clientResp;
    if (subType === 'option1') {
        clientResp = client.submitSuccesful(data);
    } else if (subType === 'option2') {
        clientResp = client.submit401(data);
    } else if (subType === 'option3') {
        clientResp = client.submit403(data);
    } else if (subType === 'option4') {
        clientResp = client.submit500(data);
    } else {
        clientResp = client.submitSuccesful(data);
    }

    if (client.isError(clientResp)) {
        let zzz = clientResp
        // const userFacingError = ErrorMapper.dummyClient.toUserFacingError(result)
        // return rrData<FailedActionResponse>({ success: false, errors: userFacingError.errors }, { status: result.status })
        return rrData<FailedServerResponse>({ success: false, message: clientResp.message })
    }

    return rrData<SuccessfulServerResponse>({ success: true, data: clientResp });
}

export default function Page() {
    const loaderData = useLoaderData<typeof loader>()
    const fetcher = useFetcher<typeof action>()

    const options = [
        { id: 'option1', name: 'Option 1 (Success)' },
        { id: 'option2', name: 'Option 2 (Error 400)' },
        { id: 'option3', name: 'Option 3 (Error 403)' },
        { id: 'option4', name: 'Option 4 (Error 500)' }
    ]
    const [selectedOption, setSelectedOption] = useState(options[0])

    if (loaderData.success == false) {
        return (
            <div className="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
                <h3 className="font-bold">Error Data from Loader:</h3>
                <pre className="text-xs overflow-auto">{JSON.stringify(loaderData.errors, null, 2)}</pre>
            </div>
        )
    }

    return (
        <div className="space-y-6">
            <fetcher.Form method="post" id="dummy-form">
                <Card>
                    <CardHeader>
                        <CardTitle>Dummy Form</CardTitle>
                    </CardHeader>
                    <TextField
                        id="username"
                        name="username"
                        label="Username"
                        form="dummy-form"
                        required
                        className="mt-6"
                    />
                    <TextField
                        id="password"
                        name="password"
                        label="Password"
                        type="password"
                        form="dummy-form"
                        required
                        className="mt-4"
                    />
                    <Select
                        id="subType-select"
                        label="Submission Type"
                        options={options}
                        value={selectedOption}
                        onChange={setSelectedOption}
                        className="mt-4"
                    />
                    <input type="hidden" name="subType" value={selectedOption.id} />
                </Card>
                <Button type="submit" form="dummy-form" className="mt-4" disabled={fetcher.state !== 'idle'}>
                    {fetcher.state !== 'idle' ? 'Submitting...' : 'Submit to /errhandle'}
                </Button>
            </fetcher.Form>

            {fetcher.data && fetcher.data.success == true && (
                <div className="p-4 bg-green-50 border border-green-200 rounded-lg text-green-700">
                    <h3 className="font-bold">Fetcher Submitted Data:</h3>
                    <pre className="text-xs overflow-auto">{JSON.stringify(fetcher.data.data, null, 2)}</pre>
                </div>
            )}

            {fetcher.data && fetcher.data.success == false && (
                <div className="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
                    <h3 className="font-bold">Error Data from Loader:</h3>
                    <pre className="text-xs overflow-auto">Errors: {JSON.stringify(fetcher.data.errors, null, 2)}</pre>
                    <pre className="text-xs overflow-auto">Message: {JSON.stringify(fetcher.data.message, null, 2)}</pre>
                </div>
            )}
        </div>
    )
}


export function ErrorBoundary() {
    const error = useRouteError()
    console.error("EROARE: ", error)

    if (isRouteErrorResponse(error)) {
        return (
            <>
                <div>EROARE isRouteErrorResponse: </div>
                <div>status: ${error.status}</div>
                <div> statusText: ${error.statusText}</div>
                <div>data: ${error.data}</div>
            </>
        )
    }

    return (
        <>
            <div>EROARE: </div>
            <div>message: ${(error as any).message}</div>
        </>
    )
}