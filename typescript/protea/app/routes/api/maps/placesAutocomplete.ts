import type { PlaceAutocompleteResult } from '@googlemaps/google-maps-services-js'
import { PlaceAutocompleteType } from '@googlemaps/google-maps-services-js'
import type { LoaderFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { mapsClient } from '~/lib/maps.server'

export const loader: LoaderFunction = async ({ request }) => {
  const url = new URL(request.url)
  const query = url.searchParams.get('query') || ''
  const country = url.searchParams.get('country') || ''

  const predictions = await mapsClient
    .placeAutocomplete({
      params: {
        input: query,
        types: PlaceAutocompleteType.address,
        components: [`country:${country}`],
        key: process.env.GOOGLE_MAPS_API_KEY || ''
      },
      signal: request.signal
    })
    .then((r) => {
      return r.data.predictions.map((place: PlaceAutocompleteResult) => ({
        id: place.place_id,
        name: place.description
      }))
    })
    .catch((e) => {
      console.log(e)
      return e
    })

  return json({
    predictions
  })
}
