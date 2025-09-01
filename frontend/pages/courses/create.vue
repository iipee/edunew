<template>
  <v-container>
    <v-row justify="center">
      <v-col cols="12" sm="8" md="6">
        <v-card class="pa-6" aria-label="Форма создания услуги">
          <v-card-title class="justify-center">
            <h2 aria-label="Создать услугу">Создать услугу</h2>
          </v-card-title>
          <v-card-text>
            <v-alert v-if="errorMessage" type="error" dismissible class="mb-4" aria-label="Ошибка">
              {{ errorMessage }}
            </v-alert>
            <v-form v-model="valid" @submit.prevent="createCourse">
              <v-text-field
                v-model="form.title"
                label="Название услуги"
                prepend-icon="mdi-food-apple"
                :rules="[v => !!v || 'Название обязательно']"
                required
                aria-label="Название услуги"
              />
              <v-textarea
                v-model="form.services"
                label="Услуги (через запятую)"
                prepend-icon="mdi-format-list-bulleted"
                :rules="[v => !!v || 'Услуги обязательны']"
                required
                aria-label="Услуги"
              />
              <v-textarea
                v-model="form.description"
                label="Описание услуги"
                prepend-icon="mdi-information"
                :rules="[v => !!v || 'Описание обязательно']"
                required
                aria-label="Описание услуги"
              />
              <v-text-field
                v-model="form.net_price"
                label="Чистая цена (руб.)"
                prepend-icon="mdi-currency-rub"
                type="number"
                :rules="[v => !!v || 'Стоимость обязательна', v => v > 0 || 'Стоимость должна быть больше 0']"
                required
                aria-label="Чистая цена"
              />
              <v-text-field
                v-model="form.video_url"
                label="URL видео"
                prepend-icon="mdi-video"
                aria-label="URL видео"
              />
              <v-btn
                color="primary"
                type="submit"
                :disabled="!valid"
                block
                class="mt-4"
                v-tooltip="'Создать новую услугу'"
                aria-label="Создать услугу"
              >
                Создать услугу
              </v-btn>
            </v-form>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useRuntimeConfig } from 'nuxt/app'

definePageMeta({ middleware: 'auth' })

const config = useRuntimeConfig()
const router = useRouter()
const valid = ref(false)
const form = ref({
  title: '',
  services: '',
  description: '',
  net_price: null,
  video_url: ''
})
const token = ref(null)
const role = ref('')
const errorMessage = ref('')

onMounted(() => {
  if (process.client) {
    token.value = localStorage.getItem('token')
    role.value = localStorage.getItem('role')
    if (!token.value) {
      errorMessage.value = 'Требуется авторизация'
      router.push('/login')
    } else if (role.value !== 'nutri') {
      errorMessage.value = 'Доступ только для нутрициологов'
      router.push('/profile')
    }
  }
})

const createCourse = async () => {
  const servicesArray = form.value.services.split(',').map(s => s.trim()).filter(s => s)
  if (servicesArray.length === 0) {
    errorMessage.value = 'Укажите хотя бы одну услугу'
    return
  }
  const headers = { Authorization: `Bearer ${token.value}` }
  const body = {
    title: form.value.title,
    services: servicesArray,
    description: form.value.description,
    net_price: parseFloat(form.value.net_price),
    video_url: form.value.video_url
  }
  const { data, error } = await useFetch(`${config.public.apiBase}/api/courses`, { 
    method: 'POST', 
    headers, 
    body 
  })
  if (error.value) {
    errorMessage.value = 'Ошибка создания услуги: ' + (error.value.data?.error || 'Неизвестная ошибка')
    return
  }
  router.push('/profile')
}
</script>