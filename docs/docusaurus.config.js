// @ts-check
const { themes } = require('prism-react-renderer');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'seqflow',
  tagline: 'High-performance lock-free Disruptor for Go',
  favicon: 'img/favicon.ico',
  url: 'https://seqflow.pages.dev',
  baseUrl: '/',
  organizationName: 'gocronx',
  projectName: 'seqflow',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  i18n: {
    defaultLocale: 'en',
    locales: ['en', 'zh-Hans'],
    localeConfigs: {
      en: { label: 'English' },
      'zh-Hans': { label: '中文' },
    },
  },

  markdown: { mermaid: true },
  themes: ['@docusaurus/theme-mermaid'],

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: { sidebarPath: './sidebars.js', routeBasePath: 'docs' },
        blog: false,
        theme: { customCss: './src/css/custom.css' },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      colorMode: { defaultMode: 'light', respectPrefersColorScheme: true },
      mermaid: { theme: { light: 'neutral', dark: 'dark' } },
      navbar: {
        title: 'seqflow',
        style: 'dark',
        items: [
          { type: 'docSidebar', sidebarId: 'docs', position: 'left', label: 'Docs' },
          { type: 'localeDropdown', position: 'right' },
          { href: 'https://github.com/gocronx/seqflow', label: 'GitHub', position: 'right' },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              { label: 'Getting Started', to: '/docs/getting-started' },
              { label: 'API Reference', to: '/docs/api' },
            ],
          },
          {
            title: 'Ecosystem',
            items: [
              { label: 'seqdelay', href: 'https://github.com/gocronx/seqdelay' },
            ],
          },
          {
            title: 'Community',
            items: [
              { label: 'GitHub Issues', href: 'https://github.com/gocronx/seqflow/issues' },
            ],
          },
        ],
        copyright: `Copyright ${new Date().getFullYear()} gocronx.`,
      },
      prism: {
        theme: themes.github,
        darkTheme: themes.dracula,
        additionalLanguages: ['go', 'bash'],
      },
    }),
};

module.exports = config;
