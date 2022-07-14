import type { LoaderFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { mapsClient } from '~/lib/maps.server'

export const loader: LoaderFunction = async ({ request }) => {
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
      console.log(e)
      return e
    })

  const addressComponents = place.address_components

  return json({
    street: { id: placeId, name: place.formatted_address.split(',')[0] },
    apartment: '',
    city: addressComponents.find((item: any) =>
      item.types.includes('administrative_area_level_2')
    ).long_name,
    state: addressComponents.find((item: any) =>
      item.types.includes('administrative_area_level_1')
    ).short_name,
    zip: addressComponents.find((item: any) =>
      item.types.includes('postal_code')
    ).short_name
  })
}
