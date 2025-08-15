<template>
  <v-app-bar app color="primary" :dark="$vuetify.theme.global.name === 'dark'" height="64">
    <v-toolbar-title class="text-h6" aria-label="Логотип NutriPlatform">
      <span style="color: #28A745;">NutriPlatform</span>
    </v-toolbar-title>
    <v-spacer />
    <v-btn icon to="/search" aria-label="Поиск курсов">
      <v-icon :color="themeStore.theme === 'dark' ? '#A5D6A7' : '#28A745'">mdi-magnify</v-icon>
    </v-btn>
    <v-btn text to="/" aria-label="Перейти на главную страницу">Главная</v-btn>
    <v-btn 
      v-if="isLoggedIn && role === 'nutri'" 
      text 
      to="/courses/create" 
      aria-label="Создать новый курс"
    >
      Создать курс
    </v-btn>
    <v-btn 
      v-if="!isLoggedIn" 
      text 
      to="/login" 
      aria-label="Войти в аккаунт"
    >
      Войти
    </v-btn>
    <v-btn 
      v-if="isLoggedIn" 
      text 
      to="/profile" 
      aria-label="Перейти в профиль"
    >
      Профиль
    </v-btn>
    <v-btn 
      v-if="isLoggedIn" 
      text 
      @click="logout" 
      aria-label="Выйти из аккаунта"
    >
      Выйти
    </v-btn>
    <v-btn 
      v-if="isLoggedIn" 
      icon 
      to="/chats" 
      aria-label="Чат"
    >
      <v-badge :content="chatStore.unreadCount" color="error" overlap v-if="chatStore.unreadCount > 0">
        <v-icon :color="themeStore.theme === 'dark' ? '#A5D6A7' : '#28A745'">mdi-message-text</v-icon>
      </v-badge>
      <v-icon :color="themeStore.theme === 'dark' ? '#A5D6A7' : '#28A745'" v-else>mdi-message-text</v-icon>
    </v-btn>
    <v-btn 
      icon 
      @click="themeStore.toggleTheme" 
      style="margin-left: 16px;" 
      aria-label="Переключить тему"
    >
      <v-icon size="24" :color="themeStore.theme === 'dark' ? '#FFB300' : '#343A40'">
        {{ themeStore.theme === 'dark' ? 'mdi-white-balance-sunny' : 'mdi-moon-waxing-crescent' }}
      </v-icon>
    </v-btn>
  </v-app-bar>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useNuxtApp } from 'nuxt/app'
import { useChatStore } from '~/stores/chat'
import { useThemeStore } from '~/stores/theme'
import { useAuthStore } from '~/stores/auth'

const { $websocket } = useNuxtApp()
const router = useRouter()
const chatStore = useChatStore()
const themeStore = useThemeStore()
const authStore = useAuthStore()

// Reactive state from authStore
const isLoggedIn = computed(() => authStore.isLoggedIn)
const role = computed(() => authStore.role || '')

onMounted(() => {
  authStore.initialize() // Ensure store is initialized with localStorage data
  if (isLoggedIn.value) {
    chatStore.fetchDialogs()
  }
})

const logout = () => {
  authStore.clearUser() // Use store to clear auth data
  $websocket.disconnect()
  router.push('/login')
}
</script>