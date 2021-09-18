/* eslint-disable import/no-extraneous-dependencies */
import daisyColors from 'daisyui/colors/index.js';
import colors from 'windicss/colors';
import { defineConfig } from 'windicss/helpers';
import typography from 'windicss/plugin/typography';

export default defineConfig({
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        green: '#4caf50',
        link: colors.blue[400],

        // status colors
        'status-red': colors.red[400],
        'status-gray': colors.gray[400],
        'status-blue': colors.blue[400],
        'status-green': '#4caf50',

        ...daisyColors,
      },
      stroke: (theme) => theme('colors'),
      fill: (theme) => theme('colors'),
    },
  },
  plugins: [typography],
});
