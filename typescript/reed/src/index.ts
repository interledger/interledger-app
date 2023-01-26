import { StreamServer } from '@interledger/stream-receiver'
import { deserializeIlpPacket, Errors, errorToReject, isIlpReply, serializeIlpFulfill, serializeIlpReply, Type } from "ilp-packet"
import PluginHttp from "ilp-plugin-http"
import koa from 'koa'
import bodyParser from 'koa-bodyparser'

const ILP_ADDRESS = process.env.ILP_ADDRESS || "t.fynbos"
const ASSET_CODE = process.env.ASSET_CODE || "USD"
const ASSET_SCALE = parseInt(process.env.ASSET_SCALE || "2")
const PEER_URL = process.env.OUTGOING_URL || "https://coil.test"
const SERVER_SECRET = process.env.SERVER_SECRET || ""
const INCOMING_SECRET = process.env.INCOMING_SECRET || ""
const OUTGOING_SECRET = process.env.OUTGOING_SECRET || ""
const ILP_PORT = parseInt(process.env.ILP_PORT || "8080")
const ADMIN_PORT = parseInt(process.env.ILP_PORT || "8081")

const server = new StreamServer({
	serverSecret: Buffer.from(process.env.SERVER_SECRET as string, 'hex'),
	serverAddress: process.env.ILP_ADDRESS as string,
})

async function main() {
	let plugin = new PluginHttp({
		incoming: {
			port: ILP_PORT,
			secretToken: INCOMING_SECRET,
		},
		outgoing: {
			url: PEER_URL,
			secretToken: OUTGOING_SECRET,
		},
		ildcp: {
			assetCode: ASSET_CODE,
			assetScale: ASSET_SCALE,
			clientAddress: ILP_ADDRESS
		}
	})

	plugin.registerDataHandler(handleRawPacket)

	await plugin.connect()

	let adminApp = new koa()
	adminApp.use(bodyParser())
	adminApp.use(function (ctx: koa.Context) {
		if (!ctx.path.trim().toLowerCase().startsWith("/credentials")) {
			ctx.throw(404)
		}

		if (ctx.method !== "POST") {
			ctx.throw(405)
		}

		const payload = ctx.request.body
		ctx.assert(isConnectionDetails(payload) && payload.paymentTag !== "", 400, "incomingPaymentID is required.")

		ctx.body = server.generateCredentials({ paymentTag: payload.paymentTag })
	})

	let adminServer = adminApp.listen(ADMIN_PORT)

	const gracefulShutdown = async function () {
		await plugin.disconnect()
		adminServer.close()
		await new Promise(resolve => {
			setTimeout(() => {
				adminServer.closeAllConnections()
			}, 1000)
		})
	}

	process.on("SIGINT", async () => {
		console.log("received SIGINT. shutting down...")
		await gracefulShutdown()
	})

	process.on("SIGKILL", async () => {
		console.log("received SIGKILL. shutting down...")
		await gracefulShutdown()
	})
}

type GenerateCredentialsPayload = {
	paymentTag: string
}

function isConnectionDetails(data: unknown): data is GenerateCredentialsPayload {
	if (typeof data === "object" && data !== null) {
		return typeof (data as any).paymentTag === "string"
	}

	return false
}

async function handleRawPacket(buffer: Buffer): Promise<Buffer> {
	try {
		let packet = await deserializeIlpPacket(buffer)
		if (packet.type !== Type.TYPE_ILP_PREPARE) {
			return Buffer.from('')
		}

		const prepare = packet.data
		const moneyOrReply = server.createReply(prepare)
		if (isIlpReply(moneyOrReply)) {
			return serializeIlpReply(moneyOrReply)
		}

		// make api call to backend
		const paymentTag = server.decodePaymentTag(prepare.destination)
		console.log("received prepare", prepare)

		return serializeIlpFulfill(moneyOrReply.accept())
	} catch (e) {
		console.error(e)
		return errorToReject(ILP_ADDRESS, new Errors.InternalError())
	}
}

main().catch(e => {
	console.error(e)
})
