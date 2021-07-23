import { defineConfig } from 'windicss/helpers';
import typography from 'windicss/plugin/typography';
import colors from 'windicss/colors';

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
      },
      stroke: (theme) => theme('colors'),
      fill: (theme) => theme('colors'),
    },
  },
  plugins: [typography],
});
