import { readFile } from 'fs/promises';

const config = JSON.parse(await readFile(new URL('../.prettierrc.json', import.meta.url)));

export default {
  ...config,
  plugins: ['@ianvs/prettier-plugin-sort-imports'],
};
