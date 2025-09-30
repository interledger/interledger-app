/**
 * Crypto utility functions for RSA key generation and management
 * Based on testnet implementation with improvements
 */

/**
 * Convert ArrayBuffer to string
 */
export function ab2str(buf: ArrayBuffer): string {
  // @ts-expect-error: We know this works with Uint8Array
  return String.fromCharCode.apply(null, new Uint8Array(buf))
}

/**
 * Check if Web Crypto API is available
 * @returns boolean indicating availability
 */
export const isWebCryptoAvailable =
  typeof window !== 'undefined' &&
  'crypto' in window &&
  'subtle' in window.crypto &&
  typeof window.crypto.subtle.generateKey === 'function'

/**
 * RSA key generation parameters
 */
export const RSA_KEY_PARAMS: RsaHashedKeyGenParams = {
  name: 'RSA-OAEP',
  modulusLength: 2048,
  publicExponent: new Uint8Array([1, 0, 1]),
  hash: { name: 'SHA-256' }
}

/**
 * Key usage permissions
 */
export const KEY_USAGES: KeyUsage[] = ['encrypt', 'decrypt']

/**
 * Validate generated key pair
 * @param keyPair CryptoKeyPair to validate
 * @returns boolean indicating if keys are valid
 */
export function validateKeyPair(keyPair: CryptoKeyPair): boolean {
  return (
    keyPair.privateKey !== null &&
    keyPair.publicKey !== null &&
    keyPair.privateKey.algorithm.name === 'RSA-OAEP' &&
    keyPair.publicKey.algorithm.name === 'RSA-OAEP'
  )
}

/**
 * Export private key to PEM format
 * @param privateKey CryptoKey private key
 * @returns Promise<string> PEM formatted private key
 */
export async function exportPrivateKeyToPEM(
  privateKey: CryptoKey
): Promise<string> {
  const exported = await crypto.subtle.exportKey('pkcs8', privateKey)
  const exportedAsBase64 = btoa(ab2str(exported))
  return `-----BEGIN PRIVATE KEY-----\n${exportedAsBase64}\n-----END PRIVATE KEY-----`
}

/**
 * Export public key to base64 format
 * @param publicKey CryptoKey public key
 * @returns Promise<string> Base64 encoded public key
 */
export async function exportPublicKeyToBase64(
  publicKey: CryptoKey
): Promise<string> {
  const exported = await crypto.subtle.exportKey('spki', publicKey)
  return btoa(ab2str(exported))
}
