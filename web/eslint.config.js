// cSpell:ignore TSES
// @ts-check

import simpleImportSort from 'eslint-plugin-simple-import-sort';
//import eslintImport from 'eslint-plugin-import';
//import eslintTypescript from '@typescript-eslint/eslint-plugin';
import vueParser from 'vue-eslint-parser';
import globals from 'globals';
import js from '@eslint/js';
import airBnbBase from 'eslint-config-airbnb-base';
//import airBnbTS from 'eslint-config-airbnb-typescript/base';
import eslintVue from 'eslint-plugin-vue';
import eslintVueScopedCSS from 'eslint-plugin-vue-scoped-css';
import eslintPrettier from 'eslint-plugin-prettier/recommended';
import eslintPromise from 'eslint-plugin-promise';
import tseslint from 'typescript-eslint';

import path from 'path';
import { fileURLToPath } from 'url';

// TODO check eslint-env
/* eslint-env node */

export default tseslint.config(
  {
    ignores: [
      'dist/**',
      'package.json',
      'tsconfig.eslint.json',
      'tsconfig.json',
      'src/assets/locales/**',
      'src/assets/dayjsLocales/**',
      'components.d.ts',
    ],
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,

  //airBnbBase,
  //airBnbTS,
  //eslintImport, used these 'plugin:import/errors', 'plugin:import/warnings', 'plugin:import/typescript'
  ...eslintVue.configs['flat/recommended'],
  ...eslintVueScopedCSS.configs['flat/recommended'],
  eslintPrettier,
  //eslintPromise.configs.recommended,
  {
    files: ['**/*.js', '**/*.ts', '**/*.vue'],

    languageOptions: {
      parser: vueParser,
      parserOptions: {
        project: ['./tsconfig.eslint.json'],
        tsconfigRootDir: path.dirname(fileURLToPath(import.meta.url)),
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore see https://github.com/vuejs/vue-eslint-parser#parseroptionsparser
        parser: tseslint.parser,
        extraFileExtensions: ['.vue', '.json'],
      },
      sourceType: 'module',
      globals: globals.browser,
    },

    linterOptions: {
      reportUnusedDisableDirectives: 'warn',
    },

    plugins: {
      //TODO 'import': eslintImport,
      'simple-import-sort': simpleImportSort,
      //TODO promise: eslintPromise,
      vue: eslintVue,
      'vue-scoped-css': eslintVueScopedCSS,
    },

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
      // SOURCE: https://github.com/iamturns/eslint-config-airbnb-typescript/blob/4aec5702be5b4e74e0e2f40bc78b4bc961681de1/lib/shared.js#L41
      '@typescript-eslint/naming-convention': [
        'error',
        // Allow camelCase variables (23.2), PascalCase variables (23.8), and UPPER_CASE variables (23.10)
        {
          selector: 'variable',
          format: ['camelCase', 'PascalCase', 'UPPER_CASE'],
          leadingUnderscore: 'allow',
        },
        // Allow camelCase functions (23.2), and PascalCase functions (23.8)
        {
          selector: 'function',
          format: ['camelCase', 'PascalCase'],
        },
        // Airbnb recommends PascalCase for classes (23.3), and although Airbnb does not make TypeScript recommendations, we are assuming this rule would similarly apply to anything "type like", including interfaces, type aliases, and enums
        {
          selector: 'typeLike',
          format: ['PascalCase'],
        },
      ],

      //'import/no-unresolved': 'off', // disable as this is handled by tsc itself
      //'import/first': 'error',
      //'import/newline-after-import': 'error',
      //'import/no-cycle': 'error',
      //'import/no-relative-parent-imports': 'error',
      //'import/no-duplicates': 'error',
      //'import/no-extraneous-dependencies': 'error',
      //'import/extensions': 'off',
      //'import/prefer-default-export': 'off',

      'simple-import-sort/imports': 'error',
      'simple-import-sort/exports': 'error',

      //'promise/prefer-await-to-then': 'error',
      //'promise/prefer-await-to-callbacks': 'error',

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
  },
);
