import {
  createContext,
  createElement,
  useContext,
  useEffect,
  useState,
  type ReactNode
} from 'react'

export interface PtiWidgetConfig {
  sdkUrl: string
  formsUrl: string
  clientId?: string
  generateTokenPath?: string
}

const PtiContext = createContext<PtiWidgetConfig | null>(null)

export function PtiConfigProvider({ children }: { children: ReactNode }) {
  const [config, setConfig] = useState<PtiWidgetConfig | null>(null)

  useEffect(() => {
    const fetchPtiConfig = async () => {
      try {
        const response = await fetch('/api/pti/widget-config')
        if (!response.ok) {
          throw new Error(`Failed to fetch PTI config: ${response.statusText}`)
        }

        const data = (await response.json()) as PtiWidgetConfig
        setConfig(data)
      } catch {
        setConfig({ sdkUrl: '', formsUrl: '' })
      }
    }

    void fetchPtiConfig()
  }, [])

  return createElement(PtiContext.Provider, { value: config }, children)
}

export function usePtiConfig(): PtiWidgetConfig | null {
  return useContext(PtiContext)
}
