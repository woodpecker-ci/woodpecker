// cSpell:ignore tseslint
// @ts-check

import antfu from '@antfu/eslint-config';
import js from '@eslint/js';
import vueI18n from '@intlify/eslint-plugin-vue-i18n';
import eslintPromise from 'eslint-plugin-promise';
import eslintPluginVueScopedCSS from 'eslint-plugin-vue-scoped-css';

export default antfu(
  {
    stylistic: false,
    typescript: {
      tsconfigPath: './tsconfig.json',
    },
    vue: true,

    // Disable jsonc and yaml support
    jsonc: false,
    yaml: false,
  },

  js.configs.recommended,
  eslintPromise.configs['flat/recommended'],
  ...eslintPluginVueScopedCSS.configs['flat/recommended'],
  ...vueI18n.configs['flat/recommended'],

  {
    rules: {
      'import/order': 'off',
      'sort-imports': 'off',
      'promise/prefer-await-to-callbacks': 'error',

      // Vue I18n
      '@intlify/vue-i18n/no-raw-text': [
        'error',
        {
          attributes: {
            '/.+/': ['label'],
          },
        },
      ],
      '@intlify/vue-i18n/key-format-style': ['error', 'snake_case'],
      '@intlify/vue-i18n/no-duplicate-keys-in-locale': 'error',
      '@intlify/vue-i18n/no-dynamic-keys': 'error',
      '@intlify/vue-i18n/no-deprecated-i18n-component': 'error',
      '@intlify/vue-i18n/no-deprecated-tc': 'error',
      '@intlify/vue-i18n/no-i18n-t-path-prop': 'error',
      '@intlify/vue-i18n/no-missing-keys-in-other-locales': 'off',
      '@intlify/vue-i18n/valid-message-syntax': 'error',
      '@intlify/vue-i18n/no-missing-keys': 'error',
      '@intlify/vue-i18n/no-unknown-locale': 'error',
      '@intlify/vue-i18n/no-unused-keys': ['error', { extensions: ['.ts', '.vue'] }],
      '@intlify/vue-i18n/prefer-sfc-lang-attr': 'error',
      '@intlify/vue-i18n/no-html-messages': 'error',
      '@intlify/vue-i18n/prefer-linked-key-with-paren': 'error',
      '@intlify/vue-i18n/sfc-locale-attr': 'error',
    },
    settings: {
      // Vue I18n
      'vue-i18n': {
        localeDir: './src/assets/locales/en.json',
        // Specify the version of `vue-i18n` you are using.
        // If not specified, the message will be parsed twice.
        messageSyntaxVersion: '^9.0.0',
      },
    },
  },

  // Vue
  {
    files: ['**/*.vue'],
    rules: {
      'vue/multi-word-component-names': 'off',
      'vue/html-self-closing': [
        'error',
        {
          html: {
            void: 'always',
            normal: 'always',
            component: 'always',
          },
          svg: 'always',
          math: 'always',
        },
      ],
      'vue/block-order': [
        'error',
        {
          order: ['template', 'script', 'style'],
        },
      ],
      'vue/singleline-html-element-content-newline': ['off'],
    },
  },

  // Ignore list
  {
    ignores: [
      'dist',
      'coverage/',
      'package.json',
      'tsconfig.eslint.json',
      'tsconfig.json',
      'src/assets/locales/**/*',
      '!src/assets/locales/en.json',
      'src/assets/dayjsLocales/',
      'components.d.ts',
    ],
  },
);
