import { ActionFunctionArgs, json } from '@remix-run/node'
import { Form, useActionData } from '@remix-run/react'
import OpenAI from 'openai'
import {
  ApplicationProps,
  Button,
  Card,
  Layouts,
  TextField
} from '~/components'
import { Prose } from '~/components/Content'
import { validateCSRFToken } from '~/lib/csrf.server'
import {commitSession, getSession} from "~/session.server";
import {MessageContentText} from "openai/resources/beta/threads";

const openai = new OpenAI({
  apiKey: process.env.OPEN_AI_TOKEN,
})

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {}
  }
}

export default function Page() {
  const actionData = useActionData<typeof action>()

  return (
    <>
      <Form id='search-gpt' action='/openai' method='post' className='hidden' />
      <Card>
        <TextField
          id='searchTerm'
          form='search-gpt'
          label='Search'
          name='searchTerm'
          type='text'
          className='mt-2'
          required
        />
      </Card>
      <Button form='search-gpt' type='submit'>
        Search
      </Button>
      {actionData?.response && <Prose>{actionData.response}</Prose>}
    </>
  )
}

export async function getThread(request: Request): Promise<string> {
  const session = await getSession(request.headers.get('Cookie'))

  let threadID = session.get('openai_thread')
  if (threadID) {
    return threadID
  }

  let thread = await openai.beta.threads.create()

  session.set('openai_thread', thread.id)
  await commitSession(session)

  return thread.id
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const searchTerm = form.get('searchTerm') as string

  await validateCSRFToken(request, form)

  let threadID = await getThread(request)

  await openai.beta.threads.messages.create(threadID, {
    role: 'user',
    content: searchTerm
  })

  // Use runs to wait for the assistant response and then retrieve it
  const run = await openai.beta.threads.runs.create(threadID, {
    assistant_id: process.env.OPEN_AI_ASSISTANT ?? ""
  })

  let runStatus = await openai.beta.threads.runs.retrieve(threadID, run.id)

  // Polling mechanism to see if runStatus is completed
  // This should be made more robust.
  while (runStatus.status !== 'completed') {
    await new Promise((resolve) => setTimeout(resolve, 2000))
    runStatus = await openai.beta.threads.runs.retrieve(threadID, run.id)
  }

  const messages = await openai.beta.threads.messages.list(threadID)

  const lastMessageForRun = messages.data
    .filter(
      (message) => message.run_id === run.id && message.role === 'assistant'
    )
    .pop()

  let msg = ''
  if (lastMessageForRun) {
    let content = lastMessageForRun.content[0] as MessageContentText
    msg = content.text.value as string
  }

  return json({
    searchTerm: '',
    response: msg
  })
}
