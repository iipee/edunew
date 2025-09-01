<template>
  <v-app :theme="themeStore.theme">
    <AppHeader />
    <v-main style="padding-top: 64px; min-height: calc(100vh - 64px); display: flex; flex-direction: column;">
      <NuxtPage />
    </v-main>
    <v-footer color="primary" :dark="themeStore.theme === 'dark'" class="text-center" style="border-top: 1px solid #CED4DA; padding: 16px; position: relative;">
      <v-col cols="12">
        practice © 2025 | <NuxtLink to="/">Главная</NuxtLink> | <NuxtLink to="/search">Поиск</NuxtLink> | <NuxtLink to="/profile">Профиль</NuxtLink>
      </v-col>
    </v-footer>
  </v-app>
</template>

<script setup>
import { onMounted } from 'vue'
import { useNuxtApp } from 'nuxt/app'
import { useThemeStore } from '~/stores/theme'

const { $vuetify } = useNuxtApp()
const themeStore = useThemeStore()

onMounted(() => {
  if (process.client) {
    const savedTheme = localStorage.getItem('theme') || 'light'
    themeStore.setTheme(savedTheme)
    $vuetify.theme.global.name = savedTheme
    document.documentElement.setAttribute('data-theme', savedTheme)
  }
})
</script>

<style scoped>
.v-footer .text-center a {
  color: #FFFFFF !important;
  text-decoration: none !important;
}
.v-footer .text-center a:hover {
  text-decoration: underline !important;
  color: #E0E0E0 !important;
}
</style>