import { zodResolver } from '@hookform/resolvers/zod'
import { Form } from '@remix-run/react'
import { Controller, useForm } from 'react-hook-form'
import * as z from 'zod/mini'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  OutlineButton,
  Select,
  TextField,
  type SelectOptions
} from '~/components'
import {
  CustomerDeliveryAddressType,
  type NewCustomerDeliveryAddress
} from '~/generated/connect/backend/v1/backend_pb'
import { OrderCardStep, useOrderCardStore } from '~/lib/useOrderCardStore'
import {
  createNewAddress,
  getAddressTypeValue,
  getCountryOptions,
  getCountryValue
} from './utils'

export const addressTypeOptions = [
  { id: 'other', name: 'Other' },
  { id: 'work', name: 'Work' }
]

const addressFormSchema = z.object({
  details: z.object({
    type: z.nativeEnum(CustomerDeliveryAddressType),
    country: z.string().check(z.minLength(1, 'Country is required')),
    line1: z
      .string()
      .check(z.minLength(1, 'Address line 1 is required'))
      .check(z.maxLength(128, 'Address line 1 is maximum 128 characters long')),
    line2: z.optional(
      z
        .string()
        .check(
          z.maxLength(128, 'Address line 2 is maximum 128 characters long')
        )
    ),
    line3: z.optional(
      z
        .string()
        .check(
          z.maxLength(128, 'Address line 3 is maximum 128 characters long')
        )
    ),
    postOffice: z.nullable(
      z.optional(
        z
          .string()
          .check(z.maxLength(32, 'Post office is maximum 32 characters long'))
      )
    ),
    city: z
      .string()
      .check(z.minLength(1, 'City is required'))
      .check(z.maxLength(32, 'City is maximum 32 characters long')),
    zipCode: z
      .string()
      .check(z.minLength(1, 'Zip code is required'))
      .check(z.maxLength(8, 'Zip code is maximum 8 characters long'))
      .check(z.regex(/^\d+$/, 'Zip code must contain digits only'))
  }),
  reason: z
    .string()
    .check(z.minLength(1, 'Reason is required'))
    .check(z.maxLength(256, 'Reason is maximum 256 characters long'))
})
export type AddressFormData = z.infer<typeof addressFormSchema>

const defaultValues: AddressFormData = {
  details: {
    type: CustomerDeliveryAddressType.CUSTOMER_DELIVERY_ADDRESS_OTHER,
    country: '',
    line1: '',
    line2: '',
    line3: '',
    postOffice: '',
    city: '',
    zipCode: ''
  },
  reason: ''
}

const getInitialValues = (
  newAddress: NewCustomerDeliveryAddress | null
): AddressFormData => {
  if (newAddress) {
    return {
      details: {
        type:
          newAddress.details?.type ||
          CustomerDeliveryAddressType.CUSTOMER_DELIVERY_ADDRESS_OTHER,
        country: newAddress.details?.countryCode || '',
        line1: newAddress.details?.line1 || '',
        line2: newAddress.details?.line2 || '',
        line3: newAddress.details?.line3 || '',
        postOffice: newAddress.details?.postOffice || '',
        city: newAddress.details?.city || '',
        zipCode: newAddress.details?.zipCode || ''
      },
      reason: newAddress.reason || ''
    }
  }
  return defaultValues
}

export const CreateAddress = () => {
  const { newAddress, setNewAddress, setStep, countries } = useOrderCardStore()

  const {
    control,
    handleSubmit,
    watch,
    setValue,
    formState: { errors, isValid }
  } = useForm<AddressFormData>({
    defaultValues: getInitialValues(newAddress),
    mode: 'onChange',
    resolver: zodResolver(addressFormSchema)
  })

  const handleAddressTypeChange = (option: SelectOptions) => {
    let type: CustomerDeliveryAddressType
    switch (option.id) {
      case 'work':
        type = CustomerDeliveryAddressType.CUSTOMER_DELIVERY_ADDRESS_WORK
        break
      default:
        type = CustomerDeliveryAddressType.CUSTOMER_DELIVERY_ADDRESS_OTHER
    }
    setValue('details.type', type)
  }

  const handleCountryChange = (option: SelectOptions) => {
    setValue('details.country', option.id)
  }

  const onSubmit = (data: AddressFormData) => {
    const newAddress = createNewAddress(data)
    setNewAddress(newAddress)
    setStep(OrderCardStep.DELIVERY)
  }

  return (
    <Form method='post' onSubmit={handleSubmit(onSubmit)}>
      <Card>
        <CardContent>
          <CardHeader>
            <p className='text-medium'>
              Please provide your delivery address details
            </p>
          </CardHeader>

          <div className='space-y-6'>
            {/* Address Type */}
            <Select
              id='addressType'
              label='Address Type'
              value={getAddressTypeValue(watch('details.type'))}
              options={addressTypeOptions}
              onChange={handleAddressTypeChange}
              errorMessage={errors.details?.type?.message}
            />

            {/* Country */}
            <Select
              id='country'
              label='Country'
              value={getCountryValue(watch('details.country'), countries)}
              options={getCountryOptions(countries)}
              onChange={handleCountryChange}
              errorMessage={errors.details?.country?.message}
            />

            {/* Address Line 1 */}
            <Controller
              name='details.line1'
              control={control}
              rules={{
                required: 'Address line 1 is required'
              }}
              render={({ field }) => (
                <TextField
                  id='line1'
                  label='Address Line 1'
                  placeholder='123 Main Street'
                  value={field.value}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                  errorMessage={errors.details?.line1?.message}
                  required
                />
              )}
            />

            {/* Address Line 2 */}
            <Controller
              name='details.line2'
              control={control}
              render={({ field }) => (
                <TextField
                  id='line2'
                  label='Address Line 2 (Optional)'
                  placeholder='Apt 4B'
                  value={field.value}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                  errorMessage={errors.details?.line2?.message}
                />
              )}
            />

            {/* Address Line 3 */}
            <Controller
              name='details.line3'
              control={control}
              render={({ field }) => (
                <TextField
                  id='line3'
                  label='Address Line 3 (Optional)'
                  placeholder='Building C'
                  value={field.value}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                  errorMessage={errors.details?.line3?.message}
                />
              )}
            />

            {/* Post Office */}
            <Controller
              name='details.postOffice'
              control={control}
              render={({ field }) => (
                <TextField
                  id='postOffice'
                  label='Post Office (Optional)'
                  placeholder='Main Post Office'
                  value={field.value || ''}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                  errorMessage={errors.details?.postOffice?.message}
                />
              )}
            />

            {/* City */}
            <Controller
              name='details.city'
              control={control}
              rules={{
                required: 'City is required'
              }}
              render={({ field }) => (
                <TextField
                  id='city'
                  label='City'
                  placeholder='New York'
                  value={field.value}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                  errorMessage={errors.details?.city?.message}
                  required
                />
              )}
            />

            {/* Zip Code */}
            <Controller
              name='details.zipCode'
              control={control}
              rules={{
                required: 'Zip code is required'
              }}
              render={({ field }) => (
                <TextField
                  id='zipCode'
                  label='Zip Code'
                  placeholder='123456'
                  value={field.value}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                  errorMessage={errors.details?.zipCode?.message}
                  required
                />
              )}
            />

            {/* Reason */}
            <Controller
              name='reason'
              control={control}
              rules={{
                required: 'Reason is required'
              }}
              render={({ field }) => (
                <TextField
                  id='reason'
                  label='Reason for this address'
                  placeholder='Card delivery address'
                  value={field.value}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                  errorMessage={errors.reason?.message}
                  required
                />
              )}
            />
          </div>

          <div className='mt-8 flex gap-4'>
            <OutlineButton
              type='button'
              onClick={() => setStep(OrderCardStep.DELIVERY)}
              shrink
            >
              Back
            </OutlineButton>
            <Button type='submit' disabled={!isValid} shrink>
              Continue
            </Button>
          </div>
        </CardContent>
      </Card>
    </Form>
  )
}
