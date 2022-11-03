import type { LoaderArgs } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { Button, Layouts } from '~/components'
import { json } from '@remix-run/node'
import clsx from 'clsx'
import { generateQR, qrSvg } from '~/lib/qr.server'
import QrScanner from 'qr-scanner'
import { useEffect } from 'react'

// export async function loader({ request }: LoaderArgs) {
//   const qr = await generateQR('https://fynbos.me/cairin')
//   return json({ qr, svg: qrSvg(qr) })
// }

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  // const { qr, svg } = useLoaderData<typeof loader>()
  useEffect(() => {
    if (typeof document != 'undefined') {
      const videoElem = document.getElementById('qr-video') as HTMLVideoElement
      //
      const qrScanner = new QrScanner(
        videoElem,
        (result) => console.log('decoded qr code:', result),
        {
          highlightScanRegion: true,
          highlightCodeOutline: true
          /* your options or returnDetailedScanResult: true if you're not specifying any other options */
        }
      )
      // window.scanner = qrScanner
      qrScanner.start().then((r) => console.log('started scanning', r))

      QrScanner.hasCamera().then((val) =>
        console.log('QR scanner has camera', val)
      )
    }

    // return qrScanner.destroy()
  }, [])

  return (
    <>
      {/* TODO Make this a dialogue that is full screen.*/}
      <div
        id='video-container'
        className='example-style-2 fixed relative inset-0 flex w-full flex-col overflow-hidden rounded-2xl bg-page p-4 pb-8'
      >
        <video
          id='qr-video'
          className='relative flex w-full flex-col overflow-hidden rounded-lg'
        ></video>
      </div>
      {/*<div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>*/}
      {/*  {qr.map((row, index) => {*/}
      {/*    return (*/}
      {/*      <div key={`row${index}`} className='flex'>*/}
      {/*        {row.map((col, colIndex) => {*/}
      {/*          return (*/}
      {/*            <div*/}
      {/*              key={`row${index}col${colIndex}`}*/}
      {/*              className={clsx(*/}
      {/*                'h-2 w-2',*/}
      {/*                col.fill ? 'bg-slate-900' : 'bg-white',*/}
      {/*                col.shape*/}
      {/*              )}*/}
      {/*            ></div>*/}
      {/*          )*/}
      {/*        })}*/}
      {/*      </div>*/}
      {/*    )*/}
      {/*  })}*/}
      {/*</div>*/}
      {/*<div*/}
      {/*  className='mt-6 flex w-full flex-col rounded-2xl bg-page bg-clip-border p-4 pb-8'*/}
      {/*  dangerouslySetInnerHTML={{ __html: svg }}*/}
      {/*/>*/}
    </>
  )
}
