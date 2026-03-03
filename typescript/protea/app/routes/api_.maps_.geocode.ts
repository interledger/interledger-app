import type { Route } from './+types/api_.maps_.geocode'
import type { AddressComponent } from '@googlemaps/google-maps-services-js'
import { AddressType } from '@googlemaps/google-maps-services-js'
import { data } from 'react-router';
import { mapsClient } from '~/lib/maps.server'
import logger from '~/lib/logger.server'

export async function loader({ request }: Route.LoaderArgs) {
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
      logger.error(
        { error: e instanceof Error ? e.message : String(e) },
        'Google Maps geocode API error'
      )
      return e
    })

  const addressComponents: AddressComponent[] = place.address_components

  return data({
    street: { id: placeId, name: place.formatted_address },
    line1: `${
      addressComponents.find((item) =>
        item.types.includes(AddressType.street_number)
      )?.long_name
    } ${
      addressComponents.find((item) => item.types.includes(AddressType.route))
        ?.long_name
    }`,
    line2: addressComponents.find((item) =>
      item.types.includes(AddressType.locality)
    )?.long_name,
    building: '',
    apartment: '',
    city: addressComponents.find((item) =>
      item.types.includes(AddressType.administrative_area_level_2)
    )?.long_name,
    state: `${
      addressComponents.find((item) => item.types.includes(AddressType.country))
        ?.short_name
    }-${
      addressComponents.find((item) =>
        item.types.includes(AddressType.administrative_area_level_1)
      )?.short_name
    }`,
    zipCode: addressComponents.find((item) =>
      item.types.includes(AddressType.postal_code)
    )?.short_name,
    countryCode: addressComponents.find((item) =>
      item.types.includes(AddressType.country)
    )?.short_name,
    placeID: placeId,
    formattedAddress: place.formatted_address
  })
}
