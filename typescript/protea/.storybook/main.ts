import type { StorybookConfig } from '@storybook/react-vite' // (or whatever framework you are using)

import { mergeConfig } from 'vite'

const config: StorybookConfig = {
  stories: ['../app/**/*.stories.mdx', '../app/**/*.stories.@(ts|tsx)'],
  addons: [
    '@storybook/addon-links',
    '@storybook/addon-essentials',
    '@storybook/addon-interactions'
  ],
  features: {
    // interactionsDebugger: true // 👈 Enable playback controls
  },
  core: {
    builder: '@storybook/builder-vite'
  },
  framework: '@storybook/react-vite',
  async viteFinal(config) {
    return mergeConfig(config, {})
  },
  docs: {
    autodocs: false
  }
}
export default config
