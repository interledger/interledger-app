import type { AddressComponent } from '@googlemaps/google-maps-services-js'
import type { LoaderFunctionArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { mapsClient } from '~/lib/maps.server'
import { getLogger } from '~/lib/logger.server'

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  const placeId = url.searchParams.get('place-id') || ''

  const place = await mapsClient
    .geocode({
      params: {
        place_id: placeId,
        key: process.env.GOOGLE_MAPS_API_KEY || ''
      }
    })
    .then((r) => {
      //Generally, only one entry in the "results" array is returned for address lookups
      return r.data.results[0]
    })
    .catch((e) => {
      const logger = getLogger()
      logger.error(
        { error: e instanceof Error ? e.message : String(e) },
        'Google Maps geocode API error'
      )
      return e
    })

  const addressComponents: AddressComponent[] = place.address_components

  return json({
    street: { id: placeId, name: place.formatted_address },
    line1: `${
      addressComponents.find((item: any) =>
        item.types.includes('street_number')
      )?.long_name
    } ${
      addressComponents.find((item: any) => item.types.includes('route'))
        ?.long_name
    }`,
    line2: addressComponents.find((item: any) =>
      item.types.includes('locality')
    )?.long_name,
    building: '',
    apartment: '',
    city: addressComponents.find((item: any) =>
      item.types.includes('administrative_area_level_2')
    )?.long_name,
    state: `${
      addressComponents.find((item: any) => item.types.includes('country'))
        ?.short_name
    }-${
      addressComponents.find((item: any) =>
        item.types.includes('administrative_area_level_1')
      )?.short_name
    }`,
    zipCode: addressComponents.find((item: any) =>
      item.types.includes('postal_code')
    )?.short_name,
    countryCode: addressComponents.find((item: any) =>
      item.types.includes('country')
    )?.short_name,
    placeID: placeId,
    formattedAddress: place.formatted_address
  })
}
