import { href, redirect } from 'react-router'

export async function loader() {
  return redirect(href('/payments'))
}
