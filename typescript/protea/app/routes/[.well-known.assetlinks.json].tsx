import type { LoaderFunctionArgs } from '@remix-run/node'

export const loader = async ({ request }: LoaderFunctionArgs) => {
  const packageName = process.env.ANDROID_PACKAGE_NAME
  const sha256 = process.env.ANDROID_SHA256

  if (!packageName || !sha256) {
    return new Response('ANDROID_PACKAGE_NAME or ANDROID_SHA256 not configured', {
      status: 500
    })
  }

  const assetlinks = [
    {
      relation: [
        'delegate_permission/common.handle_all_urls',
        'delegate_permission/common.get_login_creds'
      ],
      target: {
        namespace: 'android_app',
        package_name: packageName,
        sha256_cert_fingerprints: [sha256]
      }
    }
  ]

  return new Response(JSON.stringify(assetlinks), {
    headers: { 'Content-Type': 'application/json' }
  })
}
