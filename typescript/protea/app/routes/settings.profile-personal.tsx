import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { DateTime } from 'luxon'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Card, Icon, Layouts } from '~/components'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { getKycDetails, getKycStatus } from '~/lib/wallet.server'

export async function loader({ request }: LoaderArgs) {
  const kycStatus = await getKycStatus(request)
  const kycDetails = await getKycDetails(request)

  let countries = await grpcClient
    .getCountries({})
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(countries)) {
    throw json({}, httpMapping(countries.code))
  }

  let gender = {
    icon: '',
    title: ''
  }
  switch (kycDetails.gender) {
    case 0:
      gender.title = 'Unknown'
      break
    case 1:
      gender.icon = 'man'
      gender.title = 'Male'
      break
    case 2:
      gender.icon = 'woman'
      gender.title = 'Female'
      break
    case 3:
      gender.title = 'Other'
      break
  }

  return json({
    kycStatus,
    kycDetails,
    gender,
    dateOfBirth: DateTime.fromSeconds(
      parseInt(kycDetails?.dateOfBirth?.seconds as string)
    ).toFormat('dd MMMM yyyy'),
    country: countries.response.countries.find(
      (country) => country.id == kycDetails.address?.countryCode
    )?.name
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: route('/settings'),
      title: 'Personal information'
    },
    isNested: true
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Personal information'
  }
}

export default function Page() {
  const { dateOfBirth, gender, kycDetails, country } =
    useLoaderData<typeof loader>()
  return (
    <Card>
      <h2 className='mt-6 text-sm font-medium'>Legal name</h2>
      <div className='mt-2 flex items-center justify-start rounded-xl bg-nav p-3 text-medium'>
        <div className='flex space-x-3'>
          <Icon>face</Icon>
          <span>
            {kycDetails.firstName} {kycDetails.lastName}
          </span>
        </div>
      </div>
      <h2 className='mt-6 text-sm font-medium'>Address</h2>
      <div className='mt-2 flex items-center justify-start rounded-xl bg-nav p-3 text-medium'>
        <div className='flex space-x-3'>
          <Icon>location_on</Icon>
          <span>
            {kycDetails.address?.formattedAddress?.split(',').map((line) => (
              <>
                <span>{line}</span>
                <br />
              </>
            ))}
          </span>
        </div>
      </div>
      <h2 className='mt-6 text-sm font-medium'>Country of residence</h2>
      <div className='mt-2 flex items-center justify-start rounded-xl bg-nav p-3 text-medium'>
        <div className='flex space-x-3'>
          <Icon>flag</Icon>
          <span>{country}</span>
          <span>{kycDetails.countryCode}</span>
        </div>
      </div>
      <h2 className='mt-6 text-sm font-medium'>Gender</h2>
      <div className='mt-2 flex items-center justify-start rounded-xl bg-nav p-3 text-medium'>
        <div className='flex space-x-3'>
          <Icon>{gender.icon}</Icon>
          <span>{gender.title}</span>
        </div>
      </div>
      <h2 className='mt-6 text-sm font-medium'>Birth date</h2>
      <div className='mt-2 flex items-center justify-start rounded-xl bg-nav p-3 text-medium'>
        <div className='flex space-x-3'>
          <Icon>calendar_today</Icon>
          <span>{dateOfBirth}</span>
        </div>
      </div>
    </Card>
  )
}
