import type { Route } from './+types/api_.maps_.placesAutocomplete'
import { PlaceAutocompleteType } from '@googlemaps/google-maps-services-js'
import { data } from 'react-router';
import { mapsClient } from '~/lib/maps.server'
import logger from '~/lib/logger.server'

export async function loader({ request }: Route.LoaderArgs) {
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
      logger.error(
        e instanceof Error ? e : { error: String(e) },
        'Google Maps places autocomplete API error'
      )
      return e
    })

  return data({
    predictions
  })
}
