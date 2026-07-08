import { mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'
import { z } from 'zod'
import { parseConfig } from './parse.js'
import { ErrNoSecretClient } from './errors.js'
import { FakeSecretClient } from './fake.js'
import type { SecretClient } from './types.js'

const TestConfigSchema = z.object({
  database_url: z.string().min(1),
  api_key: z.string().default(''),
  port: z.number().default(0)
})

function writeYAML(content: string, name = 'config.yaml'): string {
  const dir = mkdtempSync(join(tmpdir(), 'configa-test-'))
  const file = join(dir, name)
  writeFileSync(file, content)
  return file
}

describe('parseConfig', () => {
  it('parses valid yaml', () => {
    const f = writeYAML(`
database_url: postgres://localhost/db
port: 5432
`)
    expect(() => parseConfig([f])).not.toThrow()
  })

  it('throws when the file does not exist', () => {
    expect(() => parseConfig(['/nonexistent/config.yaml'])).toThrow(
      /read config file/
    )
  })

  it('throws on invalid yaml', () => {
    const f = writeYAML('key: {unclosed')
    expect(() => parseConfig([f])).toThrow(/invalid yaml/)
  })

  it('throws when no filenames are given', () => {
    expect(() => parseConfig([])).toThrow(/at least one filename/)
  })
})

describe('ParsedConfig.resolve', () => {
  it('does not call the secret client when no templates are present', async () => {
    const f = writeYAML(`
database_url: postgres://localhost/db
port: 5432
`)
    const client = new FakeSecretClient()
    const cfg = parseConfig([f], { secretClient: client })
    const result = await cfg.resolve(TestConfigSchema)

    expect(result.database_url).toBe('postgres://localhost/db')
    expect(result.port).toBe(5432)
    expect(client.callCount.size).toBe(0)
  })

  it('resolves a single secret', async () => {
    const f = writeYAML(`
database_url: postgres://localhost/db
api_key: '{{ secret "my-secret" "key" }}'
`)
    const client = new FakeSecretClient({
      'my-secret': { key: 'super-secret-value' }
    })
    const cfg = parseConfig([f], { secretClient: client })
    const result = await cfg.resolve(TestConfigSchema)

    expect(result.api_key).toBe('super-secret-value')
  })

  it('dedupes multiple references to the same secret', async () => {
    const f = writeYAML(`
database_url: 'postgres://user:{{ secret "db-creds" "password" }}@localhost/db'
api_key: '{{ secret "db-creds" "api_key" }}'
port: 8080
`)
    const client = new FakeSecretClient({
      'db-creds': { password: 's3cr3t', api_key: 'key-value' }
    })
    const cfg = parseConfig([f], { secretClient: client })
    const result = await cfg.resolve(TestConfigSchema)

    expect(result.database_url).toBe('postgres://user:s3cr3t@localhost/db')
    expect(result.api_key).toBe('key-value')
    expect(client.callCount.get('db-creds')).toBe(1)
  })

  it('resolves a template embedded in a larger string', async () => {
    const f = writeYAML(`
database_url: 'postgres://user:{{ secret "db-creds" "password" }}@localhost/mydb'
port: 5432
`)
    const client = new FakeSecretClient({
      'db-creds': { password: 'p@ssw0rd' }
    })
    const cfg = parseConfig([f], { secretClient: client })
    const result = await cfg.resolve(TestConfigSchema)

    expect(result.database_url).toBe('postgres://user:p@ssw0rd@localhost/mydb')
  })

  it('resolves secrets nested inside maps', async () => {
    const NestedSchema = z.object({
      database: z.object({
        url: z.string().min(1),
        password: z.string().default('')
      })
    })
    const f = writeYAML(`
database:
  url: postgres://localhost/db
  password: '{{ secret "db-secret" "password" }}'
`)
    const client = new FakeSecretClient({
      'db-secret': { password: 'nested-pass' }
    })
    const cfg = parseConfig([f], { secretClient: client })
    const result = await cfg.resolve(NestedSchema)

    expect(result.database.password).toBe('nested-pass')
  })

  it('throws when a referenced secret cannot be fetched', async () => {
    const f = writeYAML(`
database_url: postgres://localhost/db
api_key: '{{ secret "missing-secret" "key" }}'
`)
    const client = new FakeSecretClient()
    const cfg = parseConfig([f], { secretClient: client })

    await expect(cfg.resolve(TestConfigSchema)).rejects.toThrow()
  })

  it('throws when the secret exists but the key is missing', async () => {
    const f = writeYAML(`
database_url: postgres://localhost/db
api_key: '{{ secret "my-secret" "nonexistent-key" }}'
`)
    const client = new FakeSecretClient({
      'my-secret': { 'other-key': 'value' }
    })
    const cfg = parseConfig([f], { secretClient: client })

    await expect(cfg.resolve(TestConfigSchema)).rejects.toThrow(
      /nonexistent-key/
    )
  })

  it('throws on schema validation failure', async () => {
    const f = writeYAML(`
database_url: ""
port: 5432
`)
    const cfg = parseConfig([f])
    await expect(cfg.resolve(TestConfigSchema)).rejects.toThrow(/configa:/)
  })

  it('throws on an unsupported template expression', async () => {
    const f = writeYAML(`
database_url: postgres://localhost/db
api_key: '{{ unknown "arg" }}'
`)
    const client = new FakeSecretClient()
    const cfg = parseConfig([f], { secretClient: client })

    await expect(cfg.resolve(TestConfigSchema)).rejects.toThrow(/configa:/)
  })

  it('throws ErrNoSecretClient when templates are present but no client was given', async () => {
    const f = writeYAML(`
database_url: postgres://localhost/db
api_key: '{{ secret "my-secret" "key" }}'
`)
    const cfg = parseConfig([f])

    await expect(cfg.resolve(TestConfigSchema)).rejects.toBeInstanceOf(
      ErrNoSecretClient
    )
  })

  it('passes the configured namespace through to the secret client', async () => {
    const f = writeYAML(`
database_url: postgres://localhost/db
api_key: '{{ secret "my-secret" "key" }}'
`)
    let capturedNamespace = ''
    const client: SecretClient = {
      async getSecret(namespace, name) {
        capturedNamespace = namespace
        if (name === 'my-secret') return { key: 'value' }
        throw new Error('unexpected secret')
      }
    }
    const cfg = parseConfig([f], {
      secretClient: client,
      namespace: 'custom-namespace'
    })
    await cfg.resolve(TestConfigSchema)

    expect(capturedNamespace).toBe('custom-namespace')
  })

  it('resolves secrets embedded in array items', async () => {
    const ArraySchema = z.object({ items: z.array(z.string()) })
    const f = writeYAML(`
items:
  - plain-value
  - '{{ secret "list-secret" "item" }}'
`)
    const client = new FakeSecretClient({
      'list-secret': { item: 'resolved-item' }
    })
    const cfg = parseConfig([f], { secretClient: client })
    const result = await cfg.resolve(ArraySchema)

    expect(result.items).toEqual(['plain-value', 'resolved-item'])
  })
})

describe('overlay merging', () => {
  it('lets an overlay replace a top-level scalar', async () => {
    const base = writeYAML(
      `
database_url: postgres://base/db
port: 5432
`,
      'base.yaml'
    )
    const overlay = writeYAML(
      `
port: 9999
`,
      'overlay.yaml'
    )
    const cfg = parseConfig([base, overlay])
    const result = await cfg.resolve(TestConfigSchema)

    expect(result.database_url).toBe('postgres://base/db')
    expect(result.port).toBe(9999)
  })

  it('deep-merges nested maps, preserving sibling keys', async () => {
    const NestedSchema = z.object({
      database: z.object({
        url: z.string().min(1),
        host: z.string().default(''),
        port: z.number().default(0)
      })
    })
    const base = writeYAML(
      `
database:
  url: postgres://base/db
  host: base-host
  port: 5432
`,
      'base.yaml'
    )
    const overlay = writeYAML(
      `
database:
  host: overlay-host
`,
      'overlay.yaml'
    )
    const cfg = parseConfig([base, overlay])
    const result = await cfg.resolve(NestedSchema)

    expect(result.database.url).toBe('postgres://base/db')
    expect(result.database.host).toBe('overlay-host')
    expect(result.database.port).toBe(5432)
  })

  it('lets the last of three files win', async () => {
    const base = writeYAML(
      `
database_url: postgres://base/db
port: 1111
`,
      'base.yaml'
    )
    const mid = writeYAML(
      `
port: 2222
`,
      'mid.yaml'
    )
    const top = writeYAML(
      `
port: 3333
`,
      'top.yaml'
    )
    const cfg = parseConfig([base, mid, top])
    const result = await cfg.resolve(TestConfigSchema)

    expect(result.port).toBe(3333)
  })

  it('resolves a template introduced by an overlay file', async () => {
    const base = writeYAML(
      `
database_url: postgres://base/db
port: 5432
`,
      'base.yaml'
    )
    const overlay = writeYAML(
      `
api_key: '{{ secret "my-secret" "key" }}'
`,
      'overlay.yaml'
    )
    const client = new FakeSecretClient({
      'my-secret': { key: 'overlay-secret-value' }
    })
    const cfg = parseConfig([base, overlay], { secretClient: client })
    const result = await cfg.resolve(TestConfigSchema)

    expect(result.api_key).toBe('overlay-secret-value')
  })

  it('throws when an overlay file does not exist', () => {
    const base = writeYAML(`
database_url: postgres://base/db
port: 5432
`)
    expect(() =>
      parseConfig([base, '/nonexistent/overlay.yaml'])
    ).toThrow(/read config file/)
  })
})
