import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "UPSM",
  description: "A universal modification format for the Playdate.",
  themeConfig: {
    logo: 'https://github.com/CrossScarDev/upsm/blob/main/icon.png?raw=true',

    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Spec', link: '/spec' },
      { text: 'Install', link: '/installer' }
    ],

    sidebar: [
      {
        text: 'UPSM',
        items: [
          { text: 'Specification', link: '/spec' },
          { text: 'Distributing', link: '/distribute' },
          { text: 'Install Guide', link: '/installer' }
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/CrossScarDev/upsm' }
    ],

    search: {
      provider: "local"
    }
  }
})
