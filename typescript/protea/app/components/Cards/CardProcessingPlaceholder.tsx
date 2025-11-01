import { Card, CardContent, Icon } from '~/components'

export const CardProcessingPlaceholder = () => {
  return (
    <Card>
      <CardContent className='flex flex-col items-center space-y-4 py-8'>
        <div className='flex h-16 w-16 items-center justify-center rounded-full bg-nav'>
          <Icon className='-translate-x-0.5 -translate-y-px text-3xl text-medium'>
            hourglass_empty
          </Icon>
        </div>
        <div className='flex flex-col items-center space-y-2 text-center'>
          <h3 className='text-lg font-medium text-strong'>
            Your first card is being processed
          </h3>
          <p className='max-w-md text-sm text-weak'>
            We're preparing your card. This usually takes a few moments. You'll
            be able to view and manage your card once processing is complete.
          </p>
        </div>
      </CardContent>
    </Card>
  )
}
