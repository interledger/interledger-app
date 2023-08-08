import { motion } from 'framer-motion'
import type { FC } from 'react'
import { useEffect } from 'react'
import { useScaffoldStore } from '~/lib/useScaffoldStore'

interface SnackbarProps {}

export const Snackbar: FC<SnackbarProps> = ({}) => {
  /**
   * TODO
   * - [ ] Add reducers for actions and icons so that they can be passed as strings
   * - [ ] Change the default timers: 10s for action, 4s for icon.
   * - [ ] Use a ref for the timer
   */
  const dismissAfter = 4000

  const [snackbars, shiftSnackbar] = useScaffoldStore((state) => [
    state.snackbars,
    state.shiftSnackbar
  ])

  useEffect(() => {
    let timer: NodeJS.Timeout
    if (dismissAfter) {
      timer = setTimeout(() => {
        shiftSnackbar()
      }, dismissAfter)
    }
    return () => clearTimeout(timer)
  }, [dismissAfter, shiftSnackbar])

  if (snackbars.findIndex((s) => s.canShow) == -1) return null
  return (
    <motion.div
      key='snackbar'
      animate={{ opacity: 1, scale: 1, y: 0 }}
      initial={{ opacity: 0, scale: 0.5, y: 8 }}
      exit={{
        opacity: 0,
        scale: 0.5,
        y: 8,
        transition: {
          duration: 0.2
        }
      }}
      transition={{
        type: 'spring',
        stiffness: 400,
        damping: 20,
        duration: 0.3
      }}
      onClick={() => shiftSnackbar()}
      className='z-50 flex w-full items-center justify-between space-x-3 overflow-hidden rounded-xl bg-snackbar px-4 py-3 text-left align-middle shadow-lg sm:max-w-xs'
    >
      <span className='text-inverted'>
        {snackbars.find((s) => s.canShow)?.message}
      </span>
    </motion.div>
  )

  // return (
  //   <Transition
  //     id={id}
  //     appear
  //     show={show}
  //     as={'div'}
  //     className={clsx(
  //       'fixed left-0 z-[100] mx-auto w-full overflow-y-visible lg:bottom-auto lg:top-4',
  //       xOffset ? 'lg:pl-64' : '',
  //       yOffset ? 'bottom-32' : 'bottom-4'
  //     )}
  //   >
  //     <div className='flex justify-center text-center'>
  //       <Transition.Child
  //         as={Fragment}
  //         enter='ease-out duration-300'
  //         enterFrom='opacity-0 scale-95'
  //         enterTo='opacity-100 scale-100'
  //         leave='ease-in duration-200'
  //         leaveFrom='opacity-100 scale-100'
  //         leaveTo='opacity-0 scale-95'
  //       >
  //         <div className='mx-4 flex w-full transform items-center justify-between space-x-3 overflow-hidden rounded-xl bg-snackbar px-4 py-3 text-left align-middle shadow-lg transition-all sm:max-w-[22rem]'>
  //           <p className='text-sm text-inverted'>{message}</p>
  //           {action && (
  //             <TextButton onClick={() => onClose()}>{action}</TextButton>
  //           )}
  //           {icon && (
  //             <div className='-mr-2'>
  //               <IconButton className='text-inverted' onClick={() => onClose()}>
  //                 {icon}
  //               </IconButton>
  //             </div>
  //           )}
  //         </div>
  //       </Transition.Child>
  //     </div>
  //   </Transition>
  // )
}
