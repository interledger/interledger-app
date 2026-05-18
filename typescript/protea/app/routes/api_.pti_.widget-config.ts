import type { Route } from './+types/api_.pti_.widget-config'

function resolvePtiWidgetConfig() {
  let sdkUrl = process.env.PTI_SDK_URL || 'https://sdk.staging.fiant.io/latest/index.js'
  let formsUrl = process.env.PTI_FORMS_URL || 'https://forms.staging.fiant.io'

  // if (process.env.FYNBOS_ENV === 'prod') {
  //   sdkUrl = process.env.PTI_SDK_URL || 'https://sdk.platform.fiant.io/0.0.23/index.js'
  //   formsUrl = process.env.PTI_FORMS_URL || 'https://forms.platform.fiant.io'
  // }

  // if (process.env.FYNBOS_ENV === 'local') {
  //   sdkUrl = process.env.PTI_SDK_URL || 'https://mockpti.interledger.test/sdk/index.js'
  //   formsUrl = process.env.PTI_FORMS_URL || 'https://mockpti.interledger.test/forms'
  // }

  return { sdkUrl, formsUrl }
}

export async function loader({}: Route.LoaderArgs) {
  return Response.json(resolvePtiWidgetConfig())
}