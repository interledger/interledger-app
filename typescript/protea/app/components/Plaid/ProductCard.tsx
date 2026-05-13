import { Card, CardContent, CardHeader, CardTitle } from '~/components'
import type { PlaidProduct } from '~/lib/usePlaidStore'
import { usePlaidStore } from '~/lib/usePlaidStore'

interface ProductCardProps {
  product: PlaidProduct
}

export function ProductCard({ product }: ProductCardProps) {
  const response = usePlaidStore((s) => s.lastResponses[product])

  if (!response) {
    return null
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="capitalize">{product} Response</CardTitle>
      </CardHeader>
      <CardContent>
        <pre className="overflow-x-auto whitespace-pre-wrap break-all text-xs">
          {JSON.stringify(response, null, 2)}
        </pre>
      </CardContent>
    </Card>
  )
}
