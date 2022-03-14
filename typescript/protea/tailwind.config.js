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
        primary: '#F35167',
        secondary: '#7DD043'
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
        'container-primary-active': withOpacity(
          '--bg-container-primary-active'
        ),
        snackbar: withOpacity('--bg-snackbar')
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
  plugins: [require('@tailwindcss/forms'), require('@tailwindcss/typography')]
}
