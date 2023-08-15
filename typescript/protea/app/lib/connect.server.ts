import { createPromiseClient } from '@bufbuild/connect'
import { createGrpcWebTransport } from '@bufbuild/connect-web'

import { BackendService } from '~/generated/connect/backend/v1/backend_connect'

const BACKEND_GRPC_URL = process.env.BACKEND_GRPC_URL || 'dns:backend:443'

// Import service definition that you want to connect to.

// The transport defines what type of endpoint we're hitting.
// In our example we'll be communicating with a Connect endpoint.
// If your endpoint only supports gRPC-web, make sure to use
// `createGrpcWebTransport` instead.
const transport = createGrpcWebTransport({
  baseUrl: BACKEND_GRPC_URL
})

// Here we make the client itself, combining the service
// definition with the transport.
const grpcConnectClient = createPromiseClient(BackendService, transport)

export { grpcConnectClient }
