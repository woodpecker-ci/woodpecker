const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').DocusaurusConfig} */
module.exports = {
  title: 'Woodpecker CI',
  tagline: 'Woodpecker is a simple CI engine with great extensibility.',
  url: 'https://woodpecker-ci.github.io',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',
  organizationName: 'woodpecker-ci',
  projectName: 'woodpecker-ci.github.io',
  trailingSlash: false,
  themeConfig: {
    navbar: {
      title: 'Woodpecker',
      logo: {
        alt: 'Woodpecker Logo',
        src: 'img/logo.svg',
        srcDark: 'img/logo-darkmode.svg',
      },
      items: [
        {
          type: 'doc',
          docId: 'intro',
          position: 'left',
          label: 'Docs',
        },
        {
          type: 'doc',
          docId: 'faq',
          position: 'left',
          label: 'FAQ',
        },
        {
          type: 'doc',
          docId: 'migrations',
          position: 'left',
          label: 'Migrations',
        },
        // {to: '/blog', label: 'Blog', position: 'left'},
        {
          href: 'https://github.com/woodpecker-ci/woodpecker',
          label: 'GitHub',
          position: 'right',
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
              label: 'Usage',
              to: '/docs/usage/intro',
            },
            {
              label: 'Server setup',
              to: '/docs/administration/setup',
            },
            {
              label: 'FAQ',
              to: '/docs/faq',
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
          ],
        },
        {
          title: 'More',
          items: [
            // {
            //   label: 'Blog',
            //   to: '/blog',
            // },
            {
              label: 'GitHub',
              href: 'https://github.com/woodpecker-ci/woodpecker',
            },
            {
              href: 'https://wp.laszlo.cloud/woodpecker-ci/woodpecker',
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
  },
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: 'https://github.com/woodpecker-ci/woodpecker/edit/master/docs/',
        },
        // blog: {
        //   showReadingTime: true,
        //   editUrl: 'https://github.com/woodpecker-ci/woodpecker/edit/master/docs/blog/',
        // },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      },
    ],
  ],
};
