import { useEffect, useRef, useState } from 'react'

const SCRIPT_ID = 'recaptcha-enterprise-script'
const SCRIPT_SRC =
  'https://www.google.com/recaptcha/enterprise.js?render=explicit'
const G_RECAPTCHA_PRESENT_IN_THE_WINDOW_MAX_POLLING_TIME = 5 * 1000
const G_RECAPTCHA_PRESENT_IN_THE_WINDOW_POLLING_INTERVAL = 100

/**
 * Custom hook to render the Google Recaptcha V2 widget
 * @returns An object with the token (generated after the user solves the captcha) 
 *  and the recaptcha widget to be rendered in the UI
 */
export const useRecaptchaV2 = () => {
  const siteKey =
    (typeof window !== 'undefined' && (window as any).ENV?.recaptchaSiteKey) ||
    ''
  const [token, setToken] = useState('')
  const [isScriptLoaded, setIsScriptLoaded] = useState(false)
  const [error, setError] = useState('')
  const containerRef = useRef<HTMLDivElement | null>(null)
  const widgetIdRef = useRef<string | null>(null)

  useEffect(() => {
    if (!siteKey) return

    if (!document.getElementById(SCRIPT_ID)) {
      const script = document.createElement('script')
      script.src = SCRIPT_SRC
      script.id = SCRIPT_ID
      script.async = true
      script.defer = true

      script.onload = () => setIsScriptLoaded(true)
      script.onerror = () => setError('Failed to load recaptcha.')
      document.body.appendChild(script)
    } else {
      setIsScriptLoaded(true)
    }
  }, [siteKey])

  useEffect(() => {
    const currentContainer = containerRef.current
    if (!siteKey || !isScriptLoaded || !currentContainer) return

    const renderWidget = () => {
      if (!(window as any).grecaptcha || !(window as any).grecaptcha.enterprise)
        return
      ;(window as any).grecaptcha.enterprise.ready(() => {
        if (widgetIdRef.current !== null) {
          // Prevent double rendering
          return
        }

        try {
          if (
            typeof (window as any).grecaptcha.enterprise.render !== 'function'
          ) {
            setError('reCAPTCHA loaded but render method missing.')
            return
          }

          const id = (window as any).grecaptcha.enterprise.render(
            currentContainer,
            {
              sitekey: siteKey,
              callback: (newToken: string) => setToken(newToken),
              'expired-callback': () => setToken(''),
              theme: 'light'
            }
          )
          widgetIdRef.current = id
          setError('') // Clear errors on success
        } catch (e: any) {
          console.error('Recaptcha Render Error:', e)
          if (!e.message.includes('already been rendered')) {
            setError('Error rendering widget. Invalid Site Key?')
          }
        }
      })
    }

    // Poll for the global object if script is loaded but object isn't ready
    const gRecaptchaReadyInterval = setInterval(() => {
      if ((window as any).grecaptcha && (window as any).grecaptcha.enterprise) {
        clearInterval(gRecaptchaReadyInterval)
        renderWidget()
      }
    }, G_RECAPTCHA_PRESENT_IN_THE_WINDOW_POLLING_INTERVAL)
    // Stop polling after X milliseconds
    const gRecaptchaReadyCheckTimeout = setTimeout(
      () => clearInterval(gRecaptchaReadyInterval),
      G_RECAPTCHA_PRESENT_IN_THE_WINDOW_MAX_POLLING_TIME
    )

    return () => {
      clearInterval(gRecaptchaReadyInterval)
      clearTimeout(gRecaptchaReadyCheckTimeout)
      widgetIdRef.current = null
      if (currentContainer) {
        currentContainer.innerHTML = ''
      }
      setToken('')
    }
  }, [siteKey, isScriptLoaded])

  const RecaptchaWidget = siteKey ? (
    <div className='flex min-h-[100px] w-full flex-col items-center justify-center rounded border border-slate-100 bg-slate-50 p-4'>
      {error && (
        <div className='mb-2 text-center text-xs text-red-500'>{error}</div>
      )}

      {/* The actual container for the Google Widget */}
      <div
        ref={containerRef}
        className='min-h-[78px] transition-all duration-500 ease-in-out'
      ></div>

      {/* Loading State */}
      {!isScriptLoaded && !error && (
        <span className='mt-2 animate-pulse text-xs text-gray-400'>
          Loading Recaptcha...
        </span>
      )}
    </div>
  ) : null

  return {
    token,
    RecaptchaWidget
  }
}
