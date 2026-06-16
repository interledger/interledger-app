import { Form, useNavigation } from 'react-router'
import { OutlineButton } from '~/components'
import type { PlaidProduct } from '~/lib/usePlaidStore'

interface EndpointButtonProps {
  product: PlaidProduct
  disabled?: boolean
}

export function EndpointButton({ product, disabled }: EndpointButtonProps) {
  const navigation = useNavigation()
  const isSubmitting =
    navigation.state === 'submitting' &&
    navigation.formData?.get('intent') === 'fetch_product' &&
    navigation.formData?.get('product') === product

  return (
    <Form method='post'>
      <input type='hidden' name='intent' value='fetch_product' />
      <input type='hidden' name='product' value={product} />
      <OutlineButton type='submit' disabled={disabled || isSubmitting} className="capitalize">
        {isSubmitting ? '...' : product}
      </OutlineButton>
    </Form>
  )
}
