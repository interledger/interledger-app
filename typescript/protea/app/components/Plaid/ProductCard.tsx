import { useState } from 'react'

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  TextButton
} from '~/components'
import type { PlaidProduct } from '~/lib/usePlaidStore'
import { usePlaidStore } from '~/lib/usePlaidStore'

interface ProductCardProps {
  product: PlaidProduct
}

export function ProductCard({ product }: ProductCardProps) {
  const response = usePlaidStore((s) => s.lastResponses[product])
  const setActiveProduct = usePlaidStore((s) => s.setActiveProduct)
  const [showJson, setShowJson] = useState(false)

  if (!response) {
    return null
  }

  return (
    <Card>
      <CardHeader>
        <div className='flex items-center justify-between'>
          <CardTitle className='capitalize'>{product} Response</CardTitle>
          <div className='flex items-center gap-3'>
            <TextButton
              onClick={() => setShowJson((s) => !s)}
              aria-expanded={showJson}
            >
              {showJson ? 'Hide JSON' : 'Show JSON'}
            </TextButton>
            <TextButton
              onClick={() => setActiveProduct(null)}
              aria-label={`Clear ${product} response`}
            >
              Clear
            </TextButton>
          </div>
        </div>
      </CardHeader>
      {showJson && (
        <CardContent>
          <pre className='overflow-x-auto whitespace-pre-wrap break-all text-xs'>
            {JSON.stringify(response, null, 2)}
          </pre>
        </CardContent>
      )}
    </Card>
  )
}
