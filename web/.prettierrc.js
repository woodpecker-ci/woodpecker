import { readFile } from 'node:fs/promises';

const config = JSON.parse(await readFile(new URL('../.prettierrc.json', import.meta.url)));

export default {
  ...config,
  plugins: ['@ianvs/prettier-plugin-sort-imports'],
  importOrder: [
    '<THIRD_PARTY_MODULES>', // Imports not matched by other special words or groups.
    '', // Empty string will match any import not matched by other special words or groups.
    '^(#|@|~|\\$)(/.*)$',
    '',
    '^[./]',
  ],
};
