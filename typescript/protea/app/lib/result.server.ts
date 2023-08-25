/**
 * @category model
 * @since 2.0.0
 */
export interface Err<E> {
  readonly _tag: 'Err'
  readonly error: E
}

/**
 * @category model
 * @since 2.0.0
 */
export interface Ok<A> {
  readonly _tag: 'Ok'
  readonly value: A
}

/**
 * @category model
 * @since 2.0.0
 */
export type Result<E, A> = Err<E> | Ok<A>

export const isErr = <E>(ma: Result<E, unknown>): ma is Err<E> =>
  ma._tag === 'Err'
export const isOk = <A>(ma: Result<unknown, A>): ma is Ok<A> => ma._tag === 'Ok'
export const err = <E, A = never>(e: E): Result<E, A> => ({
  _tag: 'Err',
  error: e
})
export const ok = <A, E = never>(a: A): Result<E, A> => ({
  _tag: 'Ok',
  value: a
})
