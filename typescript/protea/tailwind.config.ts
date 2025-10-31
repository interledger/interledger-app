import type { Config } from 'tailwindcss'
import colors from 'tailwindcss/colors'
// const colors = require('tailwindcss/colors')

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
    fontSize: {
      xs: ['0.75rem', { lineHeight: '1rem' }],
      sm: ['0.875rem', { lineHeight: '1.25rem' }],
      base: ['1rem', { lineHeight: '1.5rem' }],
      lg: ['1.125rem', { lineHeight: '1.75rem' }],
      xl: ['1.25rem', { lineHeight: '1.75rem' }],
      '2xl': ['1.5rem', { lineHeight: '2rem' }],
      '3xl': ['1.875rem', { lineHeight: '2.25rem' }],
      '4xl': ['2.25rem', { lineHeight: '2.6875rem' }],
      '5xl': ['3rem', { lineHeight: '3.5625rem' }],
      '6xl': ['3.75rem', { lineHeight: '4.5rem' }],
      '7xl': ['4.5rem', { lineHeight: '15.375rem' }],
      '8xl': ['6rem', { lineHeight: '7.1875rem' }],
      '9xl': ['8rem', { lineHeight: '9.5625rem' }]
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
        success: 'rgb(var(--text-success) / <alpha-value>)',
        'on-color': 'rgb(var(--text-on-color) / <alpha-value>)',
        inverted: 'rgb(var(--text-inverted) / <alpha-value>)',
        'chip-red': 'var(--text-chip-red)',
        'chip-sky': 'var(--text-chip-sky)',
        'chip-green': 'var(--text-chip-green)',
        'chip-orange': 'var(--text-chip-orange)',
        'chip-blue': 'var(--text-chip-blue)',
        'chip-indigo': 'var(--text-chip-indigo)',
        'chip-purple': 'var(--text-chip-purple)',
        'chip-yellow': 'var(--text-chip-yellow)',
        'chip-lime': 'var(--text-chip-lime)',
        'chip-rose': 'var(--text-chip-rose)',
        'chip-slate': 'var(--text-chip-slate)',
        charcoal: 'rgb(var(--charcoal) / <alpha-value>)'
      },
      ringOffsetColor: {
        'container-strong': 'rgb(var(--bg-container-strong) / <alpha-value>)'
      },
      backgroundColor: {
        app: 'rgb(var(--bg-app) / <alpha-value>)',
        page: 'rgb(var(--bg-page) / <alpha-value>)',
        container: 'rgb(var(--bg-container) / <alpha-value>)',
        'container-strong': 'rgb(var(--bg-container-strong) / <alpha-value>)',
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
        nav: 'rgb(var(--bg-nav) / <alpha-value>)',
        error: 'rgb(var(--bg-error) / <alpha-value>)',
        'error-hover': 'rgb(var(--bg-error-hover) / <alpha-value>)',
        'nav-active': 'rgb(var(--bg-nav-active) / <alpha-value>)',
        'nav-hover': 'rgb(var(--bg-nav-hover) / <alpha-value>)',
        'nav-disabled': 'rgb(var(--bg-nav-disabled) / <alpha-value>)',
        'mk-footer': 'rgb(var(--mk-bg-footer) / <alpha-value>)',
        'mk-page': 'rgb(var(--mk-bg-page) / <alpha-value>)',
        'mk-section': 'rgb(var(--mk-bg-section) / <alpha-value>)',
        'mk-section-hover': 'rgb(var(--mk-bg-section-hover) / <alpha-value>)',
        'chip-red': 'var(--bg-chip-red)',
        'chip-sky': 'var(--bg-chip-sky)',
        'chip-green': 'var(--bg-chip-green)',
        'chip-orange': 'var(--bg-chip-orange)',
        'chip-blue': 'var(--bg-chip-blue)',
        'chip-indigo': 'var(--bg-chip-indigo)',
        'chip-purple': 'var(--bg-chip-purple)',
        'chip-yellow': 'var(--bg-chip-yellow)',
        'chip-lime': 'var(--bg-chip-lime)',
        'chip-rose': 'var(--bg-chip-rose)',
        'chip-slate': 'var(--bg-chip-slate)',
        'alert-slate': 'var(--bg-alert-slate)',
        'alert-success': 'var(--bg-alert-success)',
        'mint-light': 'rgb(var(--mint-light) / <alpha-value>)',
        'mint-dark': 'rgb(var(--mint-dark) / <alpha-value>)',
        'teal-deep': 'rgb(var(--teal-deep) / <alpha-value>)'
      },
      borderColor: {
        base: 'rgb(var(--border) / <alpha-value>)',
        focus: 'rgb(var(--border-focus) / <alpha-value>)',
        hover: 'rgb(var(--border-hover) / <alpha-value>)',
        active: 'rgb(var(--border-active) / <alpha-value>)',
        error: 'rgb(var(--border-error) / <alpha-value>)'
      },
      ringColor: {
        base: 'rgb(var(--border) / <alpha-value>)',
        focus: 'rgb(var(--border-focus) / <alpha-value>)',
        hover: 'rgb(var(--border-hover) / <alpha-value>)',
        active: 'rgb(var(--border-active) / <alpha-value>)',
        error: 'rgb(var(--border-error) / <alpha-value>)'
      },
      outlineColor: {
        base: 'rgb(var(--border) / <alpha-value>)',
        focus: 'rgb(var(--border-focus) / <alpha-value>)',
        hover: 'rgb(var(--border-hover) / <alpha-value>)',
        active: 'rgb(var(--border-active) / <alpha-value>)',
        error: 'rgb(var(--border-error) / <alpha-value>)'
      },
      divideColor: {
        base: 'rgb(var(--border) / <alpha-value>)',
        focus: 'rgb(var(--border-focus) / <alpha-value>)',
        hover: 'rgb(var(--border-hover) / <alpha-value>)',
        active: 'rgb(var(--border-active) / <alpha-value>)',
        error: 'rgb(var(--border-error) / <alpha-value>)'
      },
      gradientColorStops: {
        'mint-light': 'rgb(var(--mint-light) / <alpha-value>)',
        'mint-dark': 'rgb(var(--mint-dark) / <alpha-value>)'
      }
      // End token colours
    }
  },
  plugins: [require('@tailwindcss/forms'), require('@tailwindcss/typography')]
} satisfies Config
