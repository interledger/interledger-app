import { Button } from '~/components/Buttons'

import { usePlaidLinkFlow } from './usePlaidLinkFlow'

export function PlaidLinkButton() {
  const { connect, busy, scriptError } = usePlaidLinkFlow()

  // NOTE: react-plaid-link's `ready` flag is bound to `options.token`. With
  // no token yet (initial render) `ready` is permanently false, so it can't
  // gate the button. Click-readiness comes from the SDK script being present
  // — failure of which surfaces as `scriptError`. The token round-trip + the
  // hook's pendingOpenRef effect take care of opening once both prerequisites
  // (token + ready) line up.
  const label = scriptError
    ? 'Plaid SDK failed to load'
    : busy
      ? 'Connecting…'
      : 'Connect a bank'

  return (
    <Button
      type='button'
      onClick={connect}
      disabled={busy || !!scriptError}
    >
      {label}
    </Button>
  )
}
