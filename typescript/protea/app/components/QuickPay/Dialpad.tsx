import clsx from 'clsx'
import { useEffect, useState } from 'react'
import { getCurrencySymbol } from '~/lib/helpers'
import { useDialPadStore } from '~/lib/useDialPadStore'

export enum DialPadIds {
  Backspace = 'Backspace',
  Dot = '.',
  Zero = '0',
  One = '1',
  Two = '2',
  Three = '3',
  Four = '4',
  Five = '5',
  Six = '6',
  Seven = '7',
  Eight = '8',
  Nine = '9'
}
const handleDialPadInputs = (
  id: string,
  amountValue: string,
  setAmountValue: (amount: string) => void,
  triggerKey?: (key: string) => void
) => {
  if (triggerKey) {
    triggerKey(id)
  }
  const label = id === DialPadIds.Backspace ? '<' : id
  if (Object.values<string>(DialPadIds).includes(id)) {
    if (id === DialPadIds.Backspace) {
      setAmountValue(`${amountValue.substring(0, amountValue.length - 1)}`)
    } else if (amountValue === '0.00' && id !== DialPadIds.Dot) {
      setAmountValue(
        `${amountValue.substring(0, amountValue.length - 4)}${label}`
      )
    } else if (amountValue === '0' && id !== DialPadIds.Dot) {
      setAmountValue(
        `${amountValue.substring(0, amountValue.length - 1)}${label}`
      )
    } else if (
      (id === DialPadIds.Dot &&
        amountValue.indexOf(DialPadIds.Dot) === -1 &&
        amountValue.length !== 0) ||
      id !== DialPadIds.Dot
    ) {
      setAmountValue(`${amountValue}${label}`)
    }
  }
}

export const DialPad = () => {
  const { amountValue, setAmountValue } = useDialPadStore()
  const [activeKey, setActiveKey] = useState<string | null>(null)

  const triggerKey = (key: string) => {
    setActiveKey(String(key))

    setTimeout(() => {
      setActiveKey(null)
    }, 500)
  }

  useEffect(() => {
    const eventHandleDialPadInputs = (e: KeyboardEvent) =>
      handleDialPadInputs(e.key, amountValue, setAmountValue, triggerKey)

    document.addEventListener('keydown', eventHandleDialPadInputs)

    return () => {
      document.removeEventListener('keydown', eventHandleDialPadInputs)
    }
  }, [amountValue, setAmountValue, triggerKey])
  return (
    <div className='flex flex-col text-xl'>
      <AmountDisplay />
      <DialPadRow
        first={DialPadIds.One}
        second={DialPadIds.Two}
        third={DialPadIds.Three}
        activeKey={activeKey}
      />
      <DialPadRow
        first={DialPadIds.Four}
        second={DialPadIds.Five}
        third={DialPadIds.Six}
        activeKey={activeKey}
      />
      <DialPadRow
        first={DialPadIds.Seven}
        second={DialPadIds.Eight}
        third={DialPadIds.Nine}
        activeKey={activeKey}
      />
      <DialPadRow
        first={DialPadIds.Dot}
        second={DialPadIds.Zero}
        third='<'
        idThird={DialPadIds.Backspace}
        activeKey={activeKey}
      />
    </div>
  )
}
DialPad.displayName = 'Dialpad'

type DialPadRowProps = {
  first: string
  second: string
  third: string
  idFirst?: string
  idSecond?: string
  idThird?: string
  activeKey: string | null
}
const DialPadRow = ({
  first,
  second,
  third,
  idFirst,
  idSecond,
  idThird,
  activeKey
}: DialPadRowProps) => {
  return (
    <div>
      <ul className='flex justify-between'>
        <DialPadKey
          label={first}
          id={idFirst ? idFirst : first}
          activeKey={activeKey}
        />
        <DialPadKey
          label={second}
          id={idSecond ? idSecond : second}
          activeKey={activeKey}
        />
        <DialPadKey
          label={third}
          id={idThird ? idThird : third}
          activeKey={activeKey}
        />
      </ul>
    </div>
  )
}
DialPadRow.displayName = 'DialPadRow'

type DialPadKeyProps = {
  label: string
  id: string
  activeKey: string | null
}
const DialPadKey = ({ label, id, activeKey }: DialPadKeyProps) => {
  const { amountValue, setAmountValue } = useDialPadStore()
  const isActive = id == activeKey
  const handleKeyboardNavigation = (e: React.KeyboardEvent<HTMLLIElement>) => {
    if (e.key === 'Enter' || e.key === ' ') {
      handleDialPadInputs(id, amountValue, setAmountValue)
    }
  }

  return (
    <li>
      <span
        role='button'
        className={clsx(
          'flex h-16 w-16 cursor-pointer select-none items-center justify-center rounded-lg text-base font-medium transition-all duration-100 ease-out',
          isActive
            ? 'scale-95 bg-gray-200/60 text-rose-600 dark:bg-white/10 dark:text-rose-600'
            : 'text-gray-700 dark:text-gray-300',
          'hover:bg-gray-200/60 hover:text-rose-600 active:scale-95 dark:hover:bg-white/10 dark:hover:text-rose-600'
        )}
        tabIndex={0}
        id={id}
        onClick={() => handleDialPadInputs(id, amountValue, setAmountValue)}
        onKeyDown={(e: React.KeyboardEvent<HTMLLIElement>) =>
          handleKeyboardNavigation(e)
        }
      >
        {label}
      </span>
    </li>
  )
}
DialPadKey.displayName = 'DialPadKey'

type AmountDisplayProps = {
  displayAmount?: string
  assetCode?: string
}

export const AmountDisplay = (args: AmountDisplayProps) => {
  const { amountValue, assetCode } = useDialPadStore()

  const value = args.displayAmount
    ? `${getCurrencySymbol(args?.assetCode ?? 'usd')} ${args.displayAmount}`
    : `${getCurrencySymbol(assetCode)} ${amountValue}`

  return (
    <div className='amount-display text-green-1 flex w-full items-center justify-center whitespace-nowrap text-5xl'>
      {value}
    </div>
  )
}
AmountDisplay.displayName = 'AmountDisplay'
