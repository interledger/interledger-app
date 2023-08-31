import DOMPurify from 'dompurify'
import { JSDOM } from 'jsdom'

export async function fetchAndSanitizeHTML(url: string): Promise<string> {
  const response = await fetch(url, {
    headers: {
      'content-type': 'text/html; charset=UTF-8'
    }
  })

  const rawHTML = await response.text()

  return sanitizeHTML(rawHTML)
}

export function sanitizeHTML(rawHTML: string): string {
  const parser = new JSDOM(rawHTML)

  const purify = DOMPurify(parser.window)
  // Sanitize the HTML content using DOMPurify
  return purify.sanitize(parser.window.document.body.innerHTML)
}
