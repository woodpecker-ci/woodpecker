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
    },
  },
  plugins: [typography],
});
