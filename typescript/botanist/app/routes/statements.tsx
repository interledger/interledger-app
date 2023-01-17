import { ActionArgs, json } from '@remix-run/node'
import { Form, useActionData } from '@remix-run/react'
import { Button, Shape, TextField } from '~/components'
import { grpcClient, GrpcError, httpMapping, isGrpcError, StatusError } from '~/lib/proto.server'

export default function Page() {
	const actionData = useActionData<typeof action>()

	return (
		<>
			<div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
				<div className='flex justify-between'>
					<h1 className='font-display text-2xl font-medium'>
						Email statements
					</h1>
					<div className='hidden sm:flex'>
						<Shape
							width={'w-8'}
							radius={'rounded-br-full'}
							color={'bg-rose-300'}
						/>
						<Shape
							width={'w-8'}
							radius={'rounded-full'}
							color={'bg-lime-500'}
						/>
					</div>
				</div>

				<Form
					id='email-statements'
					action='/statements'
					method='post'
					className='hidden'
				/>

				<TextField
					id='walletID'
					form='email-statements'
					label='Wallet ID'
					name='walletID'
					defaultValue=""
					type='text'
					className='mt-6'
					aria-invalid={Boolean(actionData?.errors.walletID) || undefined}
					aria-describedby={
						actionData?.errors ? 'wallet-id-error' : undefined
					}
					required
					errorMessage={actionData?.errors.walletID}
				/>

				<TextField
					id='period'
					form='email-statements'
					label='Period'
					name='period'
					defaultValue=""
					type='text'
					className='mt-1'
					aria-invalid={Boolean(actionData?.errors.period) || undefined}
					aria-describedby={
						actionData?.errors.period ? 'period-error' : undefined
					}
					required
					errorMessage={actionData?.errors.period}
				/>

				<Button className='mt-12' form='email-statements' type='submit'>
					Submit
				</Button>
			</div>
		</>
	)
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'WalletID' | 'Period'

function mapper(
	field: fieldErrorsMap
): 'walletID' | 'period' | null {
	switch (field) {
		case 'WalletID':
			return 'walletID'
		case 'Period':
			return 'period'
		default:
			return null
	}
}

export async function action({ request }: ActionArgs) {
	const form = await request.formData()
	const period = form.get("period") as string
	const walletID = form.get("walletID") as string

	const fieldErrors = {
		form: '',
		period: '',
		walletID: ''
	}
	const response = await grpcClient
		.emailWalletStatement(
			{
				period: period?.toString(),
				walletID: walletID?.toString()
			},
			{
				meta: {
					cookies: String(request.headers.get('cookie')) || ''
				}
			}
		)
		.then((v) => v)
		.catch(StatusError)
	if (isGrpcError(response)) {
		if (response.code == 3) {
			for (let violation of (response as GrpcError).details[0]
				.fieldViolations) {
				const field = mapper(violation.field as fieldErrorsMap)
				if (field != null) fieldErrors[field] = violation.description
			}
			return json({ errors: { ...fieldErrors } }, { status: 400 })
		} else throw json({}, httpMapping(response.code))
	}

	return json({ errors: { ...fieldErrors } }, { status: 200 })
}
