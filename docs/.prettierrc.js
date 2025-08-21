const config = require('../.prettierrc.json');

module.exports = {
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
