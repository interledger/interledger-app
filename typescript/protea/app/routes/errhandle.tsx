import { ActionFunction, Form, LoaderFunction, useFetcher, useLoaderData, data as rrData } from 'react-router';
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
// import { ErrorMapper } from '~/lib/error-handling/bff-error';
import { createDummyClient } from '~/lib/error-handling/client';
import { FailedActionResponse, SuccessfulActionResponse } from '~/lib/error-handling/types';

export async function loader({ request, params }: LoaderFunction) {
    const client = createDummyClient(request)
    const data = client.getTransactions(true)

    if (client.isError(data)) {
        return rrData<FailedActionResponse>({ success: false, errors: data.message }, { status: 500 })
    }

    return rrData<SuccessfulActionResponse>({ success: true, data })
}

export const handle: ApplicationProps = {
    layout: Layouts.Focus,
    scaffold: {
        header: {
            title: 'Error Handling Test'
        }
    }
}

export async function action({ request, params }: ActionFunction) {
    const client = createDummyClient(request)
    const formData = await request.formData();
    const data = Object.fromEntries(formData);
    const subType = data.subType as string;

    let result;
    if (subType === 'option1') {
        result = client.submitSuccesful(data);
    } else if (subType === 'option2') {
        result = client.submit401(data);
    } else if (subType === 'option3') {
        result = client.submit403(data);
    } else {
        result = client.submitSuccesful(data);
    }

    if (client.isError(result)) {
        // const userFacingError = ErrorMapper.dummyClient.toUserFacingError(result)
        // return rrData<FailedActionResponse>({ success: false, errors: userFacingError.errors }, { status: result.status })
        return rrData<FailedActionResponse>({ success: false, errors: result.message }, { status: result.status })
    }

    return rrData<SuccessfulActionResponse>({ success: true, data: result });
}

export default function Page() {
    const loaderData = useLoaderData<typeof loader>()
    const fetcher = useFetcher<typeof action>()

    const options = [
        { id: 'option1', name: 'Option 1 (Success)' },
        { id: 'option2', name: 'Option 2 (Error 400)' },
        { id: 'option3', name: 'Option 3 (Error 403)' }
    ]
    const [selectedOption, setSelectedOption] = useState(options[0])

    if (loaderData.success == false) {
        return (
            <div className="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
                <h3 className="font-bold">xxxx Error Data from Loader:</h3>
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
                    <pre className="text-xs overflow-auto">{JSON.stringify(fetcher.data.errors, null, 2)}</pre>
                </div>
            )}
        </div>
    )
}
