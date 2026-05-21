import type { Route } from './+types/api_.pti_.widget-config'
import { envValue } from '~/env.server'

function resolvePtiWidgetConfig() {
  const sdkUrl = envValue("PTI_SDK_URL")
  const formsUrl = envValue("PTI_FORMS_URL")

  return { sdkUrl, formsUrl }
}

export async function loader({}: Route.LoaderArgs) {
  return Response.json(resolvePtiWidgetConfig())
}