import { json, LoaderArgs } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useEffect } from 'react'
import { useScript } from '~/lib/useScript'

export async function loader({ request }: LoaderArgs) {
  // await requireUserSession(request)
  // let rpc = await grpcClient.getMachnetWidgetToken({}, {
  //   meta: {
  //     "cookies": request.headers.get("cookie") || ""
  //   }
  // }).then(v => v).catch(StatusError)
  // if (isGrpcError(rpc)) {
  //   throw json({}, httpMapping(rpc.code))
  // }

  let widgetScriptUrl = 'https://widget.v4sandbox.machpay.com/widget/widget.js'
  return json({
    widgetScriptUrl,
    widgetUserId: "", //rpc.response.userId,
    widgetToken: "", //rpc.response.value
  })
}

export default function Page() {
  const data = useLoaderData<typeof loader>()
  const scriptStatus = useScript(data.widgetScriptUrl)
  useEffect(() => {
    if (scriptStatus === 'ready') {
      var widget = new (window as any).MachnetWidget({
        elementId: 'widget',
        userId: data.widgetUserId,
        width: '100%',
        height: '200px',
        type: 'card',
        locale: 'en',
        token: data.widgetToken
      })
      widget.init()

      window.addEventListener("message", (event) => {
        console.log(event)
      });
    }
  }, [scriptStatus])

  return (
    <div id='widget' className='flex h-screen w-full items-center justify-center font-display text-5xl font-medium text-medium' />
  )
}
