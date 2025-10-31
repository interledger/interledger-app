const colors = require('tailwindcss/colors')

module.exports = {
  mode: 'jit',
  content: ['./app/**/*.{ts,tsx}'],
  darkMode: 'media',
  theme: {
    fontFamily: {
      display: ['Poppins'],
      sans: ['Inter'],
      mono: ['"Overpass Mono"', '"Source Code Pro"']
    },
    extend: {
      colors: {
        brand: colors.rose[500]
      },
      // Token colours
      textColor: {
        strong: 'rgba(var(--text-strong), <alpha-value>)',
        medium: 'rgba(var(--text-medium), <alpha-value>)',
        weak: 'rgba(var(--text-weak), <alpha-value>)',
        disabled: 'rgba(var(--text-disabled), <alpha-value>)',
        primary: 'rgba(var(--text-primary), <alpha-value>)',
        error: 'rgba(var(--text-error), <alpha-value>)',
        success: 'rgba(var(--text-success), <alpha-value>)'
      },
      backgroundColor: {
        app: 'rgba(var(--bg-app), <alpha-value>)',
        page: 'rgba(var(--bg-page), <alpha-value>)',
        container: 'rgba(var(--bg-container), <alpha-value>)',
        'container-hover': 'rgba(var(--bg-container-hover), <alpha-value>)',
        strong: 'rgba(var(--bg-strong), <alpha-value>)',
        disabled: 'rgba(var(--bg-disabled), <alpha-value>)',
        primary: 'rgba(var(--bg-primary), <alpha-value>)',
        'container-primary': 'rgba(var(--bg-container-primary), <alpha-value>)',
        'container-primary-hover':
          'rgba(var(--bg-container-primary-hover), <alpha-value>)',
        'container-primary-active':
          'rgba(var(--bg-container-primary-active), <alpha-value>)',
        'container-secondary':
          'rgba(var(--bg-container-secondary), <alpha-value>)',
        scrim: 'rgba(var(--bg-scrim), <alpha-value>)',
        snackbar: 'rgba(var(--bg-snackbar), <alpha-value>)',
        footer: 'rgba(var(--bg-footer), <alpha-value>)'
      },
      borderColor: {
        base: 'rgba(var(--border-base), <alpha-value>)',
        focus: 'rgba(var(--border-focus), <alpha-value>)',
        hover: 'rgba(var(--border-hover), <alpha-value>)',
        active: 'rgba(var(--border-active), <alpha-value>)',
        error: 'rgba(var(--border-error), <alpha-value>)'
      },
      ringColor: {
        base: 'rgba(var(--border-base), <alpha-value>)',
        focus: 'rgba(var(--border-focus), <alpha-value>)',
        hover: 'rgba(var(--border-hover), <alpha-value>)',
        active: 'rgba(var(--border-active), <alpha-value>)',
        error: 'rgba(var(--border-error), <alpha-value>)'
      },
      outlineColor: {
        base: 'rgba(var(--border-base), <alpha-value>)',
        focus: 'rgba(var(--border-focus), <alpha-value>)',
        hover: 'rgba(var(--border-hover), <alpha-value>)',
        active: 'rgba(var(--border-active), <alpha-value>)',
        error: 'rgba(var(--border-error), <alpha-value>)'
      },
      divideColor: {
        base: 'rgba(var(--border-base), <alpha-value>)',
        focus: 'rgba(var(--border-focus), <alpha-value>)',
        hover: 'rgba(var(--border-hover), <alpha-value>)',
        active: 'rgba(var(--border-active), <alpha-value>)',
        error: 'rgba(var(--border-error), <alpha-value>)'
      }
    }
  },
  plugins: [require('@tailwindcss/forms'), require('@tailwindcss/typography')]
}
