import { mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import * as https from 'node:https'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import {
  ErrSecretFetchFailed,
  ErrSecretForbidden,
  ErrSecretNotFound
} from './errors.js'
import { InClusterSecretClient, newInClusterSecretClient } from './k8s.js'

const testdataDir = join(fileURLToPath(import.meta.url), '..', 'testdata')
const serverCert = readFileSync(join(testdataDir, 'test-server-cert.pem'))
const serverKey = readFileSync(join(testdataDir, 'test-server-key.pem'))

type Secrets = Record<string, Record<string, string>>

function startServer(
  secrets: Secrets,
  statusCode?: number
): Promise<https.Server> {
  const server = https.createServer(
    { key: serverKey, cert: serverCert },
    (req, res) => {
      if (statusCode) {
        res.writeHead(statusCode)
        res.end()
        return
      }

      const parts = (req.url ?? '').split('/').filter(Boolean)
      // Expected: api / v1 / namespaces / {ns} / secrets / {name}
      if (
        parts.length !== 6 ||
        parts[0] !== 'api' ||
        parts[2] !== 'namespaces' ||
        parts[4] !== 'secrets'
      ) {
        res.writeHead(404)
        res.end()
        return
      }
      const [, , , ns, , name] = parts
      const data = secrets[`${ns}/${name}`]
      if (!data) {
        res.writeHead(404)
        res.end()
        return
      }
      const encoded: Record<string, string> = {}
      for (const [k, v] of Object.entries(data)) {
        encoded[k] = Buffer.from(v, 'utf8').toString('base64')
      }
      res.writeHead(200, { 'content-type': 'application/json' })
      res.end(JSON.stringify({ data: encoded }))
    }
  )
  return new Promise((resolve) => {
    server.listen(0, '127.0.0.1', () => resolve(server))
  })
}

interface Fixture {
  tokenPath: string
  caPath: string
  namespacePath: string
  dir: string
}

function writeCredentials(namespace = 'test-ns'): Fixture {
  const dir = mkdtempSync(join(tmpdir(), 'configa-k8s-test-'))
  const tokenPath = join(dir, 'token')
  const caPath = join(dir, 'ca.crt')
  const namespacePath = join(dir, 'namespace')
  writeFileSync(tokenPath, 'test-bearer-token')
  writeFileSync(caPath, serverCert)
  writeFileSync(namespacePath, namespace)
  return { tokenPath, caPath, namespacePath, dir }
}

let server: https.Server | undefined
let fixture: Fixture | undefined
const originalHost = process.env.KUBERNETES_SERVICE_HOST
const originalPort = process.env.KUBERNETES_SERVICE_PORT

async function setup(secrets: Secrets, statusCode?: number) {
  server = await startServer(secrets, statusCode)
  const address = server.address()
  if (address === null || typeof address === 'string') {
    throw new Error('expected server to bind a port')
  }
  process.env.KUBERNETES_SERVICE_HOST = '127.0.0.1'
  process.env.KUBERNETES_SERVICE_PORT = String(address.port)
  fixture = writeCredentials()
  return fixture
}

afterEach(async () => {
  if (server) {
    await new Promise((resolve) => server?.close(resolve))
    server = undefined
  }
  if (fixture) {
    rmSync(fixture.dir, { recursive: true, force: true })
    fixture = undefined
  }
  if (originalHost === undefined) delete process.env.KUBERNETES_SERVICE_HOST
  else process.env.KUBERNETES_SERVICE_HOST = originalHost
  if (originalPort === undefined) delete process.env.KUBERNETES_SERVICE_PORT
  else process.env.KUBERNETES_SERVICE_PORT = originalPort
})

describe('newInClusterSecretClient', () => {
  it('never throws on construction', () => {
    expect(() => newInClusterSecretClient()).not.toThrow()
  })
})

describe('InClusterSecretClient credential errors', () => {
  it('surfaces a missing token file on getSecret, not construction', async () => {
    const client = new InClusterSecretClient({
      tokenPath: '/nonexistent/token',
      caPath: '/nonexistent/ca.crt',
      namespacePath: '/nonexistent/namespace'
    })
    await expect(client.getSecret('ns', 'name')).rejects.toThrow(
      /service account token/
    )
  })

  it('surfaces a missing CA cert', async () => {
    const dir = mkdtempSync(join(tmpdir(), 'configa-k8s-test-'))
    const tokenPath = join(dir, 'token')
    writeFileSync(tokenPath, 'tok')

    const client = new InClusterSecretClient({
      tokenPath,
      caPath: '/nonexistent/ca.crt',
      namespacePath: '/nonexistent/namespace'
    })
    await expect(client.getSecret('ns', 'name')).rejects.toThrow(/CA cert/)
    rmSync(dir, { recursive: true, force: true })
  })

  it('surfaces missing KUBERNETES_SERVICE_HOST/PORT', async () => {
    const dir = mkdtempSync(join(tmpdir(), 'configa-k8s-test-'))
    const tokenPath = join(dir, 'token')
    const caPath = join(dir, 'ca.crt')
    const namespacePath = join(dir, 'namespace')
    writeFileSync(tokenPath, 'tok')
    writeFileSync(caPath, serverCert)
    writeFileSync(namespacePath, 'test-ns')

    delete process.env.KUBERNETES_SERVICE_HOST
    delete process.env.KUBERNETES_SERVICE_PORT

    const client = new InClusterSecretClient({
      tokenPath,
      caPath,
      namespacePath
    })
    await expect(client.getSecret('ns', 'name')).rejects.toThrow(
      /KUBERNETES_SERVICE_HOST/
    )
    rmSync(dir, { recursive: true, force: true })
  })
})

describe('InClusterSecretClient.getSecret', () => {
  it('fetches and base64-decodes secret data', async () => {
    const f = await setup({
      'test-ns/my-secret': { password: 'secret-value', user: 'admin' }
    })
    const client = new InClusterSecretClient(f)

    const data = await client.getSecret('test-ns', 'my-secret')
    expect(data.password).toBe('secret-value')
    expect(data.user).toBe('admin')
  })

  it('uses the service account namespace when none is given', async () => {
    const f = await setup({ 'test-ns/my-secret': { key: 'value' } })
    const client = new InClusterSecretClient(f)

    const data = await client.getSecret('', 'my-secret')
    expect(data.key).toBe('value')
  })

  it('maps 404 to ErrSecretNotFound', async () => {
    const f = await setup({})
    const client = new InClusterSecretClient(f)

    await expect(
      client.getSecret('test-ns', 'nonexistent')
    ).rejects.toBeInstanceOf(ErrSecretNotFound)
  })

  it('maps 403 to ErrSecretForbidden', async () => {
    const f = await setup({}, 403)
    const client = new InClusterSecretClient(f)

    await expect(
      client.getSecret('test-ns', 'my-secret')
    ).rejects.toBeInstanceOf(ErrSecretForbidden)
  })

  it('maps 401 to ErrSecretForbidden', async () => {
    const f = await setup({}, 401)
    const client = new InClusterSecretClient(f)

    await expect(
      client.getSecret('test-ns', 'my-secret')
    ).rejects.toBeInstanceOf(ErrSecretForbidden)
  })

  it('maps 5xx to ErrSecretFetchFailed', async () => {
    const f = await setup({}, 500)
    const client = new InClusterSecretClient(f)

    await expect(
      client.getSecret('test-ns', 'my-secret')
    ).rejects.toBeInstanceOf(ErrSecretFetchFailed)
  })

  it('loads credentials only once across repeated calls', async () => {
    const f = await setup({
      'test-ns/s1': { k: 'v1' },
      'test-ns/s2': { k: 'v2' }
    })
    const client = new InClusterSecretClient(f)

    await expect(client.getSecret('test-ns', 's1')).resolves.toBeTruthy()

    // Remove the token file after the first successful call — a second call
    // must still succeed because credentials are cached (memoized promise).
    rmSync(f.tokenPath)

    await expect(client.getSecret('test-ns', 's2')).resolves.toBeTruthy()
  })
})
