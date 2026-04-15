import type { UiNodeInputAttributes } from "@ory/client";
import type { KratosFlowWithUi } from "./types.server";


/**
 * Check if a UI node has input attributes
 */
export function isUiNodeInputAttributes(n: unknown): n is UiNodeInputAttributes {
    return typeof n === 'object' && n !== null && 'name' in n
}

/**
 * Extract CSRF token from any Kratos flow with UI
 */
export function getCsrfTokenFromFlow(flow: KratosFlowWithUi | undefined): string {
    if (!flow?.ui?.nodes) return ''

    const node = flow.ui.nodes.find(
        (node) => isUiNodeInputAttributes(node?.attributes) &&
            node.attributes.name === 'csrf_token'
    )

    return node ? (node.attributes as UiNodeInputAttributes).value ?? '' : ''
}

/**
 * Extract a node's value by name from any Kratos flow with UI
 */
export function getNodeValueFromFlow(flow: KratosFlowWithUi | undefined, attributeName: string): string {
    if (!flow?.ui?.nodes) return ''

    const node = flow.ui.nodes.find(
        (node) => isUiNodeInputAttributes(node?.attributes) &&
            node.attributes.name === attributeName
    )

    return node ? (node.attributes as UiNodeInputAttributes).value ?? '' : ''
}

/**
 * Check if node exists in Kratos flow
 */
export function isNodeInFlow(flow: KratosFlowWithUi | undefined, attributeName: string): boolean {
    if (!flow?.ui?.nodes) {
        return false
    }

    const node = flow.ui.nodes.find(
        (node) => isUiNodeInputAttributes(node?.attributes) &&
            node.attributes.name === attributeName
    )

    return !!node
}
