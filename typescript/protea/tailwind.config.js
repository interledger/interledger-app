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
      },
      // End token colours
      typography: (theme) => ({
        DEFAULT: {
          css: [
            {
              h1: {
                fontFamily: theme('fontFamily.display'),
                fontWeight: theme('fontWeight.medium')
              },
              h2: {
                fontFamily: theme('fontFamily.display'),
                fontWeight: theme('fontWeight.medium')
              },
              h3: {
                fontFamily: theme('fontFamily.display'),
                fontWeight: theme('fontWeight.medium')
              },
              h4: {
                fontFamily: theme('fontFamily.display'),
                fontWeight: theme('fontWeight.medium')
              },
              color: theme('colors.black'),
              pre: {
                backgroundColor: theme('colors.gray.50'),
                borderRadius: 0,
                color: theme('colors.black'),
                padding: '2rem'
              },
              code: {
                fontWeight: '500'
              },
              'a code': {
                color: theme('colors.primary')
              },
              a: {
                color: theme('colors.primary')
              },
              blockquote: {
                fontWeight: 400
              },
              'blockquote p:first-of-type::before': {
                content: ''
              },
              'blockquote p:first-of-type::after': {
                content: ''
              }
            }
          ]
        }
      })
    }
  },
  plugins: [require('@tailwindcss/forms'), require('@tailwindcss/typography')]
}
