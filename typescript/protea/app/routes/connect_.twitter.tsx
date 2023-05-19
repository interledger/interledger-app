import {useEffect} from 'react'
import type { LoaderArgs } from '@remix-run/node'
import {ActionArgs, json, redirect} from '@remix-run/node'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import {RouteMatch, useFetcher, useLoaderData} from "@remix-run/react";
import {Layouts} from "~/components";

export const handle = {
  title: "Twitter",
  layout: Layouts.FocusLayout,
}

export async function loader({ request }: LoaderArgs) {
  // Check for state and code if not state create auth url
  let url = new URL(request.url)
  let state = url.searchParams.get('state')
  let code = url.searchParams.get('code')

  if (state && code) {
    return json({
      state: state,
      code: code
    })
  } else {
    let resp = await grpcClient
      .createTwitterAuthURL(
        {},
        {
          meta: {
            cookies: String(request.headers.get('cookie')) || ''
          }
        }
      )
      .then((resp) => resp.response)
      .catch(StatusError)

    if (isGrpcError(resp)) {
      throw json({}, httpMapping(resp.code))
    }

    return redirect(resp.url)
  }
}

export default function Page() {
  // get state and code from loader data
  let { state, code } = useLoaderData<typeof loader>()
  const fetcher = useFetcher();

  // use effect for the state and code to call the action programmatically
  useEffect(() => {
    console.log(state, code, "state and code")
    if (state && code) {
      // call action
      fetcher.submit({ state: state, code: code }, { method: "post" });
    }
  }, [state, code])

  // handle the action response
  useEffect(() => {

  }, [fetcher.data])

  return (
      <div>

      </div>
  );
}

export async function action({ request }: ActionArgs) {
  const formData = await request.formData();
  const state = formData.get('state')
  const code = formData.get('code')

  if (!state || !code) {
    return json({}, httpMapping(400))
  }

  console.log(state,code)
  return "hello"
  // let resp = await grpcClient.twitterCallback({
  //   state: state.toString(),
  //   code: code.toString()
  // }, {
  //   meta: {
  //     cookies: String(request.headers.get('cookie')) || ''
  //   }
  // })

  // tell window to close itself?
}