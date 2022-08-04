import { ActionArgs, json, LoaderArgs } from "@remix-run/node";
import { useActionData, useLoaderData, useSubmit } from "@remix-run/react";
import { useEffect } from "react";
import { requireUserSession } from "~/lib/kratos.server";
import { grpcClient, isGrpcError, StatusError } from "~/lib/proto.server";
import { useScript } from "~/lib/useScript";

export async function loader({ request }: LoaderArgs) {
	await requireUserSession(request)
	const cookie = request.headers.get('cookie')
	let response = await grpcClient.getBankAccountWidget(
		{},
		{
			meta: {
				"cookies": cookie ?? "",
			},
		},
	).then((v) => v).catch(StatusError)
	if (isGrpcError(response)) {
		throw response
	}

	return json({ widgetUrl: response.response.url })
}

export async function action({ request }: ActionArgs) {
	let data = await request.formData()
	const userGuid = data.get("userGuid")?.toString()
	const memberGuid = data.get("memberGuid")?.toString()
	console.log("member=", memberGuid, "user=", userGuid)

	const cookie = request.headers.get('cookie')
	let response = await grpcClient.initiateCreateBankAccount(
		{
			memberGuid: data.get("memberGuid")?.toString() ?? "",
			userGuid: data.get("userGuid")?.toString() ?? "",
			name: "test",
		},
		{
			meta: {
				"cookies": cookie ?? "",
			},
		},
	).then((v) => v).catch(StatusError)
	if (isGrpcError(response)) {
		throw response
	}

	return json({ reference: response.response.reference });
}

interface MemberConnectedEvent {
	type: string
	mx: boolean
	metadata: { user_guid: string, member_guid: string }
}

function isMemberConnectedEvent(event: any): event is MemberConnectedEvent {
	return event.type === "mx/connect/memberConnected"
}

export default function Page() {
	const submit = useSubmit()

	const actionData = useActionData<typeof action>()
	const { widgetUrl } = useLoaderData<typeof loader>()
	const state = useScript("https://atrium.mx.com/connect.js")

	useEffect(() => {
		window.addEventListener('message', function (event) {
			try {
				/**
				 * Versions null -> 3. These versions are deprecated. They JSON encode
				 * the `data` attribute so you must wrap in this in a try catch to try
				 * to get the data.
				 */
				const data = JSON.parse(event.data);

				if (data.moneyDesktop) {
					console.log('versions null -> 3', data);
				}
			} catch (error) {
				/**
				 * Versions 4 and on no longer JSON encode the data attribute. It should
				 * be immediately accessible. If there is a data.mx attribute that is
				 * `true` then this is a version 4 message at this point.
				 */
				if (event.data.mx) {
					console.log('version 4 message', event.data);
					if (isMemberConnectedEvent(event.data)) {
						let formData = new FormData()
						formData.append("memberGuid", event.data.metadata.member_guid)
						formData.append("userGuid", event.data.metadata.user_guid)

						submit(formData, { method: "post", action: "/mx" })
						let connectIFrame = document.querySelector('iframe[title="Connect"]')
						connectIFrame?.remove()
					}
				}
			}
		})
	}, [])

	useEffect(() => {
		if (state === "ready") {
			console.log("ready url=", widgetUrl)
			let mxConnect = new (window as any).MXConnect({
				id: "widget",
				iframeTitle: 'Connect',
				/**
				 * Callback that for handling all events withhin connect.
				 * Only called in  ui_message_version 4 or higher
				 *
				 * The events called here are the same events that come through post
				 * messages.
				 *
				 * Details about the events:
				 * https://atrium.mx.com/docs#connect-events
				 */
				onEvent: function (type: string, payload: unknown) {
					console.log("onEvent", type, payload);
				},
				/**
				 * *Deprecated*.Called when the connect widget is mounted in the DOM.
				 * Only called if the`ui_message_version` is below `4` or not set all.
				 */
				onLoad: function (event: unknown) {
					console.log("onLoad", event);
				},
				/**
				 * *Deprecated*. Called when a member is successfully 'CONNECTED'.
				 * Only called if the `ui_message_version` is below `4` or not set all.
				 */
				onSuccess: function (event: unknown) {
					console.log("onSuccess", event);
				},
				/**
				 * Any configuration options documented can be configured here:
				 * https://atrium.mx.com/docs#configuration
				 *
				 * These values will override any configuration values that were set when
				 * the URL was requested.
				 */
				// config: {},
				targetOrigin: "*",
			})

			mxConnect.load(widgetUrl)
		}
	}, [state])

	return <>
		<div id="widget" className="mx-auto w-full"></div>
		<div>reference={actionData?.reference}</div>
	</>
}
