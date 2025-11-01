import { Chip, ChipColor } from '../Chip'

export const VirtualCardChip = () => {
  return <Chip color={ChipColor.purple}>Virtual</Chip>
}

export const PhysicalCardChip = () => {
  return <Chip color={ChipColor.indigo}>Physical</Chip>
}
