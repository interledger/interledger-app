import type { StorybookConfig } from '@storybook/react-vite'

const config: StorybookConfig = {
  stories: ['../app/**/*.stories.mdx', '../app/**/*.stories.@(ts|tsx)'],
  addons: [
    '@storybook/addon-links',
    '@storybook/addon-essentials',
    '@storybook/addon-interactions'
    // 'storybook-addon-react-router-v6',
    // ^ Add this if stories use routing hooks (useNavigate, useParams, useLocation, etc.).
    // Also wrap affected stories with the `withRouter` decorator from that package.
    // Not needed currently — no stories use routing hooks.
  ],
  framework: {
    name: '@storybook/react-vite',
    options: {
      builder: {
        viteConfigPath: '.storybook/vite.config.ts'
      }
    }
  },
  docs: {
    autodocs: false
  }
}
export default config
