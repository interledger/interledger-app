import { ActionArgs } from "@remix-run/node";
import { url } from "inspector";

export async function action({ request }: ActionArgs) {
	try {
		let expectedDsn = process.env.SENTRY_DSN || ""
		let envelope = await request.text()
		let header = envelope.split("\n")[0]
		let headerObject = JSON.parse(header)
		if (typeof headerObject.dsn == "undefined" || headerObject.dsn == "") {
			return new Response(null, { status: 404 })
		}

		// only allow requests for our dsn
		if (headerObject.dsn !== expectedDsn) {
			return new Response(null, { status: 404 })
		}

		let url = new URL(expectedDsn)
		let projectID = url.pathname.replace("/", ``)
		let proxiedRequest = await fetch(`https://${url.hostname}/api/${projectID}/envelope`, {
			method: 'POST',
			body: envelope,
			headers: { 'Content-Type': 'application/x-sentry-envelope' }
		})

		return new Response(null, { status: 200 })
	} catch (error) {
		console.error('sentry error: ', error)
		return new Response(null, { status: 404 })
	}
}
