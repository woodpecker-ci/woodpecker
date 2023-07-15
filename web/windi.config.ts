/* eslint-disable import/no-extraneous-dependencies */
import colors from 'windicss/colors';
import { defineConfig } from 'windicss/helpers';
import typography from 'windicss/plugin/typography';

export default defineConfig({
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        greenish: '#4caf50',
        woodpecker: {
          100: '#BDFFAC',
          200: '#7FD777',
          300: '#4CAF50',
          400: '#2A8737',
          500: '#115F24',
          600: '#043818',
        },
        link: colors.blue[400],
        'dark-gray': {
          600: '#383c4a',
          700: '#303440',
          800: '#2a2e3a',
          900: '#2e323e',
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
  plugins: [typography],
});
