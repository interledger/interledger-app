import { DateTime } from 'luxon'
import { href, useLoaderData } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Card, CardContent, Icon, Layouts } from '~/components'
import { Label } from '~/components/Label'
import { getKycStatus } from '~/data/wallet.server'
import {
  ErrorHandler,
  ErrorMapper,
  UserFacingError
} from '~/lib/error-handling/bff-error'
import { ServerResponse } from '~/lib/error-handling/types'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import type { Route } from './+types/settings.profile-personal'

export async function loader({
  request
}: Route.LoaderArgs): Promise<ServerResponse> {
  const kycStatus = await getKycStatus(request)

  const kycDetails = await grpc.getIndividualKYC(request, {})
  if (isConnectError(kycDetails)) {
    const userFacingError = ErrorMapper.grpc.toUserFacingError(kycDetails)
    return ErrorHandler(request, userFacingError, {
      cb: () => {
        return {
          success: false,
          error: UserFacingError(
            'Personal information not available, please try again or contact support if the issue persists.'
          )
        }
      }
    })
  }

  let countries = await grpc.getCountries(request, {})
  if (isConnectError(countries)) throw countries

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

  return {
    success: true,
    data: {
      kycStatus,
      kycDetails,
      gender,
      dateOfBirth: DateTime.fromSeconds(
        Number(kycDetails?.dateOfBirth?.seconds)
      ).toFormat('dd MMMM yyyy'),
      country: countries.countries.find(
        (country) => country.id == kycDetails.address?.countryCode
      )?.name
    }
  }
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: href('/settings'),
      title: 'Personal information'
    },
    isNested: true
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Personal information'
  }
])

export default function Page() {
  const loaderData = useLoaderData<typeof loader>()

  if (loaderData.success == false) {
    return (
      <Card>
        <CardContent className='flex flex-col space-y-4'>
          {loaderData.error?.message}
        </CardContent>
      </Card>
    )
  }

  const { dateOfBirth, gender, kycDetails, country } = loaderData.data

  return (
    <Card>
      <CardContent className='flex flex-col space-y-4'>
        <div className='flex w-full flex-col space-y-1'>
          <Label>Legal name</Label>
          <div className='mt-1 flex w-full justify-between p-3'>
            <div className='flex space-x-3'>
              <Icon>account_circle</Icon>
              <span>
                {kycDetails.firstName} {kycDetails.lastName}
              </span>
            </div>
          </div>
        </div>
        <div className='flex w-full flex-col space-y-1'>
          <Label>Address</Label>
          <div className='mt-1 flex w-full justify-between p-3'>
            <div className='flex space-x-3'>
              <Icon>location_on</Icon>
              <span>
                {kycDetails.address?.formattedAddress
                  ?.split(',')
                  .filter((line: string) => line.length > 0)
                  .map((line: string) => (
                    <>
                      <span>{line}</span>
                      <br />
                    </>
                  ))}
              </span>
            </div>
          </div>
        </div>
        <div className='flex w-full flex-col space-y-1'>
          <Label>Country of residence</Label>
          <div className='mt-1 flex w-full justify-between p-3'>
            <div className='flex space-x-3'>
              <Icon>flag</Icon>
              <span>{country}</span>
            </div>
          </div>
        </div>
        {kycDetails.gender > 0 && (
          <div className='flex w-full flex-col space-y-1'>
            <Label>Gender</Label>
            <div className='mt-1 flex w-full justify-between p-3'>
              <div className='flex space-x-3'>
                <Icon>{gender.icon}</Icon>
                <span>{gender.title}</span>
              </div>
            </div>
          </div>
        )}
        <div className='flex w-full flex-col space-y-1'>
          <Label>Birth date</Label>
          <div className='mt-1 flex w-full justify-between p-3'>
            <div className='flex space-x-3'>
              <Icon>calendar_today</Icon>
              <span>{dateOfBirth}</span>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
