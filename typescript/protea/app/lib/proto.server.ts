import { ChannelCredentials } from '@grpc/grpc-js'
import { BackendServiceClient } from '~/generated/protobuf-ts/backend/v1/backend_client'
import { GrpcTransport } from '@protobuf-ts/grpc-transport'

const transport = new GrpcTransport({
  host: 'dns:backend-admin:443',
  channelCredentials: ChannelCredentials.createInsecure()
})

let grpcClient: BackendServiceClient

declare global {
  var __grpcClient: BackendServiceClient | undefined
}

// this is needed because in development we don't want to restart
// the server with every change, but we want to make sure we don't
// create a new connection to the Client with every change either.
if (process.env.NODE_ENV === 'production') {
  grpcClient = new BackendServiceClient(transport)
} else {
  if (!global.__grpcClient) {
    global.__grpcClient = new BackendServiceClient(transport)
  }
  grpcClient = global.__grpcClient
}

export { grpcClient }
