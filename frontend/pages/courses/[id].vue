<template>
  <v-container>
    <v-row justify="center">
      <v-col cols="12" sm="8" md="6">
        <v-skeleton-loader v-if="courseStore.loading" type="card" aria-label="Загрузка карточки" />
        <v-alert v-if="courseStore.error" type="error" dismissible class="mb-4" aria-label="Ошибка загрузки">{{ courseStore.error }}</v-alert>
        <v-card v-else class="pa-6" elevation="2" style="border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); background-color: #FFFFFF;" aria-label="Детали услуги">
          <v-card-title class="justify-center" style="font-size: 32px; font-weight: bold; color: #212529;" aria-label="Название услуги">{{ courseStore.course.title || 'Услуга' }}</v-card-title>
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
            <p style="font-size: 16px; line-height: 1.5; color: #212529;" aria-label="Автор услуги">Автор: {{ courseStore.course.teacher ? courseStore.course.teacher.full_name : 'Не указан' }}</p>
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
            <p style="font-size: 16px; line-height: 1.5; color: #212529; margin: 16px 0;" aria-label="Описание услуги">Описание: {{ courseStore.course.description || 'Описание не указано' }}</p>
            <p style="font-size: 24px; font-weight: bold; color: #28A745; margin: 16px 0;" aria-label="Стоимость">Цена: {{ displayedPrice }} руб.</p>
            <h4 style="font-size: 18px; color: #2E7D32;" aria-label="Отзывы">Отзывы</h4>
            <v-row v-if="courseStore.course.reviews && courseStore.course.reviews.length > 0" class="mt-4">
              <v-col v-for="(review, i) in courseStore.course.reviews.slice(0, 3)" :key="i" cols="12">
                <v-card flat class="pa-3" style="border-radius: 8px; border: 1px solid #e0e0e0;">
                  <v-card-text style="font-size: 14px; color: #212529;">
                    "{{ review.content }}" — {{ review.author ? review.author.full_name : 'Аноним' }}
                  </v-card-text>
                </v-card>
              </v-col>
            </v-row>
            <p v-else style="font-size: 16px; color: #6C757D;" aria-label="Нет отзывов">Нет отзывов</p>
          </v-card-text>
          <v-card-actions class="justify-center pa-4">
            <v-btn 
              v-if="!isNutri && isLoggedIn" 
              color="#28A745" 
              @click="openPayment" 
              v-tooltip="'Оплатить услугу'" 
              aria-label="Записаться на услугу"
            >
              Записаться
            </v-btn>
            <!-- Добавлена кнопка "Связаться" для всех авторизованных, если не автор услуги -->
            <v-btn 
              v-if="isLoggedIn && courseStore.course.teacher_id !== userId" 
              color="#28A745" 
              variant="outlined" 
              @click="openChat" 
              class="ml-2" 
              v-tooltip="'Открыть чат с автором'" 
              aria-label="Связаться с автором услуги"
            >
              <v-icon left>mdi-message-text</v-icon>
              Связаться
            </v-btn>
            <v-btn 
              v-if="courseStore.canReview" 
              color="secondary" 
              @click="openReviewModal" 
              class="ml-2" 
              aria-label="Оставить отзыв"
            >
              Оставить отзыв
            </v-btn>
          </v-card-actions>
          <v-dialog v-model="dialogReview" max-width="500">
            <v-card aria-label="Форма отзыва">
              <v-card-title class="justify-center" aria-label="Оставить отзыв">Оставить отзыв</v-card-title>
              <v-card-text>
                <v-textarea
                  v-model="reviewText"
                  label="Ваш отзыв"
                  :rules="[v => !!v || 'Отзыв обязателен']"
                  required
                  aria-label="Поле для отзыва"
                />
              </v-card-text>
              <v-card-actions>
                <v-spacer />
                <v-btn color="primary" @click="submitReview" aria-label="Отправить отзыв">Отправить</v-btn>
                <v-btn color="secondary" @click="dialogReview = false" aria-label="Закрыть">Закрыть</v-btn>
              </v-card-actions>
            </v-card>
          </v-dialog>
          <v-snackbar v-model="snackbar" :color="snackbarColor" timeout="3000">
            {{ snackbarText }}
          </v-snackbar>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCourseStore } from '~/stores/course'
import { useRuntimeConfig } from 'nuxt/app'

const config = useRuntimeConfig()
const route = useRoute()
const router = useRouter()
const courseStore = useCourseStore()
const token = ref(null)
const userId = ref(null)
const isLoggedIn = ref(false)
const snackbar = ref(false)
const snackbarText = ref('')
const snackbarColor = ref('error')
const dialogReview = ref(false)
const reviewText = ref('')

const role = computed(() => localStorage.getItem('role') || '')
const isNutri = computed(() => role.value === 'nutri')
const displayedPrice = computed(() => isNutri.value ? courseStore.course.net_price : courseStore.course.gross_price)

onMounted(async () => {
  if (process.client) {
    token.value = localStorage.getItem('token')
    userId.value = localStorage.getItem('userId') ? parseInt(localStorage.getItem('userId')) : null
    isLoggedIn.value = !!token.value
  }
  await courseStore.loadCourse(route.params.id)
  if (isLoggedIn.value && !isNutri.value) {
    await courseStore.checkCanReview(userId.value)
  }
})

const openPayment = async () => {
  if (!isLoggedIn.value) {
    router.push('/login')
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
    snackbarText.value = 'Ошибка: ID автора услуги не найден'
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