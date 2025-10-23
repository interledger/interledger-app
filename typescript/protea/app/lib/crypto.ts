import NodeRSA from 'node-rsa'

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

export const RSA_KEY_PARAMS: RsaHashedKeyGenParams = {
  name: 'RSA-OAEP',
  modulusLength: 2048,
  publicExponent: new Uint8Array([1, 0, 1]),
  hash: { name: 'SHA-256' }
}

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

/**
 * Decrypt data using private key with RSA/ECB/PKCS1Padding
 * Note: Web Crypto API doesn't support PKCS1 padding directly, so we use NodeRSA library
 * @param privateKeyString PEM formatted private key
 * @param encryptedData Base64 encoded encrypted data
 * @returns Promise<T> Decrypted data
 */
export async function decryptWithPrivateKey<T>(
  privateKeyString: string,
  encryptedData: string
): Promise<T> {
  const privateKey = new NodeRSA(privateKeyString)
  privateKey.setOptions({
    encryptionScheme: 'pkcs1',
    environment: 'browser'
  })

  const decryptedRequestData = privateKey
    .decrypt(encryptedData)
    .toString('utf8')

  const cardData = JSON.parse(decryptedRequestData)
  return cardData as T
}

export async function encryptWithPublicKey(
  publicKey: string,
  data: string
): Promise<string> {
  const pemPublicKey = `-----BEGIN PUBLIC KEY-----\n${publicKey}\n-----END PUBLIC KEY-----`
  
  const key = new NodeRSA(pemPublicKey)
  key.setOptions({
    encryptionScheme: 'pkcs1',
    environment: 'browser'
  })
  return key.encrypt(data, 'base64')
}

export function parseJwt(token: string) {
  const base64Url = token.split('.')[1]
  const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
  const jsonPayload = decodeURIComponent(
    atob(base64)
      .split('')
      .map(function (c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
      })
      .join('')
  )

  return JSON.parse(jsonPayload)
}

export async function encryptWithToken(
  token: string,
  data: string
): Promise<string> {
  const { publicKey } = parseJwt(token) as {
    publicKey: string
  }
  
  return await encryptWithPublicKey(publicKey, data)
}
