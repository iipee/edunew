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
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useNuxtApp } from 'nuxt/app'
import { useChatStore } from '~/stores/chat'
import { useThemeStore } from '~/stores/theme'

const { $emitter, $websocket } = useNuxtApp()
const router = useRouter()
const chatStore = useChatStore()
const themeStore = useThemeStore()
const token = ref(null)
const role = ref('')

const isLoggedIn = computed(() => !!token.value)

onMounted(() => {
  if (process.client) {
    token.value = localStorage.getItem('token')
    role.value = localStorage.getItem('role') || ''
    if (isLoggedIn.value) {
      chatStore.fetchDialogs()
    }
    $emitter.on('login', () => {
      token.value = localStorage.getItem('token')
      role.value = localStorage.getItem('role') || ''
      chatStore.fetchDialogs()
    })
    $emitter.on('logout', () => {
      token.value = null
      role.value = ''
    })
  }
})

const logout = () => {
  if (process.client) {
    localStorage.removeItem('token')
    localStorage.removeItem('role')
    localStorage.removeItem('userId')
    $websocket.disconnect()
    $emitter.emit('logout')
  }
  token.value = null
  role.value = ''
  router.push('/login')
}
</script>