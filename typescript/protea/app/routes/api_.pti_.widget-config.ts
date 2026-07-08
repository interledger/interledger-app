import { config } from '~/config.server'
import type { Route } from './+types/api_.pti_.widget-config'

function resolvePtiWidgetConfig() {
  const sdkUrl = config.pti.sdk_url
  const formsUrl = config.pti.forms_url

  return { sdkUrl, formsUrl }
}

export async function loader(_args: Route.LoaderArgs) {
  return Response.json(resolvePtiWidgetConfig())
}
