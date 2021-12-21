# Documentation

The documentation is using docusaurus as framework. You can learn more about it from its [official documentation](https://docusaurus.io/docs/).

If you only want to change some text it probably is enough if you just search for the corresponding [Markdown](https://www.markdownguide.org/basic-syntax/) file inside the `docs/docs/` folder and adjust it. If you want to change larger parts and test the rendered documentation you can run docusaurus locally. Similarly to the UI you need to install [Node.js and Yarn](/docs/development/getting-started#nodejs--yarn). After that you can run and build docusaurus locally by using the following commands:

```bash
cd docs/

yarn install

# build plugins used by the docs
yarn build:woodpecker-plugins

# start docs with hot-reloading, so you can change the docs and directly see the changes in the browser without reloading it manually
yarn start

# or build the docs to deploy it to some static page hosting
yarn build
```
