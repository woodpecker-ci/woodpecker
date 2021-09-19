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
      },
      stroke: (theme) => theme('colors'),
      fill: (theme) => theme('colors'),
    },
  },
  plugins: [typography],
});
