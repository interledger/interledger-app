import js from '@eslint/js'
import prettier from 'eslint-config-prettier'
import jsxA11y from 'eslint-plugin-jsx-a11y'
import react from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'
import tseslint from 'typescript-eslint'

export default tseslint.config(
  {
    ignores: [
      'build/',
      '.react-router/',
      'app/generated/',
      'public/',
      'coverage/',
      'remix.env.d.ts'
    ]
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  react.configs.flat.recommended,
  react.configs.flat['jsx-runtime'],
  jsxA11y.flatConfigs.recommended,
  {
    files: ['**/*.{js,jsx,ts,tsx}'],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module'
    },
    settings: { react: { version: 'detect' } },
    plugins: { 'react-hooks': reactHooks },
    rules: {
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',
      // low-value style debt (separated let/assignment) — burn down, then promote.
      'prefer-const': 'warn',
      // catch unused vars/imports;
      // ignore intentional unused function args and `catch` bindings (selectors, loaders, etc.).
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          args: 'none',
          varsIgnorePattern: '^_',
          caughtErrors: 'none',
          ignoreRestSiblings: true
        }
      ],
      // autoFocus is used intentionally in some modals/popups — track, don't block.
      'jsx-a11y/no-autofocus': 'warn',
      // pre-existing tech debt: tracked as warnings to burn down,
      // then promote to 'error' once cleared. Not blocking CI today.
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/ban-ts-comment': 'warn',
      'react/no-unescaped-entities': 'warn',
      // clickable <div>s in admin actions that should become <button>s — burn down, then promote.
      'jsx-a11y/click-events-have-key-events': 'warn',
      'jsx-a11y/no-static-element-interactions': 'warn'
    }
  },
  {
    // tailwind build/config files legitimately use CommonJS require()/module.exports
    // and run in a Node context. (protea's equivalent is a .ts file, where
    // typescript-eslint disables no-undef for us.)
    files: ['**/*.config.{js,ts}'],
    languageOptions: { globals: { require: 'readonly', module: 'readonly' } },
    rules: { '@typescript-eslint/no-require-imports': 'off' }
  },
  prettier
)
