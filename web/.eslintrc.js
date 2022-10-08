// @ts-check
/** @type {import('@typescript-eslint/experimental-utils').TSESLint.Linter.Config} */

/* eslint-env node */
module.exports = {
  env: {
    browser: true,
  },
  reportUnusedDisableDirectives: true,

  parser: 'vue-eslint-parser',
  parserOptions: {
    project: ['./tsconfig.eslint.json'],
    tsconfigRootDir: __dirname,
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore see https://github.com/vuejs/vue-eslint-parser#parseroptionsparser
    parser: '@typescript-eslint/parser',
    sourceType: 'module',
    extraFileExtensions: ['.vue'],
  },

  plugins: ['@typescript-eslint', 'import', 'simple-import-sort'],
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'airbnb-base',
    'airbnb-typescript/base',
    'plugin:import/errors',
    'plugin:import/warnings',
    'plugin:import/typescript',
    'plugin:promise/recommended',
    'plugin:vue/vue3-recommended',
    'plugin:prettier/recommended',
    'plugin:vue-scoped-css/recommended',
  ],

  rules: {
    // enable scope analysis rules
    'no-unused-vars': 'off',
    '@typescript-eslint/no-unused-vars': 'error',
    'no-use-before-define': 'off',
    '@typescript-eslint/no-use-before-define': 'error',
    'no-shadow': 'off',
    '@typescript-eslint/no-shadow': 'error',
    'no-redeclare': 'off',
    '@typescript-eslint/no-redeclare': 'error',

    // make typescript eslint rules even more strict
    '@typescript-eslint/no-explicit-any': 'error',
    '@typescript-eslint/explicit-module-boundary-types': 'off',
    '@typescript-eslint/no-non-null-assertion': 'error',

    'import/no-unresolved': 'off', // disable as this is handled by tsc itself
    'import/first': 'error',
    'import/newline-after-import': 'error',
    'import/no-cycle': 'error',
    'import/no-relative-parent-imports': 'error',
    'import/no-duplicates': 'error',
    'import/no-extraneous-dependencies': 'error',
    'import/extensions': 'off',
    'import/prefer-default-export': 'off',

    'simple-import-sort/imports': 'error',
    'simple-import-sort/exports': 'error',

    'promise/prefer-await-to-then': 'error',
    'promise/prefer-await-to-callbacks': 'error',

    'no-underscore-dangle': 'off',
    'no-else-return': ['error', { allowElseIf: false }],
    'no-return-assign': ['error', 'always'],
    'no-return-await': 'error',
    'no-useless-return': 'error',
    'no-restricted-imports': [
      'error',
      {
        patterns: ['src', 'dist'],
      },
    ],
    'no-console': 'warn',
    'no-useless-concat': 'error',
    'prefer-const': 'error',
    'spaced-comment': ['error', 'always'],
    'object-shorthand': ['error', 'always'],
    'no-useless-rename': 'error',
    eqeqeq: 'error',

    'vue/attribute-hyphenation': 'error',
    // enable in accordance with https://github.com/prettier/eslint-config-prettier#vuehtml-self-closing
    'vue/html-self-closing': [
      'error',
      {
        html: {
          void: 'any',
        },
      },
    ],
    'vue/no-static-inline-styles': 'error',
    'vue/v-on-function-call': 'error',
    'vue/no-useless-v-bind': 'error',
    'vue/no-useless-mustaches': 'error',
    'vue/no-useless-concat': 'error',
    'vue/no-boolean-default': 'error',
    'vue/html-button-has-type': 'error',
    'vue/component-name-in-template-casing': 'error',
    'vue/match-component-file-name': [
      'error',
      {
        extensions: ['vue'],
        shouldMatchCase: true,
      },
    ],
    'vue/require-name-property': 'error',
    'vue/v-for-delimiter-style': 'error',
    'vue/no-empty-component-block': 'error',
    'vue/no-duplicate-attr-inheritance': 'error',
    'vue/no-unused-properties': [
      'error',
      {
        groups: ['props', 'data', 'computed', 'methods', 'setup'],
      },
    ],
    'vue/new-line-between-multi-line-property': 'error',
    'vue/padding-line-between-blocks': 'error',
    'vue/multi-word-component-names': 'off',
    'vue/no-reserved-component-names': 'off',

    // css rules
    'vue-scoped-css/no-unused-selector': 'error',
    'vue-scoped-css/no-parsing-error': 'error',
    'vue-scoped-css/require-scoped': 'error',

    // enable in accordance with https://github.com/prettier/eslint-config-prettier#curly
    curly: ['error', 'all'],

    // risky because of https://github.com/prettier/eslint-plugin-prettier#arrow-body-style-and-prefer-arrow-callback-issue
    'arrow-body-style': 'error',
    'prefer-arrow-callback': 'error',
  },
};
