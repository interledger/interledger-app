import type { LoaderArgs } from '@remix-run/node'
import { requireUserSession } from '~/lib/kratos.server'
import { FAB, Icon } from '~/components'
import { motion, AnimatePresence } from 'framer-motion'
import { useState } from 'react'

export async function loader({ request }: LoaderArgs) {
  // return requireUserSession(request)
  return null
}

export default function Page() {
  const [show, setShow] = useState<boolean>(false)
  return (
    <div className='flex h-screen w-full items-center justify-center font-display text-5xl font-medium text-medium'>
      Coming soon <span className='text-primary'>.</span>
      {/*<FAB icon='attach_money'></FAB>*/}
      <motion.div
        onClick={() => setShow(true)}
        layoutId='floatingactionbutton'
        className='fixed right-4 bottom-4 flex h-14 w-min items-center space-x-3 rounded-2xl bg-container-primary p-4 font-display text-sm font-medium text-medium shadow-lg hover:bg-container-primary-hover focus-visible:outline-2 focus-visible:outline-focus active:bg-container-primary-active lg:hidden'
      >
        <Icon>attach_money</Icon>
      </motion.div>
      <AnimatePresence>
        {show && (
          <motion.div
            className='fixed inset-0 flex h-screen w-screen'
            layoutId='floatingactionbutton'
          >
            {/*<motion.h5>sub title</motion.h5>*/}
            {/*<motion.h2>title</motion.h2>*/}
            <motion.button onClick={() => setShow(false)}>
              <Icon>close</Icon>
            </motion.button>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
