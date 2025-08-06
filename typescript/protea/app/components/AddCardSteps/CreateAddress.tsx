import { zodResolver } from '@hookform/resolvers/zod'
import { Form } from '@remix-run/react'
import { Controller, useForm } from 'react-hook-form'
import { z } from 'zod'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  OutlineButton,
  Select,
  SelectOptions,
  TextField
} from '~/components'
import {
  CustomerDeliveryAddressType,
  NewCustomerDeliveryAddress
} from '~/generated/connect/backend/v1/backend_pb'
import { AddCardStep, useAddCardStore } from '~/lib/useAddCardStore'

const addressTypeOptions = [
  { id: 'other', name: 'Other' },
  { id: 'permanent', name: 'Permanent Residence' },
  { id: 'temporary', name: 'Temporary Residence' },
  { id: 'work', name: 'Work' }
]

const getAddressTypeValue = (currentType: CustomerDeliveryAddressType) => {
  switch (currentType) {
    case CustomerDeliveryAddressType.CUSTOMER_DELIVERY_ADDRESS_PERMANENT_RESIDENCE:
      return addressTypeOptions[1]
    case CustomerDeliveryAddressType.CUSTOMER_DELIVERY_ADDRESS_TEMPORARY_RESIDENCE:
      return addressTypeOptions[2]
    case CustomerDeliveryAddressType.CUSTOMER_DELIVERY_ADDRESS_WORK:
      return addressTypeOptions[3]
    default:
      return addressTypeOptions[0]
  }
}

const createNewAddress = (
  data: AddressFormData
): NewCustomerDeliveryAddress => {
  return new NewCustomerDeliveryAddress({
    details: {
      type: data.details.type,
      countryCode: data.details.countryCode.toUpperCase(),
      line1: data.details.line1,
      line2: data.details.line2 || undefined,
      line3: data.details.line3 || undefined,
      postOffice: data.details.postOffice || undefined,
      city: data.details.city,
      zipCode: data.details.zipCode
    },
    reason: data.reason
  })
}

const addressFormSchema = z.object({
  details: z.object({
    type: z.nativeEnum(CustomerDeliveryAddressType),
    countryCode: z
      .string()
      .length(2, 'Country code must have 2 characters')
      .regex(/^[A-Za-z]+$/, 'Country code must contain only letters')
      .toUpperCase(),
    line1: z.string().min(1, 'Address line 1 is required'),
    line2: z.string().optional(),
    line3: z.string().optional(),
    postOffice: z.string().optional(),
    city: z.string().min(1, 'City is required'),
    zipCode: z
      .string()
      .regex(/^\d+$/, 'Zip code must contain digits only')
      .length(6, 'Zip code must have 6 digits')
  }),
  reason: z.string().min(1, 'Reason is required')
})
type AddressFormData = z.infer<typeof addressFormSchema>

const defaultValues: AddressFormData = {
  details: {
    type: CustomerDeliveryAddressType.CUSTOMER_DELIVERY_ADDRESS_TYPE_OTHER,
    countryCode: '',
    line1: '',
    line2: '',
    line3: '',
    postOffice: '',
    city: '',
    zipCode: ''
  },
  reason: ''
}

export const CreateAddress = () => {
  const {
    control,
    handleSubmit,
    watch,
    setValue,
    formState: { errors, isValid }
  } = useForm<AddressFormData>({
    defaultValues,
    mode: 'onChange',
    resolver: zodResolver(addressFormSchema)
  })
  const { setNewAddress, setStep } = useAddCardStore()

  const handleAddressTypeChange = (option: SelectOptions) => {
    let type: CustomerDeliveryAddressType
    switch (option.id) {
      case 'permanent':
        type =
          CustomerDeliveryAddressType.CUSTOMER_DELIVERY_ADDRESS_PERMANENT_RESIDENCE
        break
      case 'temporary':
        type =
          CustomerDeliveryAddressType.CUSTOMER_DELIVERY_ADDRESS_TEMPORARY_RESIDENCE
        break
      case 'work':
        type = CustomerDeliveryAddressType.CUSTOMER_DELIVERY_ADDRESS_WORK
        break
      default:
        type = CustomerDeliveryAddressType.CUSTOMER_DELIVERY_ADDRESS_TYPE_OTHER
    }
    setValue('details.type', type)
  }

  const onSubmit = (data: AddressFormData) => {
    const newAddress = createNewAddress(data)
    setNewAddress(newAddress)
    console.log('newAddress', newAddress)
    setStep(AddCardStep.DELIVERY)
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

            {/* Country Code */}
            <Controller
              name='details.countryCode'
              control={control}
              rules={{
                required: 'Country code is required',
                maxLength: {
                  value: 2,
                  message: 'Country code must be 2 characters'
                }
              }}
              render={({ field }) => (
                <TextField
                  id='countryCode'
                  label='Country Code'
                  placeholder='US'
                  value={field.value}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                  errorMessage={errors.details?.countryCode?.message}
                  maxLength={2}
                />
              )}
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
                  value={field.value}
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
                  placeholder='10001'
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
              onClick={() => setStep(AddCardStep.DELIVERY)}
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
