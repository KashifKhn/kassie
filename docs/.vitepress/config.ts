import { defineConfig } from 'vitepress'
import versionInfo from '../version.json'

const SITE_URL = 'https://kassie.kashifkhan.dev'

export default defineConfig({
  title: 'Kassie',
  description: 'Modern Database Explorer for Cassandra & ScyllaDB',
  base: '/',
  ignoreDeadLinks: false,

  sitemap: {
    hostname: SITE_URL,
  },

  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/logo.svg' }],
    ['link', { rel: 'canonical', href: SITE_URL }],
    ['meta', { name: 'theme-color', content: '#5f67ee' }],
    ['meta', { name: 'google-site-verification', content: 'GsnVvhnHNdZC3YMYInZxanh97DATIV3-Ooup2YHYcS8' }],

    // Open Graph
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:locale', content: 'en' }],
    ['meta', { property: 'og:title', content: 'Kassie | Database Explorer for Cassandra & ScyllaDB' }],
    ['meta', { property: 'og:site_name', content: 'Kassie' }],
    ['meta', { property: 'og:description', content: 'Modern terminal and web explorer for Apache Cassandra and ScyllaDB' }],
    ['meta', { property: 'og:url', content: SITE_URL }],
    ['meta', { property: 'og:image', content: `${SITE_URL}/og-image.png` }],
    ['meta', { property: 'og:image:width', content: '1200' }],
    ['meta', { property: 'og:image:height', content: '630' }],
    ['meta', { property: 'og:image:alt', content: 'Kassie - Database Explorer for Cassandra & ScyllaDB' }],

    // Twitter Card
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:title', content: 'Kassie | Database Explorer for Cassandra & ScyllaDB' }],
    ['meta', { name: 'twitter:description', content: 'Modern terminal and web explorer for Apache Cassandra and ScyllaDB' }],
    ['meta', { name: 'twitter:image', content: `${SITE_URL}/og-image.png` }],
    ['meta', { name: 'twitter:image:alt', content: 'Kassie - Database Explorer for Cassandra & ScyllaDB' }],

    // JSON-LD structured data
    ['script', { type: 'application/ld+json' }, JSON.stringify({
      '@context': 'https://schema.org',
      '@type': 'SoftwareApplication',
      name: 'Kassie',
      applicationCategory: 'DeveloperApplication',
      operatingSystem: 'Linux, macOS, Windows',
      description: 'Modern terminal and web database explorer for Apache Cassandra and ScyllaDB. Single binary with embedded TUI and web UI.',
      url: SITE_URL,
      downloadUrl: 'https://github.com/KashifKhn/kassie/releases',
      author: {
        '@type': 'Person',
        name: 'KashifKhn',
        url: 'https://github.com/KashifKhn',
      },
      license: 'https://opensource.org/licenses/MIT',
      codeRepository: 'https://github.com/KashifKhn/kassie',
      keywords: 'cassandra, scylladb, database, explorer, tui, terminal, cli, web ui',
    })],
    ['script', { type: 'application/ld+json' }, JSON.stringify({
      '@context': 'https://schema.org',
      '@type': 'WebSite',
      name: 'Kassie Documentation',
      url: SITE_URL,
      description: 'Official documentation for Kassie, a modern database explorer for Apache Cassandra and ScyllaDB.',
      potentialAction: {
        '@type': 'SearchAction',
        target: `${SITE_URL}/?search={search_term_string}`,
        'query-input': 'required name=search_term_string',
      },
    })],
  ],

  themeConfig: {
    logo: '/logo.svg',
    
    nav: [
      { text: 'Guide', link: '/guide/', activeMatch: '/guide/' },
      { text: 'Reference', link: '/reference/', activeMatch: '/reference/' },
      { text: 'Architecture', link: '/architecture/', activeMatch: '/architecture/' },
      { text: 'Development', link: '/development/', activeMatch: '/development/' },
      { text: 'Examples', link: '/examples/', activeMatch: '/examples/' },
      {
        text: versionInfo.version,
        items: [
          { text: 'Changelog', link: 'https://github.com/KashifKhn/kassie/releases' },
          { text: 'Contributing', link: '/development/contributing' },
        ]
      }
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Introduction',
          items: [
            { text: 'What is Kassie?', link: '/guide/' },
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Installation', link: '/guide/installation' },
          ]
        },
        {
          text: 'Configuration',
          items: [
            { text: 'Configuration Guide', link: '/guide/configuration' },
          ]
        },
        {
          text: 'Usage',
          items: [
            { text: 'TUI Interface', link: '/guide/tui-usage' },
            { text: 'Web Interface', link: '/guide/web-usage' },
            { text: 'Compatibility', link: '/guide/compatibility' },
            { text: 'Troubleshooting', link: '/guide/troubleshooting' },
          ]
        }
      ],

      '/reference/': [
        {
          text: 'Reference',
          items: [
            { text: 'Overview', link: '/reference/' },
            { text: 'CLI Commands', link: '/reference/cli-commands' },
            { text: 'Configuration Schema', link: '/reference/configuration-schema' },
            { text: 'Keyboard Shortcuts', link: '/reference/keyboard-shortcuts' },
            { text: 'API Reference', link: '/reference/api' },
            { text: 'Error Codes', link: '/reference/error-codes' },
          ]
        }
      ],

      '/architecture/': [
        {
          text: 'Architecture',
          items: [
            { text: 'Overview', link: '/architecture/' },
            { text: 'Client-Server Model', link: '/architecture/client-server' },
            { text: 'Authentication', link: '/architecture/authentication' },
            { text: 'State Management', link: '/architecture/state-management' },
            { text: 'Protocol Design', link: '/architecture/protocol' },
          ]
        }
      ],

      '/development/': [
        {
          text: 'Development',
          items: [
            { text: 'Overview', link: '/development/' },
            { text: 'Setup', link: '/development/setup' },
            { text: 'Building', link: '/development/building' },
            { text: 'Testing', link: '/development/testing' },
            { text: 'Architecture Decisions', link: '/development/architecture-decisions' },
            { text: 'Contributing', link: '/development/contributing' },
          ]
        }
      ],

      '/examples/': [
        {
          text: 'Examples',
          items: [
            { text: 'Overview', link: '/examples/' },
            { text: 'Basic Queries', link: '/examples/basic-queries' },
            { text: 'Advanced Filtering', link: '/examples/advanced-filtering' },
            { text: 'Custom Configurations', link: '/examples/custom-configs' },
            { text: 'Scripting & Automation', link: '/examples/scripting' },
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/KashifKhn/kassie' }
    ],

    editLink: {
      pattern: 'https://github.com/KashifKhn/kassie/edit/main/docs/:path',
      text: 'Edit this page on GitHub'
    },

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024-present KashifKhn'
    },

    search: {
      provider: 'local'
    },

    lastUpdated: {
      text: 'Updated at',
      formatOptions: {
        dateStyle: 'full',
        timeStyle: 'medium'
      }
    }
  },

  markdown: {
    theme: {
      light: 'github-light',
      dark: 'github-dark'
    }
  },

  vite: {
    define: {
      __VERSION__: JSON.stringify(versionInfo.version),
      __COMMIT__: JSON.stringify(versionInfo.commit),
      __BUILD_DATE__: JSON.stringify(versionInfo.buildDate)
    }
  }
})
