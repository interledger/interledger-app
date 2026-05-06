import type { FetcherWithComponents } from 'react-router'
import { Button, TextButton } from './Buttons'
import type { PhoneAutocompleteOptions } from './PhoneTextField'
import { PhoneTextField } from './PhoneTextField'

type PhoneUpdateFetcher = FetcherWithComponents<{
  errors?: { phone?: string }
}>

interface ChangePhoneFormProps {
  fetcher: PhoneUpdateFetcher
  csrfToken: string
  defaultCountry: string
  countries: PhoneAutocompleteOptions[]
  /** Defaults to the current route when omitted */
  action?: string
  onCancel?: () => void
  submitLabel?: string
  className?: string
}

export function ChangePhoneForm({
  fetcher,
  csrfToken,
  defaultCountry,
  countries,
  action,
  onCancel,
  submitLabel = 'Save phone & send code',
  className
}: ChangePhoneFormProps) {
  return (
    <fetcher.Form method='post' action={action} className={className}>
      <input type='hidden' name='intent' value='updatePhone' />
      <input type='hidden' name='csrfToken' value={csrfToken} />
      <PhoneTextField
        id='newPhone'
        name='phone'
        defaultCountry={defaultCountry}
        options={countries}
        label='New mobile number'
        className='mt-2'
        errorMessage={fetcher.data?.errors?.phone}
      />
      <div className='mt-3 flex space-x-3'>
        <Button type='submit' disabled={fetcher.state !== 'idle'}>
          {fetcher.state !== 'idle' ? 'Saving...' : submitLabel}
        </Button>
        {onCancel && (
          <TextButton type='button' onClick={onCancel}>
            Cancel
          </TextButton>
        )}
      </div>
    </fetcher.Form>
  )
}
