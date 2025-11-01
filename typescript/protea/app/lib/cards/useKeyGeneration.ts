import { useCallback, useEffect, useState } from 'react'
import {
  KEY_USAGES,
  RSA_KEY_PARAMS,
  exportPrivateKeyToPEM,
  exportPublicKeyToBase64,
  isWebCryptoAvailable,
  validateKeyPair
} from '../crypto'

export interface KeyPair {
  publicKey: string
  privateKey: string
}

export interface KeyGenerationState {
  keys: KeyPair | null
  isGenerating: boolean
  error: string | null
  isSupported: boolean
}

export interface UseKeyGenerationReturn {
  keyPair: KeyPair | null
  generateKeys: () => Promise<void>
  clearKeys: () => void
  regenerateKeys: () => Promise<void>
  hasKeys: boolean
}

export interface UseKeyGenerationOptions {
  autoGenerate?: boolean
}

/**
 * Custom hook for RSA key generation and management
 * Based on testnet UserCardContext implementation with improvements
 *
 * @param options Configuration options
 * @returns Key generation state and control functions
 */
export function useKeyGeneration(
  options: UseKeyGenerationOptions = {}
): UseKeyGenerationReturn {
  const { autoGenerate = true } = options

  const [state, setState] = useState<KeyGenerationState>({
    keys: null,
    isGenerating: false,
    error: null,
    isSupported: isWebCryptoAvailable
  })

  const generateKeys = useCallback(async (): Promise<void> => {
    if (!state.isSupported) {
      setState((prev) => ({
        ...prev,
        error: 'Web Crypto API is not supported in this browser'
      }))
      return
    }

    setState((prev) => ({ ...prev, isGenerating: true, error: null }))

    try {
      const keyPair = await crypto.subtle.generateKey(
        RSA_KEY_PARAMS,
        true,
        KEY_USAGES
      )

      if (!validateKeyPair(keyPair)) {
        throw new Error('Generated key pair validation failed')
      }

      const [privateKeyPEM, publicKeyBase64] = await Promise.all([
        exportPrivateKeyToPEM(keyPair.privateKey),
        exportPublicKeyToBase64(keyPair.publicKey)
      ])

      const keys: KeyPair = {
        publicKey: publicKeyBase64,
        privateKey: privateKeyPEM
      }

      setState((prev) => ({
        ...prev,
        keys,
        isGenerating: false,
        error: null
      }))
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to generate keys'

      setState((prev) => ({
        ...prev,
        isGenerating: false,
        error: errorMessage
      }))
    }
  }, [state.isSupported])

  const clearKeys = useCallback((): void => {
    setState((prev) => ({ ...prev, keys: null, error: null }))
  }, [])

  const regenerateKeys = useCallback(async (): Promise<void> => {
    clearKeys()
    await generateKeys()
  }, [clearKeys, generateKeys])

  useEffect(() => {
    if (!state.isSupported) {
      console.error('Web Crypto API is not supported in this browser')
      return
    }

    // Auto-generate if enabled
    if (autoGenerate && !state.keys) {
      generateKeys()
    }
  }, [autoGenerate, state.isSupported, state.keys, generateKeys])

  return {
    keyPair: state.keys,
    generateKeys,
    clearKeys,
    regenerateKeys,
    hasKeys: state.keys !== null
  }
}
