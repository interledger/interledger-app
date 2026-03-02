import type { HTMLAttributes } from 'react'
import { forwardRef, useEffect, useImperativeHandle, useRef } from 'react'
import { useScript } from '~/lib/useScript'

const Turnstile = forwardRef<TurnstileInstance | undefined, TurnstileProps>(
  ({ ...props }, ref) => {
    const { siteKey, appearance } = props

    const scriptStatus = useScript(
      'https://challenges.cloudflare.com/turnstile/v0/api.js'
    )
    const turnstileRef = useRef<HTMLDivElement | null>(null)
    const firstRendered = useRef(false)

    useImperativeHandle(
      ref,
      () => {
        if (typeof window === 'undefined' || !scriptStatus) {
          return
        }

        const { turnstile } = window
        return {
          reset() {
            if (!turnstile?.reset) {
              console.warn('Turnstile has not been loaded')
              return
            }

            try {
              turnstile.reset()
            } catch (error) {
              console.warn(`Failed to reset Turnstile widget`, error)
            }
          }
        }
      },
      [scriptStatus]
    )

    useEffect(() => {
      if (!siteKey) {
        console.warn('sitekey was not provided')
        return
      }

      if (
        scriptStatus !== 'ready' ||
        !turnstileRef.current ||
        firstRendered.current
      ) {
        return
      }

      window.turnstile?.render(turnstileRef.current!, {
        sitekey: siteKey,
        callback: (token: string) => {
          if (props.onSuccess) {
            props.onSuccess(token)
          }
        },
        'before-interactive-callback': () => {
          if (props.beforeInteractive) {
            props.beforeInteractive()
          }
        }
      })

      firstRendered.current = true
    }, [siteKey, props, scriptStatus, turnstileRef, firstRendered])

    return (
      <div
        ref={turnstileRef}
        className='cf-turnstile'
        data-sitekey={siteKey}
        data-appearance={appearance}
      />
    )
  }
)

interface TurnstileProps extends HTMLAttributes<HTMLDivElement> {
  siteKey: string
  appearance?: TurnstileAppearance
  onSuccess?: (token: string) => void
  beforeInteractive?: () => void
}

export interface TurnstileInstance {
  reset: () => void
}

export type TurnstileAppearance = 'interaction-only' | 'execute' | 'always'

Turnstile.displayName = 'Turnstile'

export { Turnstile }
