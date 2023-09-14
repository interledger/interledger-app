import clsx from 'clsx'
import type { FC, ReactNode } from 'react'

// TODO: Refactor to use static string types for colours rather
export enum ChipColor {
  red = 'bg-chip-red text-chip-red',
  sky = 'bg-chip-sky text-chip-sky',
  green = 'bg-chip-green text-chip-green',
  orange = 'bg-chip-orange text-chip-orange',
  blue = 'bg-chip-blue text-chip-blue',
  indigo = 'bg-chip-indigo text-chip-indigo',
  purple = 'bg-chip-purple text-chip-purple',
  yellow = 'bg-chip-yellow text-chip-yellow',
  lime = 'bg-chip-lime text-chip-lime',
  rose = 'bg-chip-rose text-chip-rose',
  slate = 'bg-chip-slate text-chip-slate'
}

export type ChipProps = {
  children?: ReactNode
  color: ChipColor
}

export const Chip: FC<ChipProps> = ({ children, color = ChipColor.blue }) => {
  return (
    <div
      className={clsx(
        'flex items-center justify-center rounded px-1.5 py-1 text-xs',
        color
      )}
    >
      {children}
    </div>
  )
}
