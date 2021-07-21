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
        'status-blocked': colors.red[400],
        'status-declined': colors.red[400],
        'status-error': colors.red[400],
        'status-failure': colors.red[400],
        'status-killed': colors.red[400],
        'status-pending': colors.gray[400],
        'status-running': colors.blue[400],
        'status-skipped': colors.gray[400],
        'status-started': colors.blue[400],
        'status-success': '#4caf50',
      },
      stroke: (theme) => theme('colors'),
      fill: (theme) => theme('colors'),
    },
  },
  plugins: [typography],
});
