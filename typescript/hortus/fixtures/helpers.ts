import path from 'node:path'

export const ROOT_DIR = path.resolve(__dirname, '..')
export const STATE_FILE = path.join(ROOT_DIR, '.auth', 'state.json')
