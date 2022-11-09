// READ https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For if you want to edit this....
export function getClientIP(request: Request): string {
  const forwardedFor = (request.headers.get('x-forwarded-for') as string).split(
    ','
  )
  return forwardedFor[0]
}
