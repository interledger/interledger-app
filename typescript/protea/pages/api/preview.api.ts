import { NextApiRequest, NextApiResponse } from 'next'

/**
 * We hijack the next preview mode to enable test mode.
 * This function enables toggling of the preview mode at /api/preview.
 * @param req NextApiRequest the request body.
 * @param res NextApiResponse the request response.
 */
export default async function preview(
  req: NextApiRequest,
  res: NextApiResponse
) {
  if (req.query.toggle === 'on') {
    res.setPreviewData({})
    res.send(true)
    res.end()
  } else if (req.query.toggle === 'off') {
    res.clearPreviewData()
    res.send(false)
    res.end()
  } else {
    res.json({ preview: req.preview || false })
    res.end()
  }
}
