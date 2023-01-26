import { StreamServer } from '@interledger/stream-receiver'
import { randomBytes, randomUUID } from 'crypto'
import dotenv from 'dotenv'
import { deserializeIlpPacket, Errors, errorToReject, isIlpReply, serializeIlpFulfill, serializeIlpReply, Type } from "ilp-packet"
import PluginHttp from "ilp-plugin-http"
import { CCP_CONTROL_DESTINATION, CCP_UPDATE_DESTINATION, serializeCcpResponse } from 'ilp-protocol-ccp'
import { serializeIldcpResponse } from 'ilp-protocol-ildcp'
import koa from 'koa'
import bodyParser from 'koa-bodyparser'

const ENV_FILE = process.env.ENV_FILE || ""
if (ENV_FILE !== "") {
	const result = dotenv.config({ path: ENV_FILE, override: true })
	if (result.error) {
		throw result.error
	} 
}

const ILP_ADDRESS = process.env.ILP_ADDRESS || "t.fynbos"
const ASSET_CODE = process.env.ASSET_CODE || "USD"
const ASSET_SCALE = parseInt(process.env.ASSET_SCALE || "2")
const PEER_URL = process.env.OUTGOING_URL || "https://coil.test"
const SERVER_SECRET = process.env.SERVER_SECRET || randomBytes(32).toString("hex")
const INCOMING_SECRET = process.env.INCOMING_SECRET || ""
const OUTGOING_SECRET = process.env.OUTGOING_SECRET || ""
const ILP_PORT = parseInt(process.env.ILP_PORT || "8080")
const ADMIN_PORT = parseInt(process.env.ILP_PORT || "8081")

async function main() {
	console.log("ILP_PORT=", ILP_PORT)
	console.log("ADMIN_PORT=", ADMIN_PORT)
	console.log("ILP_ADDRESS=", ILP_ADDRESS)
	console.log("PEER_URL=", PEER_URL)
	console.log("ASSET_CODE=", ASSET_CODE)
	console.log("ASSET_SCALE=", ASSET_SCALE)
	console.log("ENV_FILE=", ENV_FILE)

	const server = new StreamServer({
		serverSecret: Buffer.from(SERVER_SECRET, 'hex'),
		serverAddress: ILP_ADDRESS,
	})

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

	plugin.registerDataHandler(handleRawPacket(server))

	await plugin.connect()

	let adminApp = new koa()
	adminApp.use(bodyParser())
	adminApp.use(function (ctx: koa.Context) {
		if (ctx.method === "GET" && ctx.path === '/healthz') {
			ctx.status = 200
			return
		}

		if (!ctx.path.trim().toLowerCase().startsWith("/credentials")) {
			ctx.throw(404)
		}

		if (ctx.method !== "POST") {
			ctx.throw(405)
		}

		const payload = ctx.request.body
		ctx.assert(
			isConnectionDetails(payload) && 
			payload.paymentTag !== "" && payload.asset.code !== "" && payload.asset.scale > 0,
			400, 
			"paymentTag and asset are required.",
		)

		const creds = server.generateCredentials({ paymentTag: payload.paymentTag, asset: payload.asset })
		ctx.body = {
			ilpAddress: creds.ilpAddress,
			sharedSecret: creds.sharedSecret.toString('hex')
		}
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
	asset: {
		code: string
		scale: number
	}
}

function isConnectionDetails(data: unknown): data is GenerateCredentialsPayload {
	if (typeof data === "object" && data !== null) {
		return typeof (data as any).paymentTag === "string" && isAsset((data as any).asset)
	}

	return false
}

function isAsset(data: unknown): data is { code: string, scale: number } {
	if (typeof data === "object" && data !== null) {
		return typeof (data as any).code === "string" && typeof (data as any).scale === "number"
	}

	return false
}

function handleRawPacket(server: StreamServer): (buf: Buffer) => Promise<Buffer> {
	return async function (buffer: Buffer): Promise<Buffer> {
		try {
			let packet = await deserializeIlpPacket(buffer)
			if (packet.type !== Type.TYPE_ILP_PREPARE) {
				return errorToReject(ILP_ADDRESS, new Errors.BadRequestError())
			}

			const prepare = packet.data
			// fulfill ccp requests
			if (
				prepare.destination === CCP_CONTROL_DESTINATION || 
				prepare.destination === CCP_UPDATE_DESTINATION
			) {
				return serializeCcpResponse()
			}

			// reject ildcp requests
			if (prepare.destination === "peer.config") {
				return serializeIldcpResponse({
					clientAddress: `${ILP_ADDRESS}.${randomUUID()}`,
					assetScale: ASSET_SCALE,
					assetCode: ASSET_CODE
				})
			} 

			const moneyOrReply = server.createReply(prepare)
			if (isIlpReply(moneyOrReply)) {
				return serializeIlpReply(moneyOrReply)
			}

			// make api call to backend
			const incomingPaymentID = server.decodePaymentTag(prepare.destination)

			return serializeIlpFulfill(moneyOrReply.accept())
		} catch (e) {
			console.error(e)
			return errorToReject(ILP_ADDRESS, new Errors.InternalError())
		}
	}
}

main().catch(e => {
	console.error(e)
})
