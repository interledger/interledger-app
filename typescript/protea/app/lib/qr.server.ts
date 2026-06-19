import QRCode from 'qrcode'
import type { Radius } from '~/components'

type QRdot = {
  shape: Radius
  fill: boolean
}

const svgShapes: { [K in Radius]: (x: number, y: number) => string } = {
  'rounded-none': (x: number, y: number) =>
    `<rect x='${x}' y='${y}'  width='40' height='40' fill='black' stroke='black'/>`,
  'rounded-full': (x: number, y: number) =>
    `<rect x='${x}' y='${y}' width='40' height='40' rx='20' fill='black' stroke='black'/>`,
  'rounded-t-full': (x: number, y: number) =>
    `<path d='M${x} ${y}m0 20a20 20 -180 0 1 40 0v20h-40v-20Z' fill='black' stroke='black' />`,
  'rounded-tl-full': (x: number, y: number) =>
    `<path fill='black' d='M${x} ${y}m40 0v40h-40a40 40 -90 0 1 40 -40Z' stroke='black' />`,
  'rounded-tr-full': (x: number, y: number) =>
    `<path fill='black' d='M${x} ${y}a40 40 -90 0 1 40 40h-40v-40Z' stroke='black' />`,
  'rounded-b-full': (x: number, y: number) =>
    `<path fill='black' d='M${x} ${y}h40v20a20 20 -180 0 1 -40 0v-20Z' stroke='black' />`,
  'rounded-bl-full': (x: number, y: number) =>
    `<path fill='black' d='M${x} ${y}h40v40a40 40 -90 0 1 -40 -40Z' stroke='black' />`,
  'rounded-br-full': (x: number, y: number) =>
    `<path fill='black' d='M${x} ${y}h40a40 40 90 0 1 -40 40v-40Z' stroke='black' />`,
  'rounded-l-full': (x: number, y: number) =>
    `<path fill='black' d='M${x} ${y}m40 0v40h-20a20 20 -180 0 1 0 -40h20Z' stroke='black' />`,
  'rounded-r-full': (x: number, y: number) =>
    `<path fill='black' d='M${x} ${y}h20a20 20 -180 0 1 0 40h-20v-40Z' stroke='black' />`,
  // TODO these need to be implemented if they get used
  'rounded-full rounded-tl-none': (x: number, y: number) =>
    `<rect x='${x}' y='${y}' width='40' height='40' rx='20' fill='black' stroke='black'/>`,
  'rounded-full rounded-tr-none': (x: number, y: number) =>
    `<rect x='${x}' y='${y}' width='40' height='40' rx='20' fill='black' stroke='black'/>`,
  'rounded-full rounded-bl-none': (x: number, y: number) =>
    `<rect x='${x}' y='${y}' width='40' height='40' rx='20' fill='black' stroke='black'/>`,
  'rounded-full rounded-br-none': (x: number, y: number) =>
    `<rect x='${x}' y='${y}' width='40' height='40' rx='20' fill='black' stroke='black'/>`,
  'rounded-tr-full rounded-bl-full': (x: number, y: number) =>
    `<rect x='${x}' y='${y}' width='40' height='40' rx='20' fill='black' stroke='black'/>`,
  'rounded-tl-full rounded-br-full': (x: number, y: number) =>
    `<rect x='${x}' y='${y}' width='40' height='40' rx='20' fill='black' stroke='black'/>`
}

const logo = (center: number) => `<path fill='#FEF08A' d='M${center - 80} ${
  center - 80
}a80 80 -90 0 1 80 80h-80v-80Z' />
<rect x='${center}' y='${
  center - 80
}' width='80' height='80' rx='40' fill='#F43F5E'/>
<path fill='#84CC16' d='M${
  center - 80
} ${center}h80v80a80 80 -90 0 1 -80 -80Z' />
<path fill='#A3E635' d='M${center} ${center}h80a80 80 90 0 1 -80 80v-80Z' />`

export function qrSvg(qr: QRdot[][]): string {
  // Can't change blockSize yet, need to configure svg paths to handle this
  const blockSize = 40
  const qrLen = qr.length
  let svg: string = `<svg width='100%' viewBox='0 0 ${qrLen * blockSize} ${
    qrLen * blockSize
  }' fill='none' xmlns='http://www.w3.org/2000/svg'>`

  svg += logo((qrLen * blockSize) / 2)
  for (let i = 0; i < qrLen; i++) {
    for (let j = 0; j < qrLen; j++) {
      const center = Math.floor(qrLen / 2)
      if (
        i >= center - 3 &&
        i <= center + 3 &&
        j >= center - 3 &&
        j <= center + 3
      ) {
        // continue
        // svg += svgShapes['rounded-r-full'](j * blockSize, i * blockSize)
      } else if (qr[i][j].fill)
        svg += svgShapes[qr[i][j].shape](j * blockSize, i * blockSize)
    }
  }

  svg += `</svg>`
  return svg
}

export async function generateQR(text: string): Promise<QRdot[][]> {
  const qr = await QRCode.create(text, {
    errorCorrectionLevel: 'H'
  })

  // Set Defaults
  const qrArr: QRdot[][] = []
  let eyeSize = 0

  let i = 0
  while (eyeSize == 0) {
    const current = qr.modules.get(0, i)
    const next = qr.modules.get(0, i + 1)
    if (current != next) {
      eyeSize = i + 1
    }
    i++
  }

  for (i = 0; i < qr.modules.size; i++) {
    qrArr.push([])
    for (let j = 0; j < qr.modules.size; j++) {
      let shape: Radius = 'rounded-none'
      const fill = qr.modules.get(i, j) > 0
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
        const topAdj = qr.modules.get(i - 1, j) ?? 0
        const leftAdj = j == 0 ? 0 : (qr.modules.get(i, j - 1) ?? 0)
        const rightAdj =
          j == qr.modules.size - 1 ? 0 : (qr.modules.get(i, j + 1) ?? 0)
        const bottomAdj = qr.modules.get(i + 1, j) ?? 0

        if (!topAdj && !leftAdj && !rightAdj && !bottomAdj) {
          // TODO possibly use other new shapes for this case
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
        fill,
        shape
      }
    }
  }
  return qrArr
}
