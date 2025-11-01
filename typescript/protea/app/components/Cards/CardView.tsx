import clsx from 'clsx'
import { AnimatePresence } from 'framer-motion'
import { CardType } from '~/generated/connect/backend/v1/backend_pb'
import type { StorableCard } from '~/lib/cards/types'
import { panLastFour } from '~/lib/cards/useCardsStore'
import { usePinChangePopup } from '~/lib/cards/usePinChangePopup'
import { useCardActions } from '../../lib/cards/useCardActions'
import { Alert, AlertBody, AlertContent, AlertTitle } from '../Alert'
import { Fade } from '../Animations/Fade'
import { Card, CardContent } from '../Card'
import { Chip, ChipColor } from '../Chip'
import { GridColumn } from '../Grid'
import { Icon } from '../Icon'
import { MasterCardLogo } from '../Logos'
import { AnchorRouter } from '../Router'
import { CardActions } from './CardActions'
import { PhysicalCardChip, VirtualCardChip } from './CardChips'
import { CardViewBack } from './CardViewBack'
import { CardViewFront } from './CardViewFront'
import { StatusPopup } from './StatusPopup'
import { TimedPinPopup } from './TimedPinPopup'

export const CardView = ({ card }: { card: StorableCard }) => {
  const {
    flip,
    showBack,
    isSensitiveDataVisible,
    isPinVisible,
    isFrozen,
    isBlockedByAdmin,
    sensitiveData,
    pin,
    actionStatus,
    toggleSensitiveDataOn,
    toggleFreeze,
    toggleUnfreeze,
    toggleBlock,
    toggleViewPin,
    toggleChangePin
  } = useCardActions(card)

  const { PinChangePopup } = usePinChangePopup()

  return (
    <GridColumn>
      <Card>
        <CardContent className='space-y-4'>
          <div
            className='relative mx-auto h-card-height w-card-width'
            style={{ perspective: '1000px' }}
          >
            <div
              className={clsx(
                'relative h-full w-full transition-transform duration-700 ease-in-out',
                showBack
                  ? '[transform:rotateY(180deg)]'
                  : '[transform:rotateY(0deg)]',
                isFrozen &&
                  'isolate aspect-video w-card-width rounded-xl bg-white/20 shadow-lg ring-1 ring-black/5'
              )}
              style={{ transformStyle: 'preserve-3d' }}
            >
              {/* Loading overlay */}
              {actionStatus === 'loading' && (
                <div className='absolute inset-0 z-50 flex items-center justify-center rounded-xl bg-black/50 backdrop-blur-sm'>
                  <div className='flex flex-col items-center space-y-3'>
                    <div className='h-8 w-8 animate-spin rounded-full border-4 border-white/30 border-t-white'></div>
                    <span className='text-sm font-medium text-white'>
                      Processing...
                    </span>
                  </div>
                </div>
              )}

              <div
                className='absolute inset-0 h-full w-full'
                style={{ backfaceVisibility: 'hidden' }}
              >
                <CardViewFront
                  nameOnCard={card.nameOnCard}
                  className={isBlockedByAdmin ? 'opacity-50 grayscale' : ''}
                />
              </div>

              <div
                className='absolute inset-0 h-full w-full'
                style={{
                  backfaceVisibility: 'hidden',
                  transform: 'rotateY(180deg)'
                }}
              >
                <CardViewBack
                  fullCardNumber={sensitiveData.Pan}
                  expiryDate={sensitiveData.ExpiryDate}
                  cvv={sensitiveData.Cvc2}
                />
              </div>
            </div>

            {isFrozen && (
              <div className='absolute inset-0 h-full w-full rounded-xl bg-gray-700 bg-opacity-10 bg-clip-padding backdrop-blur-sm backdrop-filter'></div>
            )}
          </div>
          <div className='mx-auto flex w-card-width items-center justify-between'>
            <AnimatePresence mode='wait'>
              <Fade nonce={card.id}>
                <Chip color={ChipColor.blue}>
                  <MasterCardLogo size='xs' />
                  &bull;&bull;&bull;&bull; {panLastFour(card.maskedPan)}
                </Chip>
              </Fade>
            </AnimatePresence>
            <AnimatePresence mode='wait'>
              <Fade nonce={card.type.toString()}>
                {card.type === CardType.VIRTUAL ? (
                  <VirtualCardChip />
                ) : (
                  <PhysicalCardChip />
                )}
              </Fade>
            </AnimatePresence>
          </div>
        </CardContent>
      </Card>

      {isBlockedByAdmin ? (
        <Alert>
          <Icon>warning</Icon>
          <AlertContent>
            <AlertTitle>Card was locked</AlertTitle>
            <AlertBody>
              Your card was locked by our provider. Please contact support for
              more information.
              <div className='mt-4 flex items-center space-x-2 text-medium'>
                <Icon>mail</Icon>
                <AnchorRouter
                  to='mailto:support@interledger.app'
                  className='text-sm text-primary'
                >
                  support@interledger.app
                </AnchorRouter>
              </div>
            </AlertBody>
          </AlertContent>
        </Alert>
      ) : (
        <>
          <CardActions
            isPhysical={card.type === CardType.PHYSICAL}
            showBack={showBack}
            flip={flip}
            isSensitiveDataVisible={isSensitiveDataVisible}
            isPinVisible={isPinVisible}
            isFrozen={isFrozen}
            toggleSensitiveDataOn={toggleSensitiveDataOn}
            toggleFreeze={toggleFreeze}
            toggleUnfreeze={toggleUnfreeze}
            toggleViewPin={toggleViewPin}
            toggleBlock={toggleBlock}
            toggleChangePin={toggleChangePin}
          />

          <TimedPinPopup
            pin={pin}
            isVisible={isPinVisible}
            onClose={toggleViewPin}
            duration={7}
          />

          <PinChangePopup />
        </>
      )}

      {(actionStatus === 'success' || actionStatus === 'error') && (
        <StatusPopup
          type={actionStatus === 'success' ? 'success' : 'error'}
          message={
            actionStatus === 'success'
              ? 'Card operation succeeded'
              : 'Card operation failed'
          }
        />
      )}
    </GridColumn>
  )
}
