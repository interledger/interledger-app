const colors = require('tailwindcss/colors')

module.exports = {
  mode: 'jit',
  content: ['./app/**/*.{ts,tsx}'],
  darkMode: 'media',
  theme: {
    fontFamily: {
      display: ['Poppins'],
      sans: ['Inter'],
      mono: ['"JetBrains Mono"'],
      icon: ['"Material Symbols Outlined"', { fontFeatureSettings: '"liga"' }]
    },
    extend: {
      colors: {
        brand: colors.rose[500]
      },
      // Token colours
      textColor: {
        strong: 'rgb(var(--text-strong) / <alpha-value>)',
        medium: 'rgb(var(--text-medium) / <alpha-value>)',
        weak: 'rgb(var(--text-weak) / <alpha-value>)',
        disabled: 'rgb(var(--text-disabled) / <alpha-value>)',
        primary: 'rgb(var(--text-primary) / <alpha-value>)',
        'primary-hover': 'rgb(var(--text-primary-hover) / <alpha-value>)',
        error: 'rgb(var(--text-error) / <alpha-value>)',
        success: 'rgb(var(--text-success) / <alpha-value>)'
      },
      backgroundColor: {
        app: 'rgb(var(--bg-app) / <alpha-value>)',
        page: 'rgb(var(--bg-page) / <alpha-value>)',
        container: 'rgb(var(--bg-container) / <alpha-value>)',
        'container-hover': 'rgb(var(--bg-container-hover) / <alpha-value>)',
        strong: 'rgb(var(--bg-strong) / <alpha-value>)',
        disabled: 'rgb(var(--bg-disabled) / <alpha-value>)',
        primary: 'rgb(var(--bg-primary) / <alpha-value>)',
        'container-primary': 'rgb(var(--bg-container-primary) / <alpha-value>)',
        'container-primary-hover':
          'rgb(var(--bg-container-primary-hover) / <alpha-value>)',
        'container-primary-active':
          'rgb(var(--bg-container-primary-active) / <alpha-value>)',
        'container-secondary':
          'rgb(var(--bg-container-secondary) / <alpha-value>)',
        scrim: 'rgb(var(--bg-scrim) / <alpha-value>)',
        snackbar: 'rgb(var(--bg-snackbar) / <alpha-value>)',
        footer: 'rgb(var(--bg-footer) / <alpha-value>)'
      },
      borderColor: {
        base: 'rgb(var(--border-base) / <alpha-value>)',
        focus: 'rgb(var(--border-focus) / <alpha-value>)',
        hover: 'rgb(var(--border-hover) / <alpha-value>)',
        active: 'rgb(var(--border-active) / <alpha-value>)',
        error: 'rgb(var(--border-error) / <alpha-value>)'
      },
      ringColor: {
        base: 'rgb(var(--border-base) / <alpha-value>)',
        focus: 'rgb(var(--border-focus) / <alpha-value>)',
        hover: 'rgb(var(--border-hover) / <alpha-value>)',
        active: 'rgb(var(--border-active) / <alpha-value>)',
        error: 'rgb(var(--border-error) / <alpha-value>)'
      },
      outlineColor: {
        base: 'rgb(var(--border-base) / <alpha-value>)',
        focus: 'rgb(var(--border-focus) / <alpha-value>)',
        hover: 'rgb(var(--border-hover) / <alpha-value>)',
        active: 'rgb(var(--border-active) / <alpha-value>)',
        error: 'rgb(var(--border-error) / <alpha-value>)'
      },
      divideColor: {
        base: 'rgb(var(--border-base) / <alpha-value>)',
        focus: 'rgb(var(--border-focus) / <alpha-value>)',
        hover: 'rgb(var(--border-hover) / <alpha-value>)',
        active: 'rgb(var(--border-active) / <alpha-value>)',
        error: 'rgb(var(--border-error) / <alpha-value>)'
      }
      // End token colours
    }
  },
  plugins: [require('@tailwindcss/forms'), require('@tailwindcss/typography')]
}
