<template>
  <v-container>
    <v-row justify="center">
      <v-col cols="12" sm="8" md="6">
        <v-card class="pa-6" aria-label="Форма входа">
          <v-card-title class="justify-center">
            <h2 aria-label="Вход">Вход</h2>
          </v-card-title>
          <v-card-text>
            <v-alert v-if="errorMessage" type="error" dismissible class="mb-4" aria-label="Сообщение об ошибке">
              {{ errorMessage }}
            </v-alert>
            <v-form v-model="valid" @submit.prevent="login">
              <v-text-field
                v-model="form.username"
                label="Имя пользователя"
                prepend-icon="mdi-account"
                :rules="[v => !!v || 'Имя обязательно']"
                required
                aria-label="Имя пользователя"
              />
              <v-text-field
                v-model="form.password"
                label="Пароль"
                prepend-icon="mdi-lock"
                type="password"
                :rules="[v => !!v || 'Пароль обязателен', v => v.length >= 6 || 'Минимум 6 символов']"
                required
                aria-label="Пароль"
              />
              <v-btn
                color="primary"
                type="submit"
                :disabled="!valid"
                :loading="loading"
                block
                class="mt-4"
                v-tooltip="'Войти в аккаунт'"
                aria-label="Войти"
              >
                Войти
              </v-btn>
            </v-form>
            <p class="mt-4 text-center" aria-label="Ссылка на регистрацию">Нет профиля? <NuxtLink to="/register">Зарегистрироваться</NuxtLink></p>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useRuntimeConfig } from 'nuxt/app'
import { useNuxtApp } from 'nuxt/app'

const { $emitter } = useNuxtApp()
const emitter = $emitter
const config = useRuntimeConfig()
const router = useRouter()
const valid = ref(false)
const loading = ref(false)
const errorMessage = ref('')
const form = ref({
  username: '',
  password: ''
})

const login = async () => {
  loading.value = true
  const { data, error } = await useFetch(`${config.public.apiBase}/api/login`, {
    method: 'POST',
    body: form.value
  })
  loading.value = false
  if (error.value) {
    errorMessage.value = error.value.data?.error || 'Неверные данные'
    return
  }
  if (process.client) {
    localStorage.setItem('token', data.value.token)
    localStorage.setItem('role', data.value.role)
    localStorage.setItem('userId', data.value.id)
    emitter.emit('login')
  }
  router.push('/profile')  // Мгновенный redirect в профиль для показа данных
}
</script>