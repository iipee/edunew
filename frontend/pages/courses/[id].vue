<template>
  <v-container>
    <v-row justify="center">
      <v-col cols="12" sm="8" md="6">
        <v-skeleton-loader v-if="courseStore.loading" type="card" aria-label="Загрузка карточки" />
        <v-alert v-if="courseStore.error" type="error" dismissible class="mb-4" aria-label="Ошибка загрузки">{{ courseStore.error }}</v-alert>
        <v-card v-else class="pa-6" elevation="2" style="border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); background-color: #FFFFFF;" aria-label="Детали курса">
          <v-card-title class="justify-center" style="font-size: 32px; font-weight: bold; color: #212529;" aria-label="Название курса">{{ courseStore.course.title || 'Курс' }}</v-card-title>
          <v-card-text>
            <v-avatar size="100" class="mb-4">
              <v-img 
                :src="courseStore.course.teacher?.avatar_url || '/images/nutri-placeholder.jpg'" 
                aria-label="Фото автора"
              >
                <template v-slot:placeholder>
                  <v-icon size="100" color="#CED4DA">mdi-account-circle</v-icon>
                </template>
              </v-img>
            </v-avatar>
            <p style="font-size: 16px; line-height: 1.5; color: #212529;" aria-label="Автор курса">Автор: {{ courseStore.course.teacher ? courseStore.course.teacher.full_name : 'Не указан' }}</p>
            <v-list dense style="margin: 16px 0;" aria-label="Список услуг">
              <v-list-item v-for="(service, i) in courseStore.course.services" :key="i">
                <v-list-item-icon>
                  <v-icon color="#28A745">mdi-check</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title style="font-size: 16px; color: #212529;" aria-label="Услуга">{{ service }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </v-list>
            <p style="font-size: 16px; line-height: 1.5; color: #212529; margin: 16px 0;" aria-label="Описание курса">Описание: {{ courseStore.course.description || 'Описание не указано' }}</p>
            <p style="font-size: 24px; font-weight: bold; color: #28A745; margin: 16px 0;" aria-label="Стоимость">Цена: {{ displayedPrice }} руб.</p> <!-- ИЗМЕНЕНО: Зеленый, жирный gross_price по ТЗ -->
            <v-btn 
              v-if="!isPaid && isClient" 
              color="#28A745" 
              @click="submitPayment" 
              v-tooltip="'Записаться'" 
              aria-label="Записаться" 
            > <!-- ИЗМЕНЕНО: color="#28A745" по ТЗ -->
              Записаться
            </v-btn>
            <v-btn v-if="isPaid" color="success" disabled aria-label="Уже оплачено">Оплачено</v-btn>
            <v-btn v-if="isClient" color="primary" @click="openChat" v-tooltip="'Открыть чат'" aria-label="Открыть чат" class="ml-2">
              Чат с автором
            </v-btn>
            <v-btn v-if="canReview && !hasReviewed" color="secondary" @click="openReviewModal" v-tooltip="'Оставить отзыв'" aria-label="Оставить отзыв" class="ml-2">
              Оставить отзыв
            </v-btn>
          </v-card-text>
        </v-card>
        <v-dialog v-model="dialogReview" max-width="500" aria-label="Диалог отзыва">
          <v-card>
            <v-card-title aria-label="Оставить отзыв">Оставить отзыв</v-card-title>
            <v-card-text>
              <v-textarea v-model="reviewText" label="Ваш отзыв" aria-label="Поле отзыва" />
            </v-card-text>
            <v-card-actions>
              <v-btn color="primary" @click="submitReview" aria-label="Отправить отзыв">Отправить</v-btn>
              <v-btn text @click="dialogReview = false" aria-label="Отмена">Отмена</v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
        <v-snackbar v-model="snackbar" :color="snackbarColor" timeout="3000" aria-label="Уведомление">
          {{ snackbarText }}
        </v-snackbar>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useRuntimeConfig } from 'nuxt/app'
import { useCourseStore } from '~/stores/course'
import debounce from 'lodash/debounce'

const config = useRuntimeConfig()
const route = useRoute()
const router = useRouter()
const courseStore = useCourseStore()
const token = ref(null)
const role = ref('')
const userId = ref(null)
const isLoggedIn = computed(() => !!token.value)
const isClient = computed(() => role.value === 'client')
const isPaid = computed(() => courseStore.isPaid)
const canReview = computed(() => courseStore.canReview)
const hasReviewed = ref(false)
const dialogReview = ref(false)
const reviewText = ref('')
const snackbar = ref(false)
const snackbarText = ref('')
const snackbarColor = ref('success')
const displayedPrice = computed(() => courseStore.course.gross_price || courseStore.course.net_price || 0)

onMounted(async () => {
  if (process.client) {
    token.value = localStorage.getItem('token')
    role.value = localStorage.getItem('role')
    userId.value = localStorage.getItem('userId') ? parseInt(localStorage.getItem('userId')) : null
  }
  await courseStore.loadCourse(route.params.id)
  if (isLoggedIn.value && isClient.value) {
    await courseStore.checkCanReview(userId.value)
  }
})

const submitPayment = async () => {
  if (!isLoggedIn.value) {
    router.push('/login') // ИЗМЕНЕНО: Redirect если не логирован по ТЗ
    return
  }
  try {
    const data = await courseStore.submitPayment(route.params.id, userId.value)
    window.location.href = data.confirmation_url
  } catch (error) {
    snackbarText.value = 'Ошибка оплаты: ' + (error.message || 'Неизвестная ошибка')
    snackbarColor.value = 'error'
    snackbar.value = true
  }
}

const openChat = async () => {
  if (!isLoggedIn.value) {
    router.push('/login')
    return
  }
  if (!courseStore.course?.teacher_id) {
    snackbarText.value = 'Ошибка: ID автора курса не найден'
    snackbarColor.value = 'error'
    snackbar.value = true
    return
  }
  try {
    const headers = { Authorization: `Bearer ${token.value}` }
    const data = await $fetch(`${config.public.apiBase}/api/start-chat`, {
      method: 'POST',
      headers,
      body: { receiver_id: courseStore.course.teacher_id }
    })
    if (!data?.receiver_id) {
      throw new Error('Ответ сервера не содержит receiver_id')
    }
    router.push(`/chats?selected=${data.receiver_id}`)
  } catch (error) {
    snackbarText.value = 'Ошибка открытия чата: ' + (error.response?.data?.error || error.message || 'Неизвестная ошибка')
    snackbarColor.value = 'error'
    snackbar.value = true
  }
}

const openReviewModal = () => {
  if (!isLoggedIn.value) {
    router.push('/login')
    return
  }
  dialogReview.value = true
}

const submitReview = async () => {
  if (!reviewText.value) {
    snackbarText.value = 'Отзыв обязателен'
    snackbarColor.value = 'error'
    snackbar.value = true
    return
  }
  try {
    await courseStore.submitReview(route.params.id, reviewText.value, userId.value)
    snackbarText.value = 'Отзыв отправлен'
    snackbarColor.value = 'success'
    snackbar.value = true
    dialogReview.value = false
    reviewText.value = ''
  } catch (error) {
    snackbarText.value = 'Ошибка отправки отзыва: ' + (error.message || 'Неизвестная ошибка')
    snackbarColor.value = 'error'
    snackbar.value = true
  }
}
</script>

<style scoped>
.v-card:hover {
  box-shadow: 0 4px 8px rgba(0,0,0,0.2);
}
.v-card-title {
  padding: 16px 0;
}
.v-card-text {
  padding: 16px;
}
</style>