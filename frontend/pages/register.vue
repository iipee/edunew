<template>
  <v-container>
    <v-row justify="center">
      <v-col cols="12" sm="8" md="6">
        <v-card class="pa-6" aria-label="Регистрация">
          <v-card-title class="justify-center" aria-label="Регистрация">Регистрация</v-card-title>
          <v-card-text>
            <v-alert v-if="errorMessage" type="error" dismissible class="mb-4" aria-label="Сообщение об ошибке">
              {{ errorMessage }}
            </v-alert>
            <v-form @submit.prevent="submitRegistration">
              <v-text-field 
                v-model="fullName" 
                label="Полное имя (для отображения другим пользователям)" 
                :rules="[v => !!v || 'Полное имя обязательно', v => v.length <= 255 || 'Максимум 255 символов']" 
                aria-label="Полное имя" 
              />
              <v-text-field 
                v-model="username" 
                label="Логин (для входа, уникальный)" 
                :rules="[v => !!v || 'Логин обязательно', v => v.length <= 255 || 'Максимум 255 символов']" 
                aria-label="Логин" 
              />
              <v-text-field 
                v-model="email" 
                label="Email" 
                :rules="[v => !!v || 'Email обязательно', v => /.+@.+\..+/.test(v) || 'Неверный email']" 
                aria-label="Email" 
              />
              <v-text-field 
                v-model="password" 
                label="Пароль" 
                type="password" 
                :rules="[v => !!v || 'Пароль обязательно', v => v.length >= 8 || 'Минимум 8 символов']" 
                aria-label="Пароль" 
              />
              <v-select 
                v-model="role" 
                :items="roles" 
                item-title="title"
                item-value="value"
                label="Роль" 
                :rules="[v => !!v || 'Роль обязательна']" 
                aria-label="Роль" 
              />
              <v-textarea 
                v-model="description" 
                label="Описание услуг" 
                v-if="role === 'nutri'" 
                :rules="[v => !!v || 'Описание обязательно для нутрициолога']" 
                aria-label="Описание услуг" 
              />
              <v-btn 
                color="primary" 
                type="submit" 
                v-tooltip="'Зарегистрироваться'" 
                aria-label="Зарегистрироваться"
              >
                Зарегистрироваться
              </v-btn>
            </v-form>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    <v-snackbar v-model="snackbar" :color="snackbarColor" timeout="3000" aria-label="Уведомление о результате">
      {{ snackbarText }}
    </v-snackbar>
  </v-container>
</template>

<script setup>
import { ref } from 'vue'
import { useRuntimeConfig } from 'nuxt/app'
import { useRouter } from 'vue-router'
import { useAuthStore } from '~/stores/auth'

const config = useRuntimeConfig()
const router = useRouter()
const authStore = useAuthStore()
const fullName = ref('')
const username = ref('')
const email = ref('')
const password = ref('')
const role = ref('')
const description = ref('')
const roles = ref([
  { value: 'client', title: 'Клиент' },
  { value: 'nutri', title: 'Нутрициолог' }
])
const errorMessage = ref('')
const snackbar = ref(false)
const snackbarText = ref('')
const snackbarColor = ref('success')

const submitRegistration = async () => {
  const body = {
    full_name: fullName.value,
    username: username.value,
    email: email.value,
    password: password.value,
    role: role.value,
    description: role.value === 'nutri' ? description.value : ''
  }
  try {
    const { token, role: userRole, id } = await $fetch(`${config.public.apiBase}/api/register`, {
      method: 'POST',
      body
    })
    if (process.client) {
      localStorage.setItem('token', token)
      localStorage.setItem('role', userRole)
      localStorage.setItem('userId', id)
    }
    authStore.setUser(token, userRole, id)
    authStore.refresh() // Синхронизация состояния для обновления интерфейса
    snackbarText.value = 'Регистрация прошла успешно'
    snackbarColor.value = 'success'
    snackbar.value = true
    router.push('/profile')
  } catch (error) {
    snackbarText.value = error.data?.error || 'Ошибка регистрации: Неизвестная ошибка'
    snackbarColor.value = 'error'
    snackbar.value = true
  }
}
</script>

<style scoped>
.v-card {
  padding: 24px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
</style>