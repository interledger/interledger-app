import type { LoaderFunctionArgs } from '@remix-run/node'

export const loader = async ({ request }: LoaderFunctionArgs) => {
  const appID = process.env.APPLE_APP_ID

  if (!appID) {
    return new Response('APPLE_APP_ID not configured', { status: 500 })
  }

  const aasa = {
    applinks: {
      apps: [],
      details: [
        {
          appIDs: [appID],
          components: [{ '/': '/*' }]
        }
      ]
    },
    webcredentials: {
      apps: [appID]
    }
  }

  return new Response(JSON.stringify(aasa), {
    headers: { 'Content-Type': 'application/json' }
  })
}
