import { readFile } from 'node:fs/promises';

// eslint-disable-next-line antfu/no-top-level-await
const config = JSON.parse(await readFile(new URL('../.prettierrc.json', import.meta.url)));

export default {
  ...config,
  plugins: ['prettier-plugin-tailwindcss'],
};
