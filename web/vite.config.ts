import { defineConfig } from 'vite';
import path from 'path';
import vue from '@vitejs/plugin-vue';
import WindiCSS from 'vite-plugin-windicss';
import Icons from 'vite-plugin-icons';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue(), WindiCSS(), Icons()],
  resolve: {
    alias: {
      '~/': `${path.resolve(__dirname, 'src')}/`,
    },
  },
});
