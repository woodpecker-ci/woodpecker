# Translations

Woodpecker uses [Vue I18n](https://vue-i18n.intlify.dev/) as translation library, thus you can easily translate the web UI into your language. Therefore, copy the file `web/src/assets/locales/en.json` to the same path with your language's code and `.json` as name.
Then, translate content of this file, but only the values:

```json
{
  "dont_translate": "Only translate this text"
}
```

To add support for time formatting, import the language into two files:

1. `web/src/compositions/useDate.ts`: Just add a line like `import 'dayjs/locale/en';` to the first block of `import` statements and replace `en` with your language's code.
2. `web/src/utils/timeAgo.ts`: Add a line like `import en from 'javascript-time-ago/locale/en.json';` to the other `import`-statements and replace both `en`s with your language's code. Then, add the line `TimeAgo.addDefaultLocale(en);` to the other lines of them, and replace `en` with your language's code.

Then, the web UI should be available in your language. You should open a pull request to our repository to get your changes into the next release.
