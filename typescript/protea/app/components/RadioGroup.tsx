import { RadioGroup as HeadlessRadioGroup } from '@headlessui/react'
import type { FC } from 'react'
import { Icon } from '.'

export type RadioGroupOption = {
  id: string
  name: string
  description?: string
  icon: string
  disabled?: boolean,
  label?: string
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
      <HeadlessRadioGroup
        id={id}
        disabled={disabled}
        value={value}
        onChange={onChange}
      >
        <HeadlessRadioGroup.Label className='sr-only'>
          {label}
        </HeadlessRadioGroup.Label>
        <div className='space-y-4'>
          {options.map((option) => (
            <HeadlessRadioGroup.Option
              key={option.id}
              disabled={option.disabled}
              value={option}
              className={({ checked, disabled }) =>
                `${checked ? 'bg-container-primary' : 'bg-nav'} ${
                  disabled ? 'cursor-not-allowed bg-disabled text-disabled' : ''
                }
                    relative flex cursor-pointer rounded-xl p-3 outline-focus transition-all duration-300 focus-visible:outline focus-visible:outline-2 focus-visible:outline-focus`
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
                        {option.icon && <Icon>{option.icon}</Icon>}
                      <div className='flex flex-col'>
                        <HeadlessRadioGroup.Label as='span'>
                              {option.name}
                            </HeadlessRadioGroup.Label>
                          {option.description && (
                            <HeadlessRadioGroup.Description
                              as='span'
                              className={`text-xs ${
                                disabled ? 'text-disabled' : 'text-weak'
                              }`}
                            >
                              {option.description}
                            </HeadlessRadioGroup.Description>
                          )}
                      </div>
                    </div>
                    <div
                      className={`flex items-center transition-all duration-300 ${
                        disabled
                          ? 'text-disabled'
                          : checked
                          ? 'text-primary'
                          : 'text-medium'
                      }`}
                    >
                      {option.label && (
                        <span className='mr-4 flex-shrink-0 rounded-full bg-primary px-2 py-1 text-xs font-medium text-white'>
                          {option.label}
                        </span>
                      )}
                      {checked && <Icon>radio_button_checked</Icon>}
                      {!checked && <Icon>radio_button_unchecked</Icon>}
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
