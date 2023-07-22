/* eslint-disable import/no-extraneous-dependencies */
import colors from 'windicss/colors';
import { defineConfig } from 'windicss/helpers';
import typography from 'windicss/plugin/typography';

export default defineConfig({
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Internals to keep a single source for color definitions
        'internal-wp-primary': {
          100: '#8AD97F',
          200: '#68C464',
          300: '#4CAF50',
          400: '#369943',
          500: '#248438',
          600: '#166E30',
        },
        'internal-wp-secondary': {
          200: '#434858',
          300: '#383C4A',
          400: '#303440',
          500: '#2D313D',
          600: '#2A2E3A',
          700: '#222631',
          800: '#1B1F28',
        },
        // Theme colors
        'wp-background': {
          100: 'var(--wp-background-100)',
          200: 'var(--wp-background-200)',
          300: 'var(--wp-background-300)',
          400: 'var(--wp-background-400)',
        },

        'wp-text': {
          100: 'var(--wp-text-100)',
        },
        'wp-text-alt': {
          100: 'var(--wp-text-alt-100)',
        },

        'wp-primary': {
          100: 'var(--wp-primary-100)',
          200: 'var(--wp-primary-200)',
          300: 'var(--wp-primary-300)',
        },
        'wp-primary-text': {
          100: 'var(--wp-primary-text-100)',
        },

        'wp-control-neutral': {
          100: 'var(--wp-control-neutral-100)',
          200: 'var(--wp-control-neutral-200)',
          300: 'var(--wp-control-neutral-300)',
        },
        'wp-control-ok': {
          100: 'var(--wp-control-ok-100)',
          200: 'var(--wp-control-ok-200)',
          300: 'var(--wp-control-ok-300)',
        },

        'wp-pipeline-error': {
          100: 'var(--wp-pipeline-error-100)',
        },
        'wp-pipeline-neutral': {
          100: 'var(--wp-pipeline-neutral-100)',
        },
        'wp-pipeline-ok': {
          100: 'var(--wp-pipeline-ok-100)',
        },
        'wp-pipeline-info': {
          100: 'var(--wp-pipeline-info-100)',
        },

        'wp-code': {
          100: 'var(--wp-code-100)',
          200: 'var(--wp-code-200)',
        },
        'wp-code-text': {
          100: 'var(--wp-code-text-100)',
        },

        'wp-link': {
          100: 'var(--wp-link-100)',
        },
      },
      transitionProperty: {
        height: 'max-height',
      },
      stroke: (theme) => theme('colors'),
      fill: (theme) => theme('colors'),
      fontFamily: [
        'system-ui',
        '-apple-system',
        'Segoe UI',
        'Roboto',
        'Helvetica Neue',
        'Noto Sans',
        'Liberation Sans',
        'Arial',
        'sans-serif',
      ],
    },
  },
  shortcuts: {
    'hover-effect':
      'hover:bg-black hover:bg-opacity-10 dark:hover:bg-white dark:hover:bg-opacity-5 transition-colors duration-100',
  },
  plugins: [typography({})],
});
