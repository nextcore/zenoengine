import { defineConfig } from 'vitepress'

export default defineConfig({
  title: "ZenoEngine",
  description: "The Lightning Fast, Fullstack Web Engine.",
  cleanUrls: true,

  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/logo.svg' }],
    // Add Inter font to match Laravel
    ['link', { rel: 'preconnect', href: 'https://fonts.googleapis.com' }],
    ['link', { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' }],
    ['link', { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap' }]
  ],

  themeConfig: {
    logo: '/logo.svg',

    nav: [
      { text: 'Home', link: '/' },
      { text: 'Docs', link: '/getting-started/installation' },
      {
        text: 'Ecosystem', items: [
          { text: 'ZenoLang', link: '/ecosystem/zenolang' },
          { text: 'Multi-App Architecture', link: '/ecosystem/multi-app' },
          { text: 'Blade Engine', link: '/views/blade' },
          { text: 'ORM', link: '/orm/eloquent' }
        ]
      }
    ],

    sidebar: {
      '/': [
        {
          text: 'Prologue',
          collapsed: false,
          items: [
            { text: 'What is ZenoEngine?', link: '/prologue/what-is-zeno' },
            { text: 'Migrating from Laravel', link: '/prologue/migrating' },
            { text: 'Release Notes', link: '/prologue/releases' }
          ]
        },
        {
          text: 'Getting Started',
          collapsed: false,
          items: [
            { text: 'Installation', link: '/getting-started/installation' },
            { text: 'Configuration', link: '/getting-started/configuration' },
            { text: 'Directory Structure', link: '/getting-started/structure' }
          ]
        },
        {
          text: 'The Basics',
          collapsed: false,
          items: [
            { text: 'Routing', link: '/basics/routing' },
            { text: 'Middleware', link: '/basics/middleware' },
            { text: 'CSRF Protection', link: '/basics/csrf' },
            { text: 'Controllers', link: '/basics/controllers' },
            { text: 'Requests', link: '/basics/requests' },
            { text: 'Responses', link: '/basics/responses' },
            { text: 'API Authentication', link: '/basics/authentication' },
            { text: 'Mail & Notifications', link: '/basics/mail' },
            { text: 'Views (Blade)', link: '/views/blade' },
            { text: 'Inertia.js SPA', link: '/views/inertia' },
            { text: 'Validation', link: '/basics/validation' }
          ]
        },
        {
          text: 'Database',
          collapsed: false,
          items: [
            { text: 'Getting Started', link: '/database/getting-started' },
            { text: 'Query Builder', link: '/database/query-builder' },
            { text: 'Pagination', link: '/database/pagination' },
            { text: 'Migrations', link: '/database/migrations' },
            { text: 'Seeding', link: '/database/seeding' }
          ]
        },
        {
          text: 'Eloquent ORM',
          collapsed: false,
          items: [
            { text: 'Getting Started', link: '/orm/eloquent' },
            { text: 'Relationships', link: '/orm/relationships' },
            { text: 'Eager Loading', link: '/orm/eager-loading' },
            { text: 'Mutators & Casts', link: '/orm/mutators' }
          ]
        },
        {
          text: 'Advanced',
          collapsed: false,
          items: [
            { text: 'Realtime SSE', link: '/advanced/realtime-sse' },
            { text: 'Background Jobs & Queues', link: '/advanced/jobs-queues' },
            { text: 'Web Server Gateway', link: '/advanced/gateway' },
            { text: 'Edge Security (WAF/Bot)', link: '/advanced/edge-security' },
            { text: 'Filesystem & Uploads', link: '/advanced/filesystem' }
          ]
        },
        {
          text: 'Ecosystem',
          collapsed: false,
          items: [
            { text: 'ZenoLang', link: '/ecosystem/zenolang' },
            { text: 'Reporting Engine (PDF)', link: '/ecosystem/reporting-engine' },
            { text: 'Stdlib API Reference', link: '/ecosystem/stdlib-reference' },
            { text: 'Multi-App Architecture', link: '/ecosystem/multi-app' },
            { text: 'WASM Plugins', link: '/ecosystem/wasm-plugins' },
            { text: 'Container Bridge (Docker)', link: '/ecosystem/container-bridge' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/zenoengine' }
    ],

    search: {
      provider: 'local'
    },

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2025-present ZenoEngine Contributors'
    }
  }
})
