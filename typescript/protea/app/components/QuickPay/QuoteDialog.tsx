import { Form } from 'react-router'
import { Button } from '../Buttons'
import { CardHeader } from '../Card'
import { Dialog } from '../Dialog'
import { TextField } from '../TextField'

export type QuoteArgs = {
  receiverName: string
  receiveAmount: string
  debitAmount: string
  showDialog: boolean
  setShowDialog: React.Dispatch<React.SetStateAction<boolean>>
}

export function QuoteDialog({
  receiverName,
  receiveAmount,
  debitAmount,
  showDialog,
  setShowDialog
}: QuoteArgs) {
  return (
    <Dialog open={showDialog} setOpen={setShowDialog}>
      <div className='flex h-full flex-col justify-center gap-10 p-2'>
        <div className='mx-auto w-full max-w-sm'>
          <CardHeader className='mb-2'>
            <h1 className='text-xl font-medium'>Confirm payment</h1>
          </CardHeader>
          <Form method='POST'>
            <TextField label='You send:' defaultValue={debitAmount} readOnly />
            <TextField
              className='mt-1'
              label={`${receiverName} receives (approximately): `}
              defaultValue={receiveAmount}
              readOnly
            />
            <div className='mt-5 flex items-center justify-center gap-3'>
              <Button
                aria-label='confirm-pay'
                type='submit'
                value='confirm'
                name='intent'
              >
                Confirm
              </Button>
              <Button
                aria-label='cancel-pay'
                type='submit'
                value='cancel'
                name='intent'
                onClick={() => setShowDialog(false)}
              >
                Cancel
              </Button>
            </div>
          </Form>
        </div>
      </div>
    </Dialog>
  )
}
