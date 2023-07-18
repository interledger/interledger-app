import clsx from 'clsx'

function ListItem({
  title,
  body,
  parentKey
}: {
  title: string
  body: string
  parentKey?: string
}) {
  return (
    <div className='flex flex-col space-y-1'>
      <dt className='text-xs font-medium capitalize text-weak'>
        {parentKey && `${parentKey}: `}
        {title}
      </dt>
      <dd className='text-strong'>{body || '-'}</dd>
    </div>
  )
}

function itemsIterator(items: any, parentKey?: string): any {
  if (Array.isArray(items)) {
    return items.map((value, index) => {
      if (typeof value == 'object') {
        return itemsIterator(value, `${parentKey}[${index}]`)
      } else
        return (
          <ListItem
            key={index}
            title={`${parentKey}[${index}]`}
            body={value as string}
            parentKey={parentKey}
          />
        )
    })
  }
  return Object.entries(items).map(([key, value]) => {
    if (typeof value == 'object') {
      return itemsIterator(value, key)
    } else
      return (
        <ListItem
          key={key}
          title={key}
          body={value as string}
          parentKey={parentKey}
        />
      )
  })
}

export function GridCard({
  title,
  options,
  className
}: {
  title?: string
  options: any
  className?: string
}) {
  return (
    <dl
      className={clsx(
        className,
        'flex h-max max-h-max flex-col space-y-4 rounded-2xl bg-page p-4'
      )}
    >
      {title && <h2 className='font-display text-lg font-medium'>{title}</h2>}
      {itemsIterator(options)}
    </dl>
  )
}

export function GridCardError({
  error,
  className
}: {
  error: any
  className?: string
}) {
  return (
    <dl
      className={clsx(
        className,
        'flex h-max max-h-max flex-col space-y-4 rounded-2xl bg-page p-4'
      )}
    >
      {error.status && (
        <h2 className='font-display text-5xl font-medium text-error'>
          {error.status}
        </h2>
      )}
      {error.statusText && (
        <span className='text-lg font-medium text-medium'>
          {error.statusText}
        </span>
      )}
      {itemsIterator(error.data)}
    </dl>
  )
}
