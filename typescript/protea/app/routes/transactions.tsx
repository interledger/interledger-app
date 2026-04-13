import { redirect } from 'react-router';
import { href } from 'react-router'

export async function loader() {
  return redirect(href('/payments'))
}
