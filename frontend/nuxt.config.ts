import { defineNuxtConfig } from 'nuxt/config'

export default defineNuxtConfig({
  ssr: true,
  modules: [
    ['vuetify-nuxt-module', {
      treeshaking: true,
      defaultAssets: true,
      theme: {
        defaultTheme: 'light',
        themes: {
          light: {
            dark: false,
            colors: {
              primary: '#28A745',
              light: '#FFFFFF',
              dark: '#218838',
              text: '#212529',
              bg: '#F8F9FA',
              error: '#DC3545',
              success: '#28A745',
              warning: '#FFB74D'
            }
          },
          dark: {
            dark: true,
            colors: {
              primary: '#28A745',
              light: '#212529',
              dark: '#1A1A1A',
              text: '#FFFFFF',
              bg: '#212529',
              error: '#DC3545',
              success: '#28A745',
              warning: '#FFB74D'
            }
          }
        },
        options: {
          customProperties: true,
          variations: true,
          ssr: true
        }
      }
    }],
    '@pinia/nuxt' // Добавляем модуль Pinia
  ],
  css: ['~/assets/style.css'],
  compatibilityDate: '2025-07-24',
  runtimeConfig: {
    public: {
      apiBase: process.env.API_BASE_URL || 'http://localhost:8080',
      wsBase: process.env.WS_BASE_URL || 'ws://localhost:8080'
    }
  },
  plugins: ['~/plugins/mitt.ts', '~/plugins/auth.ts', '~/plugins/websocket.js'],
  components: true,
  vite: {
    server: {
      proxy: {
        '/api': {
          target: process.env.API_BASE_URL || 'http://localhost:8080',
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, '')
        },
        '/ws': {
          target: process.env.WS_BASE_URL || 'ws://localhost:8080',
          ws: true
        }
      }
    },
    ssr: {
      noExternal: [/vuetify/]
    }
  }
})