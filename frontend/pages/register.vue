<template>
  <v-container>
    <v-row justify="center">
      <v-col cols="12" sm="8" md="6">
        <v-card class="pa-6" aria-label="Форма регистрации">
          <v-card-title class="justify-center">
            <h2 aria-label="Регистрация">Регистрация</h2>
          </v-card-title>
          <v-card-text>
            <v-alert v-if="errorMessage" type="error" dismissible class="mb-4" aria-label="Сообщение об ошибке">
              {{ errorMessage }}
            </v-alert>
            <v-form v-model="valid" @submit.prevent="register">
              <v-text-field
                v-model="form.username"
                label="Имя пользователя"
                prepend-icon="mdi-account"
                :rules="[v => !!v || 'Имя пользователя обязательно']"
                required
                aria-label="Имя пользователя"
              />
              <v-text-field
                v-model="form.email"
                label="Email"
                prepend-icon="mdi-email"
                :rules="[v => !!v || 'Email обязателен', v => /.+@.+\..+/.test(v) || 'Некорректный email']"
                required
                aria-label="Email"
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
              <v-text-field
                v-model="form.full_name"
                label="Полное имя"
                prepend-icon="mdi-account-box"
                :rules="[v => !!v || 'Полное имя обязательно']"
                required
                aria-label="Полное имя"
              />
              <v-textarea
                v-model="form.description"
                label="Описание"
                prepend-icon="mdi-information"
                aria-label="Описание"
              />
              <v-radio-group v-model="form.role" label="Роль" :rules="[v => !!v || 'Роль обязательна']" aria-label="Выбор роли">
                <v-radio label="Нутрициолог" value="nutri" aria-label="Роль Нутрициолог"></v-radio>
                <v-radio label="Клиент" value="client" aria-label="Роль Клиент"></v-radio>
              </v-radio-group>
              <v-btn
                color="primary"
                type="submit"
                :disabled="!valid"
                block
                class="mt-4"
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
  </v-container>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useRuntimeConfig } from 'nuxt/app'

definePageMeta({ middleware: 'guest' })

const config = useRuntimeConfig()
const router = useRouter()
const valid = ref(false)
const form = ref({
  username: '',
  email: '',
  password: '',
  full_name: '',
  description: '',
  role: ''
})
const errorMessage = ref('')

const register = async () => {
  const body = {
    username: form.value.username,
    email: form.value.email,
    password: form.value.password,
    full_name: form.value.full_name,
    description: form.value.description,
    role: form.value.role
  }
  try {
    const data = await $fetch(`${config.public.apiBase}/api/register`, {
      method: 'POST',
      body
    })
    if (process.client) {
      localStorage.setItem('token', data.token)
      localStorage.setItem('role', data.role)
      localStorage.setItem('userId', data.id)
    }
    router.push('/profile')
  } catch (error) {
    errorMessage.value = 'Ошибка регистрации: ' + (error.response?.data?.error || error.message || 'Неизвестная ошибка')
  }
}
</script>