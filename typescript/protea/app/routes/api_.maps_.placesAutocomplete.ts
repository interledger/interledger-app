import { PlaceAutocompleteType } from '@googlemaps/google-maps-services-js'
import type { LoaderFunctionArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { mapsClient } from '~/lib/maps.server'
import { getLogger } from '~/lib/logger.server'

export async function loader({ request }: LoaderFunctionArgs) {
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
      return r.data.predictions.map((place) => ({
        id: place.place_id,
        name: place.description
      }))
    })
    .catch((e) => {
      const logger = getLogger()
      logger.error(
        { error: e instanceof Error ? e.message : String(e) },
        'Google Maps places autocomplete API error'
      )
      return e
    })

  return json({
    predictions
  })
}
