import type { Route } from './+types/api_.pti_.widget-config'
import { envValue } from '~/env.server'

function resolvePtiWidgetConfig() {
  let sdkUrl = envValue("PTI_SDK_URL") || 'https://sdk.staging.fiant.io/latest/index.js'
  let formsUrl = envValue("PTI_FORMS_URL") || 'https://forms.staging.fiant.io'

  if (envValue("FYNBOS_ENV") === 'prod') {
    sdkUrl = envValue("PTI_SDK_URL") || 'https://sdk.platform.fiant.io/0.0.23/index.js'
    formsUrl = envValue("PTI_FORMS_URL") || 'https://forms.platform.fiant.io'
  }

  if (envValue("FYNBOS_ENV") === 'local') {
    sdkUrl = envValue("PTI_SDK_URL") || 'https://mockpti.interledger.test/sdk/index.js'
    formsUrl = envValue("PTI_FORMS_URL") || 'https://mockpti.interledger.test/forms'
  }

  return { sdkUrl, formsUrl }
}

export async function loader({}: Route.LoaderArgs) {
  return Response.json(resolvePtiWidgetConfig())
}