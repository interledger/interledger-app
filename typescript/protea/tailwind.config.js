const defaultTheme = require('tailwindcss/defaultTheme')
const colors = require('tailwindcss/colors')

module.exports = {
  mode: 'jit',
  purge: ['./pages/**/*.{ts,tsx}', './components/**/*.{ts,tsx}'],
  darkMode: 'media',
  theme: {
    fontFamily: {
      display: ['Poppins'],
      body: ['Inter'],
      mono: ['"DM Mono"'],
      icon: ['"Material Icons Sharp"']
    },
    extend: {
      colors: {
        primary: '#F35167',
        gray: colors.coolGray
      }
    }
  },
  variants: {
    extend: {}
  },
  plugins: []
}
