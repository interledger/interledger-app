import { ActionFunctionArgs, LoaderFunctionArgs } from "react-router"

/**
 * Utility type for extracting loader data type
 * Usage: LoaderData<typeof loader>
 */
export type LoaderData<T extends (args: LoaderFunctionArgs) => any> = Awaited<ReturnType<T>>['data'] 

/**
 * Utility type for extracting action data type
 * Usage: ActionData<typeof action>
 */
export type ActionData<T extends (args: ActionFunctionArgs) => any> = Awaited<ReturnType<T>>['data'] 