import ky from 'ky'
import { useRouter } from 'next/router'
import { Dispatch, SetStateAction, useEffect, useState } from 'react'

export const getPreview = async (): Promise<{ preview: boolean }> => {
  return ky.get('http://protea/api/preview').json()
}

export const setPreview = async (on: boolean): Promise<boolean> => {
  return ky
    .get(`http://fynbos.test/api/preview?toggle=${on ? 'on' : 'off'}`)
    .json()
}

export const usePreview = (
  initial = true
): [boolean, Dispatch<SetStateAction<boolean>>] => {
  const [preview, setPrev] = useState<boolean>(initial)
  const router = useRouter()

  useEffect(() => {
    if (!router.isReady) {
      return
    }

    // If router preview is not the same as stored preview set it as stored preview
    // router.isPreview won't update dynamically if the preview state is changed.
    if (router.isPreview != preview) {
      setPreview(preview).then((val) => setPrev(val as boolean))
    }
  }, [router, router.isReady, router.isPreview, preview])

  return [preview, setPrev]
}
