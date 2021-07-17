import colors from 'windicss/colors';
import { defineConfig } from 'windicss/helpers';
import typography from 'windicss/plugin/typography';

export default defineConfig({
  darkMode: 'class',
  theme: {
    colors: {
      green: '#4caf50',
    },
  },
  plugins: [typography],
});
