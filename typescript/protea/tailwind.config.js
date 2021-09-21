const defaultTheme = require('tailwindcss/defaultTheme')
const colors = require('tailwindcss/colors')

module.exports = {
  mode: 'jit',
  purge: ['./pages/**/*.{ts,tsx}', './components/**/*.{ts,tsx}'],
  darkMode: 'media',
  theme: {
    fontFamily: {
      display: ['Poppins'],
      sans: ['Inter'],
      mono: ['"Overpass Mono"', '"Source Code Pro"'],
      icon: ['"Material Icons Sharp"']
    },
    extend: {
      colors: {
        primary: '#F35167',
        secondary: '#7DD043',
        'secondary-dark': '#4F971C',
        gray: colors.coolGray,

        teal: colors.cyan,
        // for syntax highlighting
        fuchsia: colors.fuchsia,
        lime: colors.lime,
        sky: colors.sky,
        rose: colors.rose,
        emerald: colors.emerald
      },
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
        },
        dark: {
          css: [
            {
              h1: {
                fontFamily: theme('fontFamily.display'),
                fontWeight: theme('fontWeight.medium'),
                color: theme('colors.gray.50')
              },
              h2: {
                fontFamily: theme('fontFamily.display'),
                fontWeight: theme('fontWeight.medium'),
                color: theme('colors.gray.50')
              },
              h3: {
                fontFamily: theme('fontFamily.display'),
                fontWeight: theme('fontWeight.medium'),
                color: theme('colors.gray.50')
              },
              h4: {
                fontFamily: theme('fontFamily.display'),
                fontWeight: theme('fontWeight.medium'),
                color: theme('colors.gray.50')
              },
              color: theme('colors.gray.50'),
              '[class~="lead"]': {
                color: theme('colors.gray.300')
              },
              a: {
                color: theme('colors.secondary')
              },
              strong: {
                color: theme('colors.white')
              },
              'ol > li::before': {
                color: theme('colors.gray.400')
              },
              'ul > li::before': {
                backgroundColor: theme('colors.gray.600')
              },
              hr: {
                borderColor: theme('colors.gray.200')
              },
              blockquote: {
                fontWeight: 400,
                color: theme('colors.gray.200'),
                borderLeftColor: theme('colors.gray.600')
              },
              'blockquote p:first-of-type::before': {
                content: ''
              },
              'blockquote p:first-of-type::after': {
                content: ''
              },
              'figure figcaption': {
                color: theme('colors.gray.50')
              },
              code: {
                color: theme('colors.gray.50'),
                fontWeight: '500'
              },
              'a code': {
                color: theme('colors.secondary')
              },
              pre: {
                color: theme('colors.gray.50'),
                backgroundColor: theme('colors.gray.800'),
                borderRadius: 0
              },
              thead: {
                color: theme('colors.white'),
                borderBottomColor: theme('colors.gray.400')
              },
              'tbody tr': {
                borderBottomColor: theme('colors.gray.600')
              }
            }
          ]
        }
      })
    }
  },
  variants: {
    extend: {
      typography: ['dark'],
      textColor: ['selection'],
      backgroundColor: ['selection']
    }
  },
  plugins: [
    require('@tailwindcss/typography'),
    require('tailwindcss-selection-variant')
  ]
}
