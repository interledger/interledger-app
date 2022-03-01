import React from 'react'
import { Form } from 'remix'
import { TextField, Button } from '~/components'
import { getCsrfTokenFromFlow } from '~/lib/kratos'

export default function SignupPage() {
  return (
    <main className='mx-auto grid min-h-screen w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 lg:content-center xl:max-w-4xl'>
      {/* Header */}
      <div className='col-span-full flex h-10 flex-col bg-primary sm:col-span-6 sm:col-start-2 lg:col-start-4'></div>
      <div className='col-span-full flex h-10 flex-col bg-container sm:col-span-6 sm:col-start-2 lg:col-start-4'></div>
      <div className='col-span-full flex h-10 flex-col bg-container-primary sm:col-span-6 sm:col-start-2 lg:col-start-4'></div>
      {/* Form */}
      <div className='col-span-full flex h-60 flex-col bg-yellow-300 sm:col-span-6 sm:col-start-2 lg:col-start-4'></div>
      <Form
        action={`/test`}
        method='post'
        className='flex flex-col col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4 items-end'
      >
        <TextField
          id='email'
          label='done'
          enterKeyHint='done'
          name='email'
          type='email'
          required
        />
        <TextField
          id='email'
          label='enter'
          enterKeyHint='enter'
          name='email'
          type='email'
          required
        />
        <TextField
          id='email'
          label='go'
          enterKeyHint='go'
          name='email'
          type='email'
          required
        />
        <TextField
          id='email'
          label='next'
          enterKeyHint='next'
          name='email'
          type='email'
          required
        />
        <TextField
          id='email'
          label='previous'
          enterKeyHint='previous'
          name='email'
          type='email'
          required
        />
        <TextField
          id='email'
          label='search'
          enterKeyHint='search'
          name='email'
          type='email'
          required
        />
        <TextField
          id='email'
          label='send'
          enterKeyHint='send'
          name='email'
          type='email'
          required
        />

        <Button type='submit'>Create account</Button>
      </Form>
    </main>
  )
}
