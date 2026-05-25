import { Form } from 'react-router'
import type { Dispatch, FC, ReactNode, SetStateAction } from 'react'
import { TextButton } from './Buttons'
import { CardContent, CardHeader } from './Card'
import { Dialog } from './Dialog'

type ConfirmDialogProps = {
  open: boolean
  setOpen: Dispatch<SetStateAction<boolean>>
  title: string
  body: ReactNode
  confirmLabel: string
  /** Form action path. Defaults to the current route. */
  confirmAction?: string
  /** Hidden inputs injected into the confirm Form. */
  hiddenInputs?: Record<string, string>
  /** Renders the confirm button in the error/destructive colour. */
  destructive?: boolean
}

export const ConfirmDialog: FC<ConfirmDialogProps> = ({
  open,
  setOpen,
  title,
  body,
  confirmLabel,
  confirmAction,
  hiddenInputs = {},
  destructive = false
}) => {
  return (
    <Dialog open={open} setOpen={setOpen}>
      <CardHeader>
        <h1 className='text-xl font-medium'>{title}</h1>
      </CardHeader>
      <CardContent>
        <span className='text-medium'>{body}</span>
        <Form
          method='post'
          action={confirmAction}
          onSubmit={() => setOpen(false)}
        >
          {Object.entries(hiddenInputs).map(([name, value]) => (
            <input key={name} type='hidden' name={name} value={value} />
          ))}
          <div className='flex w-full justify-end space-x-6 pt-4'>
            <TextButton
              type='button'
              className='!text-medium'
              onClick={() => setOpen(false)}
            >
              Cancel
            </TextButton>
            <TextButton
              type='submit'
              className={destructive ? '!text-error' : undefined}
            >
              {confirmLabel}
            </TextButton>
          </div>
        </Form>
      </CardContent>
    </Dialog>
  )
}

ConfirmDialog.displayName = 'ConfirmDialog'
