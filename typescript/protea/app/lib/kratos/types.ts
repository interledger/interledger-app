import type {
    LoginFlow,
    LogoutFlow,
    RecoveryFlow,
    RegistrationFlow,
    Session,
    SettingsFlow,
    UiNode,
    UiNodeInputAttributes,
    UpdateLoginFlowWithPasswordMethod,
    UpdateLoginFlowWithTotpMethod,
    UpdateRecoveryFlowWithLinkMethod,
    UpdateRegistrationFlowWithPasswordMethod,
    UpdateSettingsFlowWithPasswordMethod,
    UpdateSettingsFlowWithProfileMethod,
    UpdateSettingsFlowWithTotpMethod,
    UpdateVerificationFlowWithLinkMethod,
    VerificationFlow
} from "@ory/client";
import { kratosPublic } from "./kratos-client.server";

// Re-export types for convenience
export type {
    LoginFlow,
    LogoutFlow,
    RecoveryFlow,
    RegistrationFlow,
    SettingsFlow,
    VerificationFlow,
    Session,
    UiNodeInputAttributes,
    UiNode,
    UpdateLoginFlowWithPasswordMethod,
    UpdateLoginFlowWithTotpMethod,
    UpdateRegistrationFlowWithPasswordMethod,
    UpdateSettingsFlowWithPasswordMethod,
    UpdateSettingsFlowWithProfileMethod,
    UpdateSettingsFlowWithTotpMethod,
    UpdateRecoveryFlowWithLinkMethod,
    UpdateVerificationFlowWithLinkMethod
}

export type KratosFlowWithUi =
    LoginFlow |
    RecoveryFlow |
    RegistrationFlow |
    SettingsFlow |
    VerificationFlow

export type KratosError = {
    id: string,
    response: {
        status: number
        data: any
        headers: Record<string, unknown>
    }
}

export type CreateBrowserLoginFlowResponse = Awaited<ReturnType<typeof kratosPublic.createBrowserLoginFlow>>

/** 
 * Kratos error IDs for error message overrides
 */
export enum KratosErrorId {
    ErrorValidationInvalidCredentials = 4000006,
    ErrorValidationDuplicateCredentials = 4000007
}

export type KratosMessage = {
    id: number
    text: string
    context?: object
}

/**
 * Axios request config type (inline to avoid axios dependency)
 */
export type RequestConfig = {
    headers?: Record<string, string>
    withCredentials?: boolean
}

