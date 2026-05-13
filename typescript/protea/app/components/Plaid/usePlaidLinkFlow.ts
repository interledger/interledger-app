import { useCallback, useEffect, useRef, useState } from 'react'
import { useFetcher } from 'react-router'
import { usePlaidLink } from 'react-plaid-link'
import { v4 } from 'uuid'

import { useScaffoldStore } from '~/lib/useScaffoldStore'
import { usePlaidStore } from '~/lib/usePlaidStore'

import type { ActionData } from '~/routes/plaid'

const PLAID_ACTION_PATH = '/plaid'

// usePlaidLinkFlow orchestrates the Plaid Link round-trip:
//   1. user clicks Connect → POST {intent: create_link_token} to /plaid action
//   2. action returns a link_token; the hook stashes it in the store
//   3. usePlaidLink picks up the token and (once the iframe is ready) opens
//      the Plaid Link modal
//   4. user finishes Link → react-plaid-link fires onSuccess(public_token, metadata)
//   5. hook POSTs {intent: exchange, public_token, account_id} to /plaid action
//   6. action exchanges + persists; React Router auto-revalidates the loader;
//      the /plaid route's useEffect flips the store into the "linked" state.
//
// On any exit or error path the store's `isLinking` flag is cleared and
// `lastError` is populated for the snackbar/debug card (F6 will surface).
//
// See documentation/poc/plaid/plaid-link-explained.md for the full SDK
// architecture (iframe, postMessage, CSP).
export function usePlaidLinkFlow() {
  const linkFetcher = useFetcher<ActionData>()
  const exchangeFetcher = useFetcher<ActionData>()

  const linkToken = usePlaidStore((s) => s.linkToken)
  const isLinking = usePlaidStore((s) => s.isLinking)
  const setLinkToken = usePlaidStore((s) => s.setLinkToken)
  const setIsLinking = usePlaidStore((s) => s.setIsLinking)
  const setLastError = usePlaidStore((s) => s.setLastError)
  const pushSnackbar = useScaffoldStore((s) => s.pushSnackbar)

  // Tracks the user's intent to open Link after the token round-trips. We
  // can't open() synchronously inside the click handler because the token is
  // minted on the server first.
  const pendingOpenRef = useRef(false)

  // True only after the Plaid iframe for the *current* token has fired onLoad.
  // Resets to false on every connect() and onExit so a stale factory can never
  // satisfy the trigger effect (fixes second-click no-op: see bug analysis).
  const [plaidInstanceReady, setPlaidInstanceReady] = useState(false)

  // (1) create_link_token result → store the token (or surface the error).
  useEffect(() => {
    const data = linkFetcher.data
    if (!data) return
    if (data.ok && data.intent === 'create_link_token') {
      setLinkToken(data.linkToken)
      return
    }
    if (!data.ok) {
      setLastError(data.message)
      setIsLinking(false)
      pendingOpenRef.current = false
    }
  }, [linkFetcher.data, setLinkToken, setLastError, setIsLinking])

  // Plaid Link binding (see plaid-link-explained.md).
  const { open, ready, error: scriptError } = usePlaidLink({
    token: linkToken,
    onLoad: () => setPlaidInstanceReady(true),
    onSuccess: (publicToken, metadata) => {
      console.log('Plaid Link onSuccess', publicToken, metadata)
      const accountId = metadata.accounts[0]?.id ?? ''
      const fd = new FormData()
      fd.append('intent', 'exchange')
      fd.append('public_token', publicToken)
      fd.append('account_id', accountId)
      exchangeFetcher.submit(fd, { method: 'POST', action: PLAID_ACTION_PATH })
    },
    onExit: (err) => {
      setPlaidInstanceReady(false)
      setIsLinking(false)
      pendingOpenRef.current = false
      // User cancelled (err === null) — clear the staged token so the next
      // click mints a fresh one (link_tokens are single-flow).
      setLinkToken(null)
      if (err) {
        const msg = err.display_message || err.error_message || 'Plaid Link exited with error'
        setLastError(msg)
        pushSnackbar({ id: v4(), message: msg })
      }
    }
  })

  // (3) When we have a token AND this factory's iframe has loaded AND the user
  //     has opted to open (click), launch the modal exactly once per pending flag.
  //     Uses plaidInstanceReady (set via onLoad) rather than usePlaidLink's `ready`
  //     because `ready` never resets between sessions — onLoad fires per-factory
  //     and only after setPlaid(next) is committed, so `open` is always the live
  //     factory's function by the time this effect can proceed.
  useEffect(() => {
    if (!pendingOpenRef.current) return
    if (!linkToken) return
    if (!plaidInstanceReady) return
    pendingOpenRef.current = false
    open()
  }, [linkToken, plaidInstanceReady, open])

  // (5) exchange result → React Router will revalidate the loader on its own,
  //     which kicks the /plaid useEffect → setLinked. Here we just clean up.
  useEffect(() => {
    const data = exchangeFetcher.data
    if (!data) return
    if (data.ok && data.intent === 'exchange') {
      setLinkToken(null)
      setIsLinking(false)
      return
    }
    if (!data.ok) {
      setLastError(data.message)
      setIsLinking(false)
    }
  }, [exchangeFetcher.data, setLinkToken, setIsLinking, setLastError])

  // Surface SDK-level load failures (CDN unreachable, blocked by extension,
  // …) — see plaid-link-explained.md §10.
  useEffect(() => {
    if (scriptError) {
      const msg = `Plaid SDK failed to load: ${scriptError.message ?? 'unknown error'}`
      setLastError(msg)
      setIsLinking(false)
      pendingOpenRef.current = false
      pushSnackbar({ id: v4(), message: msg })
    }
  }, [scriptError, setLastError, setIsLinking, pushSnackbar])

  const connect = useCallback(() => {
    setPlaidInstanceReady(false)
    setIsLinking(true)
    setLastError(null)
    pendingOpenRef.current = true
    const fd = new FormData()
    fd.append('intent', 'create_link_token')
    linkFetcher.submit(fd, { method: 'POST', action: PLAID_ACTION_PATH })
  }, [linkFetcher, setIsLinking, setLastError])

  const submitting =
    linkFetcher.state !== 'idle' || exchangeFetcher.state !== 'idle'

  return {
    connect,
    /** True while either action is in flight OR the modal is open. */
    busy: isLinking || submitting,
    /** Plaid SDK is loaded and ready to .open(). */
    ready,
    /** SDK-load error, if any. */
    scriptError
  }
}
