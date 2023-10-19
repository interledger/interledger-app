import { redirect } from '@remix-run/node'
import { route } from 'routes-gen'

export async function loader() {
  return redirect(route('/payments'))
}
