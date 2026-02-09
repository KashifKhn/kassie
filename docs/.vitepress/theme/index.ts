import { h } from 'vue'
import type { Theme } from 'vitepress'
import DefaultTheme from 'vitepress/theme'
import VersionInfo from '../components/VersionInfo.vue'
import './style.css'

export default {
  extends: DefaultTheme,
  Layout: () => {
    return h(DefaultTheme.Layout, null, {})
  },
  enhanceApp({ app, router, siteData }) {
    app.component('VersionInfo', VersionInfo)
  }
} satisfies Theme
