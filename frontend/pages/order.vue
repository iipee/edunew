<template>
  <v-container>
    <v-row justify="center">
      <v-col cols="12" sm="8" md="6">
        <v-card class="pa-6" aria-label="Заказ услуги">
          <v-card-title class="justify-center">
            <h2 aria-label="Заказ услуги">Заказ услуги</h2>
          </v-card-title>
          <v-card-text>
            <v-alert v-if="errorMessage" type="error" dismissible class="mb-4" aria-label="Сообщение об ошибке">
              {{ errorMessage }}
            </v-alert>
            <v-list v-if="courses.length > 0" aria-label="Список доступных услуг">
              <v-list-item v-for="(course, index) in courses" :key="index" aria-label="Услуга">
                <v-list-item-content>
                  <v-list-item-title aria-label="Название услуги">{{ course.title }} - <span style="font-size:20px;color:#28A745;font-weight:bold;">{{ course.gross_price }} руб.</span></v-list-item-title>
                  <v-list-item-subtitle aria-label="Описание услуги">{{ course.description }}</v-list-item-subtitle>
                </v-list-item-content>
                <v-list-item-action>
                  <v-btn 
                    color="#28A745" 
                    :disabled="course.is_paid" 
                    @click="orderCourse(course)" 
                    v-tooltip="course.is_paid ? 'Курс уже оплачен' : 'Записаться'" 
                    aria-label="Записаться на услугу"
                  >
                    {{ course.is_paid ? 'Оплачено' : 'Записаться' }}
                  </v-btn>
                </v-list-item-action>
              </v-list-item>
            </v-list>
            <p v-else aria-label="Нет доступных услуг">Нет доступных услуг</p>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    <v-snackbar v-model="snackbar" :color="snackbarColor" timeout="3000" aria-label="Уведомление">
      {{ snackbarText }}
    </v-snackbar>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRuntimeConfig } from 'nuxt/app'
import { useRouter } from 'vue-router'
import { useAuthStore } from '~/stores/auth'

const config = useRuntimeConfig()
const router = useRouter()
const authStore = useAuthStore()
const courses = ref([])
const token = ref(null)
const userId = ref(null)
const errorMessage = ref('')
const snackbar = ref(false)
const snackbarText = ref('')
const snackbarColor = ref('success')

onMounted(async () => {
  if (process.client) {
    token.value = localStorage.getItem('token')
    userId.value = localStorage.getItem('userId') ? parseInt(localStorage.getItem('userId')) : null
  }
  if (!token.value) {
    errorMessage.value = 'Требуется авторизация'
    router.push('/login')
    return
  }
  const headers = { Authorization: `Bearer ${token.value}` }
  try {
    const data = await $fetch(`${config.public.apiBase}/api/courses`, { headers })
    const paymentStatus = await $fetch(`${config.public.apiBase}/api/payments/status`, { headers })
    courses.value = data.map(course => ({
      ...course,
      is_paid: paymentStatus.some(status => status.course_id === course.id && status.user_id === userId.value)
    })) || []
  } catch (error) {
    errorMessage.value = 'Ошибка загрузки услуг: ' + (error.response?.data?.error || error.message || 'Неизвестная ошибка')
  }
})

const orderCourse = async (course) => {
  if (!token.value) {
    router.push('/login')
    return
  }
  if (course.is_paid) {
    snackbarText.value = 'Курс уже оплачен'
    snackbarColor.value = 'info'
    snackbar.value = true
    return
  }
  const headers = { Authorization: `Bearer ${token.value}` }
  const body = { course_id: course.id }
  try {
    const data = await $fetch(`${config.public.apiBase}/api/payments/create`, {
      method: 'POST',
      headers,
      body
    })
    window.location.href = data.confirmation_url
  } catch (error) {
    snackbarText.value = 'Ошибка заказа: ' + (error.response?.data?.error || error.message || 'Неизвестная ошибка')
    snackbarColor.value = 'error'
    snackbar.value = true
  }
}
</script>