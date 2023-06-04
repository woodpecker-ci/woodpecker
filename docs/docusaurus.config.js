const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');
const path = require('path');

/** @type {import('@docusaurus/types').Config} */
module.exports = {
  title: 'Woodpecker CI',
  tagline: 'Woodpecker is a simple CI engine with great extensibility.',
  url: 'https://woodpecker-ci.org',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'throw',
  onDuplicateRoutes: 'throw',
  organizationName: 'woodpecker-ci',
  projectName: 'woodpecker-ci.github.io',
  trailingSlash: false,
  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      navbar: {
        title: 'Woodpecker',
        logo: {
          alt: 'Woodpecker Logo',
          src: 'img/logo.svg',
        },
        items: [
          {
            type: 'doc',
            docId: 'intro',
            activeBaseRegex: 'docs/(?!migrations|awesome)',
            position: 'left',
            label: 'Docs',
          },
          {
            to: '/plugins',
            position: 'left',
            label: 'Plugins',
          },
          {
            to: '/docs/next/migrations', // Always point to newest migration guide
            activeBaseRegex: 'docs/(next/)?migrations',
            position: 'left',
            label: 'Migrations',
          },
          {
            to: '/faq',
            position: 'left',
            label: 'FAQ',
          },
          {
            to: '/docs/next/awesome', // Always point to newest awesome list
            activeBaseRegex: 'docs/(next/)?awesome',
            position: 'left',
            label: 'Awesome',
          },
          {
            to: '/api',
            position: 'left',
            label: 'API',
          },
          {
            type: 'docsVersionDropdown',
            position: 'right',
          },
          {
            href: 'https://github.com/woodpecker-ci/woodpecker',
            position: 'right',
            className: 'header-github-link',
            'aria-label': 'GitHub repository',
          },
          {
            label: '🧡 Sponsor Us',
            position: 'right',
            href: 'https://opencollective.com/woodpecker-ci',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              {
                label: 'Introduction',
                to: '/docs/intro',
              },
              {
                label: 'Usage',
                to: '/docs/usage/intro',
              },
              {
                label: 'Server setup',
                to: '/docs/administration/setup',
              },
              {
                label: 'FAQ',
                to: '/faq',
              },
            ],
          },
          {
            title: 'Community',
            items: [
              {
                label: 'Discord',
                href: 'https://discord.gg/fcMQqSMXJy',
              },
              {
                label: 'Matrix',
                href: 'https://matrix.to/#/#woodpecker:matrix.org',
              },
              {
                label: 'Mastodon',
                href: 'https://floss.social/@WoodpeckerCI',
              },
            ],
          },
          {
            title: 'More',
            items: [
              {
                label: 'Translate',
                href: 'https://translate.woodpecker-ci.org/engage/woodpecker-ci/',
              },
              {
                label: 'GitHub',
                href: 'https://github.com/woodpecker-ci/woodpecker',
              },
              {
                href: 'https://ci.woodpecker-ci.org/woodpecker-ci/woodpecker',
                label: 'CI',
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} Woodpecker CI. Built with Docusaurus.`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
      announcementBar: {
        id: 'github-star',
        content: ` If you like Woodpecker-CI, <a href=https://github.com/woodpecker-ci/woodpecker rel="noopener noreferrer" target="_blank">give us a star on GitHub</a> ! ⭐️`,
        backgroundColor: 'var(--ifm-color-primary)',
        textColor: 'var(--ifm-color-gray-900)',
      },
      tableOfContents: {
        minHeadingLevel: 2,
        maxHeadingLevel: 4,
      },
      colorMode: {
        respectPrefersColorScheme: true,
      },
    }),
  plugins: [
    () => ({
      name: 'docusaurus-plugin-favicon',
      injectHtmlTags() {
        return {
          headTags: [
            {
              tagName: 'link',
              attributes: {
                rel: 'icon',
                href: '/img/favicon.ico',
                sizes: 'any',
              },
            },
            {
              tagName: 'link',
              attributes: {
                rel: 'icon',
                href: '/img/favicon.svg',
                type: 'image/svg+xml',
              },
            },
          ],
        };
      },
    }),
    () => ({
      name: 'webpack-config',
      configureWebpack() {
        return {
          devServer: {
            client: {
              webSocketURL: 'auto://0.0.0.0:0/ws',
            },
          },
        };
      },
    }),
  ],
  themes: [
    path.resolve(__dirname, 'plugins', 'woodpecker-plugins', 'dist'),
    [
      require.resolve('@easyops-cn/docusaurus-search-local'),
      {
        hashed: true,
      },
    ],
  ],
  presets: [
    [
      '@docusaurus/preset-classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: 'https://github.com/woodpecker-ci/woodpecker/edit/master/docs/',
          includeCurrentVersion: true,
          lastVersion: '0.15',
          versions: {
            current: {
              label: 'Next',
              banner: 'unreleased',
            },
            0.15: {
              label: '0.15.x',
              banner: 'none',
            },
          },
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      }),
    ],
    [
      'redocusaurus',
      {
        // Plugin Options for loading OpenAPI files
        specs: [
          {
            spec: 'swagger.json',
            route: '/api/',
          },
        ],
        // Theme Options for modifying how redoc renders them
        theme: {
          // Change with your site colors
          primaryColor: '#1890ff',
        },
      },
    ],
  ],
  webpack: {
    jsLoader: (isServer) => ({
      loader: require.resolve('esbuild-loader'),
      options: {
        loader: 'tsx',
        target: isServer ? 'node12' : 'es2017',
      },
    }),
  },
};
