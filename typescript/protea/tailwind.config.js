const defaultTheme = require('tailwindcss/defaultTheme')
const colors = require('tailwindcss/colors')

// This function ensures tailwind opacity utilities work with tokens.
function withOpacity(variableName) {
  return ({ opacityValue }) => {
    if (opacityValue !== undefined) {
      return `rgba(var(${variableName}), ${opacityValue})`
    }
    return `rgb(var(${variableName}))`
  }
}

module.exports = {
  mode: 'jit',
  purge: ['./pages/**/*.{ts,tsx}', './components/**/*.{ts,tsx}'],
  darkMode: 'media',
  theme: {
    fontFamily: {
      display: ['Poppins'],
      sans: ['Inter'],
      mono: ['"Overpass Mono"', '"Source Code Pro"']
    },
    colors: {
      transparent: 'transparent',
      current: 'currentColor',
      white: colors.white,
      black: colors.black,
      gray: colors.gray,
      red: colors.red,
      orange: colors.orange,
      yellow: colors.yellow,
      green: colors.emerald,
      sky: colors.sky,
      blue: colors.blue,
      indigo: colors.indigo,
      purple: colors.purple,
      pink: colors.pink
    },
    extend: {
      colors: {
        primary: '#F35167',
        secondary: '#7DD043',
        'secondary-dark': '#4F971C',

        // for syntax highlighting TODO: Update colors to match palette
        teal: colors.cyan,
        fuchsia: colors.fuchsia,
        lime: colors.lime,
        sky: colors.sky,
        rose: colors.rose,
        emerald: colors.emerald
      },
      // Token colours
      textColor: {
        strong: withOpacity('--text-strong'),
        medium: withOpacity('--text-medium'),
        weak: withOpacity('--text-weak'),
        disabled: withOpacity('--text-disabled'),
        primary: withOpacity('--text-primary'),
        error: withOpacity('--text-error')
      },
      backgroundColor: {
        container: withOpacity('--bg-container'),
        'container-hover': withOpacity('--bg-container-hover'),
        strong: withOpacity('--bg-strong'),
        disabled: withOpacity('--bg-disabled'),
        primary: withOpacity('--bg-primary'),
        'container-primary': withOpacity('--bg-container-primary'),
        'container-primary-hover': withOpacity('--bg-container-primary-hover'),
        'container-primary-active': withOpacity('--bg-container-primary-active')
      },
      borderColor: {
        base: withOpacity('--border-base'),
        focus: withOpacity('--border-focus'),
        hover: withOpacity('--border-hover'),
        active: withOpacity('--border-active'),
        error: withOpacity('--border-error')
      },
      ringColor: {
        base: withOpacity('--border-base'),
        focus: withOpacity('--border-focus'),
        hover: withOpacity('--border-hover'),
        active: withOpacity('--border-active'),
        error: withOpacity('--border-error')
      },
      divideColor: {
        base: withOpacity('--border-base'),
        focus: withOpacity('--border-focus'),
        hover: withOpacity('--border-hover'),
        active: withOpacity('--border-active'),
        error: withOpacity('--border-error')
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
  variants: {
    extend: {
      textColor: ['selection'],
      backgroundColor: ['selection'],
      ringColor: ['focus-visible']
    }
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
    require('tailwindcss-selection-variant')
  ]
}
