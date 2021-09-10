/* eslint-env node */

module.exports = {
  parserOptions: {
    project: ['./tsconfig.eslint.json'],
    tsconfigRootDir: __dirname,
  },

  extends: [
    "standard",
    "plugin:jest/recommended",
    "prettier",
  ],
  plugins: ["jest", "prettier"],
  parser: "babel-eslint",
  parserOptions: {
    ecmaVersion: 2016,
    sourceType: "module",
    ecmaFeatures: {
      jsx: true
    }
  },
  env: {
    es6: true,
    browser: true,
    node: true,
    "jest/globals": true
  },
  rules: {
    "prettier/prettier": [
      "error",
      {
        trailingComma: "all",
      }
    ]
  }
};
