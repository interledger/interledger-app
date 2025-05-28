interface TestEnvVars {
  EMAIL: string
  PASSWORD: string
  BASE_URL: string
}

declare global {
  namespace NodeJS {
    interface ProcessEnv extends TestEnvVars {}
  }
}

export {}
