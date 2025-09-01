import { defineNuxtConfig } from 'nuxt/config'

export default defineNuxtConfig({
  // Включаем серверный рендеринг (SSR) для SEO и производительности
  ssr: true,

  // Подключаем модули Nuxt
  modules: [
    ['vuetify-nuxt-module', {
      treeshaking: true, // Включаем tree-shaking для оптимизации Vuetify
      defaultAssets: true, // Используем встроенные шрифты и иконки Vuetify
      theme: {
        defaultTheme: 'light', // Устанавливаем светлую тему по умолчанию
        themes: {
          light: {
            dark: false,
            colors: {
              primary: '#28A745', // Основной зелёный цвет (кнопки, акценты)
              light: '#FFFFFF', // Белый фон
              dark: '#218838', // Тёмно-зелёный для акцентов
              text: '#212529', // Основной цвет текста
              bg: '#F8F9FA', // Светлый фон страниц
              error: '#DC3545', // Красный для ошибок
              success: '#28A745', // Зелёный для успешных операций
              warning: '#FFB74D' // Оранжевый для предупреждений
            }
          },
          dark: {
            dark: true,
            colors: {
              primary: '#28A745', // Зелёный для кнопок в тёмной теме
              light: '#212529', // Тёмный фон для элементов
              dark: '#1A1A1A', // Основной тёмный фон
              text: '#FFFFFF', // Белый текст
              bg: '#212529', // Тёмный фон страниц
              error: '#DC3545', // Красный для ошибок
              success: '#28A745', // Зелёный для успешных операций
              warning: '#FFB74D' // Оранжевый для предупреждений
            }
          }
        },
        options: {
          customProperties: true, // Включаем CSS-переменные Vuetify
          variations: true, // Включаем вариации цветов
          ssr: true // Поддержка SSR для Vuetify
        }
      }
    }],
    '@pinia/nuxt' // Модуль Pinia для управления состоянием
  ],

  // Подключаем глобальный CSS
  css: ['~/assets/style.css'],

  // Указываем дату совместимости для Nuxt
  compatibilityDate: '2025-07-24',

  // Настройки переменных окружения для API и WebSocket
  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080', // Бэкенд API
      wsBase: process.env.NUXT_PUBLIC_WS_BASE || 'ws://localhost:8080' // WebSocket для чата
    }
  },

  // Подключаем плагины (порядок важен, чтобы избежать конфликтов)
  plugins: [
    '~/plugins/mitt.ts', // Глобальный эмиттер событий
    '~/plugins/auth.ts', // Аутентификация (useAuthStore)
    '~/plugins/websocket.js' // WebSocket для чата (заменено с websocket.client.js)
  ],

  // Автоматическое сканирование компонентов
  components: true,

  // Настройки Vite для проксирования, SSR и HMR
  vite: {
    server: {
      hmr: {
        protocol: 'ws', // Используем WebSocket вместо именованных каналов для HMR
        host: 'localhost',
        port: 3000
      },
      proxy: {
        '/api': {
          target: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080',
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, '') // Удаляем /api из пути
        },
        '/ws': {
          target: process.env.NUXT_PUBLIC_WS_BASE || 'ws://localhost:8080',
          ws: true // Включаем проксирование WebSocket
        }
      }
    },
    ssr: {
      noExternal: [/vuetify/] // Исключаем Vuetify из внешних модулей для SSR
    }
  }
})