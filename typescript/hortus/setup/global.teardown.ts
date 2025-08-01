import { STATE_FILE } from 'fixtures/helpers'
import fs from 'node:fs'

export default async function () {
  await fs.promises.rm(STATE_FILE)
}
