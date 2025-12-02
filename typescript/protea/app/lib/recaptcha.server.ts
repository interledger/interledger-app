const projectId = process.env.RECAPTCHA_PROJECT_ID
const apiKey = process.env.RECAPTCHA_API_KEY
const siteKey = process.env.RECAPTCHA_SITE_KEY

export const verifyRecaptcha = async (token: string): Promise<boolean> => {
  console.log("Starting recaptcha verification")
  
  if (!projectId || !apiKey || !siteKey) {
    console.warn(
      'Recaptcha configuration missing (RECAPTCHA_PROJECT_ID, RECAPTCHA_API_KEY, RECAPTCHA_SITE_KEY). Skipping verification.'
    )
    return false
  }

  try {
    const response = await fetch(
      `https://recaptchaenterprise.googleapis.com/v1/projects/${projectId}/assessments?key=${apiKey}`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          event: {
            token,
            siteKey
          }
        })
      }
    )

    const data = await response.json()

    if (!data.tokenProperties?.valid) {
      console.error('Recaptcha verification failed:', data)
      return false
    }
    console.log('Recaptcha verification passed:', data)

    return true
  } catch (error) {
    console.error('Recaptcha verification error:', error)
    return false
  }
}
