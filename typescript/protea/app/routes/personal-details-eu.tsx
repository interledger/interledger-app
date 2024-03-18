import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData, useSearchParams, useSubmit } from '@remix-run/react'
import { useEffect, useRef, useState } from 'react'
import { route } from 'routes-gen'
import { Button, Card, CardContent, Layouts, Shape } from '~/components'
import { isConnectError } from '~/lib/error.server'
import { exitFlow, flowType, requireFlow } from '~/lib/flows.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { useScript } from '~/lib/useScript'


export const handle = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Activate wallet',
      back: route('/')
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Activate wallet'
  }
])


export default function Page() {
  const [params] = useSearchParams()

  return (
    <>
      <iframe
        src={`https://onboarding.sandbox.gatehub.net?bearer=${params.get('token')}`}
        sandbox="allow-top-navigation allow-forms allow-same-origin allow-popups allow-scripts"
        scrolling="no"
        frameBorder="0"
        allow="camera;microphone"
        className='w-full'
      ></iframe>
    </>
  )
}
