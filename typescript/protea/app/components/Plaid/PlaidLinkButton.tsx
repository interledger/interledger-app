import { Button } from '~/components/Buttons'

import { usePlaidLinkFlow } from './usePlaidLinkFlow'

/**
 * PlaidLinkButton wires the Connect-a-bank flow into a single protea Button.
 *
 * The button:
 *   - is disabled while either action is in flight or the modal is open;
 *   - delegates everything to `usePlaidLinkFlow`, which mints the link_token,
 *     opens the Plaid iframe, and posts back the public_token via the /plaid
 *     action.
 *
 * Container style/layout is left to the route — this component renders just
 * the button.
 */
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
