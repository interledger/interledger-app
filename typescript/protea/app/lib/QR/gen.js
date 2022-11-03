// The field names given by the backend for field violations
const fs = require('fs')
const QRCode = require('qrcode')

const svgShapes = {
  'rounded-none': (x, y) =>
    `<rect x="${x}" y="${y}"  width="40" height="40" fill="black" stroke="black"/>`,
  'rounded-full': (x, y) =>
    `<rect x="${x}" y="${y}" width="40" height="40" rx="20" fill="black" stroke="black"/>`,
  'rounded-t-full': (x, y) =>
    `<path d="M${x} ${y}m0 20a20 20 -180 0 1 40 0v20h-40v-20Z" fill="black" stroke="black" />`,
  'rounded-tl-full': (x, y) =>
    `<path fill="black" d="M${x} ${y}m40 0v40h-40a40 40 -90 0 1 40 -40Z" stroke="black" />`,
  'rounded-tr-full': (x, y) =>
    `<path fill="black" d="M${x} ${y}a40 40 -90 0 1 40 40h-40v-40Z" stroke="black" />`,
  'rounded-b-full': (x, y) =>
    `<path fill="black" d="M${x} ${y}h40v20a20 20 -180 0 1 -40 0v-20Z" stroke="black" />`,
  'rounded-bl-full': (x, y) =>
    `<path fill="black" d="M${x} ${y}h40v40a40 40 -90 0 1 -40 -40Z" stroke="black" />`,
  'rounded-br-full': (x, y) =>
    `<path fill="black" d="M${x} ${y}h40a40 40 90 0 1 -40 40v-40Z" stroke="black" />`,
  'rounded-l-full': (x, y) =>
    `<path fill="black" d="M${x} ${y}m40 0v40h-20a20 20 -180 0 1 0 -40h20Z" stroke="black" />`,
  'rounded-r-full': (x, y) =>
    `<path fill="black" d="M${x} ${y}h20a20 20 -180 0 1 0 40h-20v-40Z" stroke="black" />`
}

async function svgToFile(text) {
  const path = `${__dirname}/${text.replace('https://fynbos.me/mug/', '')}.svg`
  const qr = await generateQR(text)
  const svg = qrSvg(qr)

  fs.writeFile(path, svg, function (err) {
    if (err) {
      console.log('File write error', err)
      return
    }
    console.log('The file was saved!')
  })
}

function qrSvg(qr) {
  const blockSize = 40
  let svg = `<svg width="${qr.length * blockSize}" height="${
    qr.length * blockSize
  }" viewBox="0 0 ${qr.length * blockSize} ${
    qr.length * blockSize
  }" fill="none" xmlns="http://www.w3.org/2000/svg">`

  for (let i = 0; i < qr.length; i++) {
    for (let j = 0; j < qr.length; j++) {
      if (qr[i][j].fill)
        svg += svgShapes[qr[i][j].shape](j * blockSize, i * blockSize)
    }
  }

  svg += `</svg>`
  // console.log(svg)
  return svg
}

async function generateQR(text) {
  const qr = await QRCode.create(text, {
    errorCorrectionLevel: 'H'
  })

  // Set Defaults
  let qrArr = []
  let eyeSize = 0

  for (let i = 0; i < qr.modules.size / 2; i++) {
    if (eyeSize == 0) {
      let current = qr.modules.get(0, i)
      let next = qr.modules.get(0, i + 1)
      if (current != next) eyeSize = i + 1
    }
  }
  console.log('eyeSize', eyeSize, qr.modules.size)

  for (let i = 0; i < qr.modules.size; i++) {
    qrArr.push([])
    for (let j = 0; j < qr.modules.size; j++) {
      let shape = 'rounded-none'
      // Top left eye
      if (i <= eyeSize && j <= eyeSize) {
        shape = 'rounded-none'
      }
      // Bottom left eye
      else if (i >= qr.modules.size - 1 - eyeSize && j <= eyeSize) {
        shape = 'rounded-none'
      }
      // Top right eye
      else if (i <= eyeSize && j >= qr.modules.size - 1 - eyeSize) {
        shape = 'rounded-none'
      } else {
        // BitMatrix.prototype.get = function (row, col) {
        //   return this.data[row * this.size + col]
        // }
        let topAdj = qr.modules.get(i - 1, j) ?? 0
        let leftAdj = j == 0 ? 0 : qr.modules.get(i, j - 1) ?? 0
        let rightAdj =
          j == qr.modules.size - 1 ? 0 : qr.modules.get(i, j + 1) ?? 0
        let bottomAdj = qr.modules.get(i + 1, j) ?? 0

        if (!topAdj && !leftAdj && !rightAdj && !bottomAdj) {
          shape = 'rounded-full'
        } else if (!topAdj && !leftAdj && !rightAdj && bottomAdj) {
          shape = 'rounded-t-full'
        } else if (topAdj && !leftAdj && !rightAdj && !bottomAdj) {
          shape = 'rounded-b-full'
        } else if (!topAdj && !leftAdj && rightAdj && !bottomAdj) {
          shape = 'rounded-l-full'
        } else if (!topAdj && leftAdj && !rightAdj && !bottomAdj) {
          shape = 'rounded-r-full'
        } else if (!topAdj && !leftAdj && rightAdj && bottomAdj) {
          shape = 'rounded-tl-full'
        } else if (!topAdj && leftAdj && !rightAdj && bottomAdj) {
          shape = 'rounded-tr-full'
        } else if (topAdj && !leftAdj && rightAdj && !bottomAdj) {
          shape = 'rounded-bl-full'
        } else if (topAdj && leftAdj && !rightAdj && !bottomAdj) {
          shape = 'rounded-br-full'
        } else shape = 'rounded-none'
      }

      qrArr[i][j] = {
        fill: qr.modules.get(i, j) > 0,
        shape: shape
      }
    }
  }
  return qrArr
}

let codes = [
  'https://fynbos.me/mug/efd9167e',
  'https://fynbos.me/mug/df3ab8ae',
  'https://fynbos.me/mug/2e0fffb3',
  'https://fynbos.me/mug/1e25f533',
  'https://fynbos.me/mug/dbaadcc6',
  'https://fynbos.me/mug/bbd6184a',
  'https://fynbos.me/mug/39426456',
  'https://fynbos.me/mug/9d4e273c',
  'https://fynbos.me/mug/84198b8a',
  'https://fynbos.me/mug/32770f71',
  'https://fynbos.me/mug/2d42b21a',
  'https://fynbos.me/mug/19820eb2',
  'https://fynbos.me/mug/625ee641',
  'https://fynbos.me/mug/d79b29ee',
  'https://fynbos.me/mug/cd14cd50',
  'https://fynbos.me/mug/472ee230',
  'https://fynbos.me/mug/6430ed30',
  'https://fynbos.me/mug/7a9487ef',
  'https://fynbos.me/mug/40395f07',
  'https://fynbos.me/mug/a320ce5e',
  'https://fynbos.me/mug/295b30e0',
  'https://fynbos.me/mug/16ba8774',
  'https://fynbos.me/mug/1637f1c2',
  'https://fynbos.me/mug/fb45a7ab',
  'https://fynbos.me/mug/e67e38ad',
  'https://fynbos.me/mug/fdd6fae6',
  'https://fynbos.me/mug/21a01e12',
  'https://fynbos.me/mug/fde52908',
  'https://fynbos.me/mug/85a6f613',
  'https://fynbos.me/mug/d0d1f2d2',
  'https://fynbos.me/mug/5c53c77b',
  'https://fynbos.me/mug/7270c732',
  'https://fynbos.me/mug/41b048f3',
  'https://fynbos.me/mug/b9de173d',
  'https://fynbos.me/mug/af2c677a',
  'https://fynbos.me/mug/414d0dfc',
  'https://fynbos.me/mug/e5a09489',
  'https://fynbos.me/mug/c1a78ccf',
  'https://fynbos.me/mug/e74d6bf3',
  'https://fynbos.me/mug/fa36933c',
  'https://fynbos.me/mug/0d7564ba',
  'https://fynbos.me/mug/07477201',
  'https://fynbos.me/mug/5bb5a2ba',
  'https://fynbos.me/mug/1b24760d',
  'https://fynbos.me/mug/420f1f9a',
  'https://fynbos.me/mug/18fde3c8',
  'https://fynbos.me/mug/c5e1b5d7',
  'https://fynbos.me/mug/f3ca7010',
  'https://fynbos.me/mug/8e2ece50',
  'https://fynbos.me/mug/a9f4978d',
  'https://fynbos.me/mug/9891b8be',
  'https://fynbos.me/mug/a41c754e',
  'https://fynbos.me/mug/36b7edf2',
  'https://fynbos.me/mug/678c03cc',
  'https://fynbos.me/mug/b003c3d1',
  'https://fynbos.me/mug/5e75e4c2',
  'https://fynbos.me/mug/94a85558',
  'https://fynbos.me/mug/bef51925',
  'https://fynbos.me/mug/8f35a224',
  'https://fynbos.me/mug/c846498a',
  'https://fynbos.me/mug/ab782aba',
  'https://fynbos.me/mug/fa198894',
  'https://fynbos.me/mug/7dff8ce6',
  'https://fynbos.me/mug/0e65aa50',
  'https://fynbos.me/mug/fad3e55f',
  'https://fynbos.me/mug/39149ec9',
  'https://fynbos.me/mug/c419f818',
  'https://fynbos.me/mug/04726b35',
  'https://fynbos.me/mug/36e2656b',
  'https://fynbos.me/mug/73beaaf9',
  'https://fynbos.me/mug/ed606fe8',
  'https://fynbos.me/mug/70dabbf3',
  'https://fynbos.me/mug/068915da',
  'https://fynbos.me/mug/dd0cf703',
  'https://fynbos.me/mug/7be383e9',
  'https://fynbos.me/mug/79739bf7',
  'https://fynbos.me/mug/a3602238',
  'https://fynbos.me/mug/b8485218',
  'https://fynbos.me/mug/ac9bc42f',
  'https://fynbos.me/mug/febf673d',
  'https://fynbos.me/mug/3b48ba92',
  'https://fynbos.me/mug/ee5c268c',
  'https://fynbos.me/mug/fbcfceb1',
  'https://fynbos.me/mug/9645f882',
  'https://fynbos.me/mug/3fd768ff',
  'https://fynbos.me/mug/77c02429',
  'https://fynbos.me/mug/e67f69be',
  'https://fynbos.me/mug/bea05b55',
  'https://fynbos.me/mug/640473f2',
  'https://fynbos.me/mug/fdbe77e6',
  'https://fynbos.me/mug/fe65b21f',
  'https://fynbos.me/mug/da1a08a3',
  'https://fynbos.me/mug/e4d4559d',
  'https://fynbos.me/mug/ec1894c2',
  'https://fynbos.me/mug/c6645796',
  'https://fynbos.me/mug/7709bae3',
  'https://fynbos.me/mug/ddea3e2b',
  'https://fynbos.me/mug/8a4c8c2e',
  'https://fynbos.me/mug/f569e896',
  'https://fynbos.me/mug/6f43cca4',
  'https://fynbos.me/mug/d736a93d',
  'https://fynbos.me/mug/99fcf650',
  'https://fynbos.me/mug/ae252e8f',
  'https://fynbos.me/mug/724829fa',
  'https://fynbos.me/mug/6d340223',
  'https://fynbos.me/mug/663d6adf',
  'https://fynbos.me/mug/f2179c0d',
  'https://fynbos.me/mug/a00dd012',
  'https://fynbos.me/mug/f1635847',
  'https://fynbos.me/mug/25c5893c',
  'https://fynbos.me/mug/16d8c6d5',
  'https://fynbos.me/mug/f48ddeeb',
  'https://fynbos.me/mug/24fb245c',
  'https://fynbos.me/mug/104c9174',
  'https://fynbos.me/mug/4c9cf70a',
  'https://fynbos.me/mug/6670389a',
  'https://fynbos.me/mug/bbfa1bc1',
  'https://fynbos.me/mug/8d8f8bf8',
  'https://fynbos.me/mug/5550c135',
  'https://fynbos.me/mug/c57bb2a1',
  'https://fynbos.me/mug/2b101667',
  'https://fynbos.me/mug/72c4036f',
  'https://fynbos.me/mug/63dbb47f',
  'https://fynbos.me/mug/97abd62e',
  'https://fynbos.me/mug/5f0f0151',
  'https://fynbos.me/mug/59794bc9',
  'https://fynbos.me/mug/bb5c15c6',
  'https://fynbos.me/mug/6e0e306d',
  'https://fynbos.me/mug/60d5e55a',
  'https://fynbos.me/mug/b92a5b5d',
  'https://fynbos.me/mug/0e6dae85',
  'https://fynbos.me/mug/6cc932c8',
  'https://fynbos.me/mug/94e7fa20',
  'https://fynbos.me/mug/930233a4',
  'https://fynbos.me/mug/cecda6b0',
  'https://fynbos.me/mug/0beea4e2',
  'https://fynbos.me/mug/3a21cb1d',
  'https://fynbos.me/mug/2d94d949',
  'https://fynbos.me/mug/95ed7199',
  'https://fynbos.me/mug/68fba320',
  'https://fynbos.me/mug/073e6c2d',
  'https://fynbos.me/mug/948edd1f',
  'https://fynbos.me/mug/8f46cc4b',
  'https://fynbos.me/mug/db28301f',
  'https://fynbos.me/mug/4b1ebd8c',
  'https://fynbos.me/mug/4e9d73c6',
  'https://fynbos.me/mug/5582c841',
  'https://fynbos.me/mug/dee85543',
  'https://fynbos.me/mug/c004c15c',
  'https://fynbos.me/mug/02111fd3',
  'https://fynbos.me/mug/03c5e8e1',
  'https://fynbos.me/mug/d740d341',
  'https://fynbos.me/mug/5634fee5',
  'https://fynbos.me/mug/8efce81c',
  'https://fynbos.me/mug/54b34eca',
  'https://fynbos.me/mug/95d16283',
  'https://fynbos.me/mug/96aff7d3',
  'https://fynbos.me/mug/7bd421f1',
  'https://fynbos.me/mug/88f8f5c5',
  'https://fynbos.me/mug/14b68336',
  'https://fynbos.me/mug/e2f5ad76',
  'https://fynbos.me/mug/6fc8c8b5',
  'https://fynbos.me/mug/777700fa',
  'https://fynbos.me/mug/4fa4de0e',
  'https://fynbos.me/mug/c30532b1',
  'https://fynbos.me/mug/dc6c672d',
  'https://fynbos.me/mug/9954159b',
  'https://fynbos.me/mug/b3e03550',
  'https://fynbos.me/mug/a70a5381',
  'https://fynbos.me/mug/e6558f2d',
  'https://fynbos.me/mug/1c90cabc',
  'https://fynbos.me/mug/30efe506',
  'https://fynbos.me/mug/fe6b481f',
  'https://fynbos.me/mug/12cbe231',
  'https://fynbos.me/mug/07ec2269',
  'https://fynbos.me/mug/8e32f2ea',
  'https://fynbos.me/mug/4a287881',
  'https://fynbos.me/mug/3ebcaafa',
  'https://fynbos.me/mug/432b18fb',
  'https://fynbos.me/mug/4bb5af0a',
  'https://fynbos.me/mug/dcc0536d',
  'https://fynbos.me/mug/e97a03cd',
  'https://fynbos.me/mug/e7153eed',
  'https://fynbos.me/mug/feeae5f1',
  'https://fynbos.me/mug/4e092fe6',
  'https://fynbos.me/mug/ec0691f8',
  'https://fynbos.me/mug/90b867e7',
  'https://fynbos.me/mug/ce2a9e5e',
  'https://fynbos.me/mug/45b343c8',
  'https://fynbos.me/mug/4befcb02',
  'https://fynbos.me/mug/93c565e7',
  'https://fynbos.me/mug/68186be8',
  'https://fynbos.me/mug/2364a83e',
  'https://fynbos.me/mug/f6f8d02c',
  'https://fynbos.me/mug/d017d51b',
  'https://fynbos.me/mug/96534e91',
  'https://fynbos.me/mug/37a16cee',
  'https://fynbos.me/mug/db17d4fc',
  'https://fynbos.me/mug/9ebafa45',
  'https://fynbos.me/mug/551123eb',
  'https://fynbos.me/mug/de600908'
]

for (let url of codes) {
  svgToFile(url)
}
