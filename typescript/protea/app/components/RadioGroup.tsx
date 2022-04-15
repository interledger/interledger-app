import { RadioGroup as HeadlessRadioGroup } from '@headlessui/react'
import type { FC } from 'react'
import React from 'react'
import { RadioActiveIcon, RadioIcon, Icons } from '.'

export type RadioGroupOption = {
  id: string
  name: string
  description: string
  icon: keyof typeof Icons
  disabled?: boolean
}

interface RadioGroupProps {
  id: string
  // Override the `className` of the root `div` of the Input. Defaults to **min-w-full**.
  className?: string
  disabled?: boolean
  // The label value.
  label?: string
  value?: RadioGroupOption
  options: RadioGroupOption[]
  onChange(value: RadioGroupOption): void
}

/**
 * RadioGroup component should be used instead of <input type="radio">
 * NOTE: a hidden input is required to submit to remix.
 * @param param0
 * @returns
 */
export const RadioGroup: FC<RadioGroupProps> = ({
  id,
  className,
  label,
  value,
  onChange,
  options,
  disabled
}) => {
  return (
    <div className={className || 'min-w-full'}>
      <HeadlessRadioGroup disabled={disabled} value={value} onChange={onChange}>
        <HeadlessRadioGroup.Label className='sr-only'>
          {label}
        </HeadlessRadioGroup.Label>
        <div className='space-y-2'>
          {options.map((option) => (
            <HeadlessRadioGroup.Option
              key={option.id}
              disabled={option.disabled}
              value={option}
              className={({ checked, disabled }) =>
                `${checked ? 'bg-container-primary' : 'bg-container'} ${
                  disabled ? 'cursor-not-allowed bg-disabled text-disabled' : ''
                }
                    relative flex cursor-pointer rounded-xl p-3 outline-primary transition-all duration-300 focus-visible:outline-2 focus-visible:outline-primary`
              }
            >
              {({ checked, disabled }) => (
                <>
                  <div className='flex w-full items-center justify-between'>
                    <div
                      className={`flex items-center space-x-3 ${
                        disabled ? 'text-disabled' : 'text-medium'
                      }`}
                    >
                      {option.icon && Icons[option.icon]()}
                      <div className='flex flex-col'>
                        <HeadlessRadioGroup.Label
                          as='span'
                          className='font-sans text-base font-normal'
                        >
                          {option.name}
                        </HeadlessRadioGroup.Label>
                        <HeadlessRadioGroup.Description
                          as='span'
                          className={`font-sans text-xs font-normal ${
                            disabled ? 'text-disabled' : 'text-weak'
                          }`}
                        >
                          {option.description}
                        </HeadlessRadioGroup.Description>
                      </div>
                    </div>
                    <div
                      className={`transition-all duration-300 ${
                        disabled
                          ? 'text-disabled'
                          : checked
                          ? 'text-primary'
                          : 'text-medium'
                      }`}
                    >
                      {checked && <RadioActiveIcon />}
                      {!checked && <RadioIcon />}
                    </div>
                  </div>
                </>
              )}
            </HeadlessRadioGroup.Option>
          ))}
        </div>
      </HeadlessRadioGroup>
    </div>
  )
}

RadioGroup.displayName = 'RadioGroup'
