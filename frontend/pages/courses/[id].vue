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
            <p style="font-size: 24px; font-weight: bold; color: #28A745; margin: 16px 0;" aria-label="Стоимость">Стоимость: {{ courseStore.course.price || 0 }} руб.</p>
            <p v-if="courseStore.course.video_url" style="font-size: 16px; color: #212529; margin: 16px 0;" aria-label="Видео">Видео: <a :href="courseStore.course.video_url" style="color: #28A745;" aria-label="Ссылка на видео">Ссылка</a></p>
            <div style="display: flex; justify-content: space-between; gap: 16px; margin: 16px 0;">
              <v-btn 
                color="#28A745" 
                style="min-width: 100px; border-radius: 8px;" 
                @click="openPayment" 
                v-if="isLoggedIn && !courseStore.isPaid" 
                v-tooltip="'Записаться на курс'" 
                aria-label="Записаться"
              >
                Записаться
              </v-btn>
              <v-btn 
                outlined 
                color="#28A745" 
                style="min-width: 100px; border-radius: 8px;" 
                @click="openChat" 
                v-if="isLoggedIn" 
                v-tooltip="'Связаться с автором'" 
                aria-label="Связаться"
              >
                Связаться
              </v-btn>
              <v-btn 
                text 
                to="/login" 
                style="min-width: 100px;" 
                v-else 
                v-tooltip="'Войти для действий'" 
                aria-label="Войти"
              >
                Войти для действий
              </v-btn>
            </div>
            <v-divider class="my-4" />
            <h3 style="font-size: 24px; font-weight: bold; color: #2E7D32; margin: 16px 0;" aria-label="Отзывы">Отзывы</h3>
            <v-list dense style="margin: 16px 0;" aria-label="Список отзывов">
              <v-list-item v-for="(review, index) in courseStore.reviews" :key="index" aria-label="Отзыв">
                <v-list-item-content>
                  <v-list-item-title style="font-size: 16px; color: #212529;" aria-label="Содержание отзыва">{{ review.content }}</v-list-item-title>
                  <v-list-item-subtitle style="font-size: 14px; color: #6C757D;" aria-label="Автор отзыва">Автор #{{ review.author_id }}</v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
            </v-list>
            <v-btn 
              color="secondary" 
              style="margin: 16px 0; border-radius: 8px;" 
              @click="openReviewModal" 
              v-if="courseStore.canReview" 
              v-tooltip="'Оставить отзыв'" 
              aria-label="Оставить отзыв"
            >
              Оставить отзыв
            </v-btn>
          </v-card-text>
          <v-card-actions>
            <v-btn 
              color="primary" 
              style="margin: 16px 0; border-radius: 8px;" 
              @click="goBack" 
              v-tooltip="'Вернуться назад'" 
              aria-label="Вернуться"
            >
              Вернуться
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
    <v-dialog v-model="dialogPayment" max-width="400" aria-label="Диалог оплаты">
      <v-card style="border-radius: 8px; padding: 24px;">
        <v-card-title style="font-size: 24px; color: #212529;" aria-label="Оплата">Оплата</v-card-title>
        <v-card-text>
          <v-form>
            <v-text-field 
              label="Номер карты" 
              v-model="cardNum" 
              :rules="[v => !!v || 'Поле обязательно']" 
              style="margin-bottom: 16px;" 
              aria-label="Номер карты" 
            />
            <v-text-field 
              label="Срок действия" 
              v-model="expiry" 
              :rules="[v => !!v || 'Поле обязательно']" 
              style="margin-bottom: 16px;" 
              aria-label="Срок действия карты" 
            />
            <v-text-field 
              label="CVV" 
              v-model="cvv" 
              :rules="[v => !!v || 'Поле обязательно']" 
              style="margin-bottom: 16px;" 
              aria-label="CVV код" 
            />
          </v-form>
        </v-card-text>
        <v-card-actions style="justify-content: flex-end; gap: 8px;">
          <v-btn 
            color="primary" 
            @click="submitPayment" 
            v-tooltip="'Оплатить'" 
            aria-label="Оплатить"
          >
            Оплатить
          </v-btn>
          <v-btn 
            text 
            @click="dialogPayment = false" 
            v-tooltip="'Отмена оплаты'" 
            aria-label="Отмена"
          >
            Отмена
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="dialogReview" max-width="400" aria-label="Диалог отзыва">
      <v-card style="border-radius: 8px; padding: 24px;">
        <v-card-title style="font-size: 24px; color: #212529;" aria-label="Оставить отзыв">Оставить отзыв</v-card-title>
        <v-card-text>
          <v-textarea 
            v-model="reviewText" 
            label="Ваш отзыв" 
            :rules="[v => !!v || 'Отзыв обязателен']" 
            style="margin-bottom: 16px;" 
            aria-label="Ваш отзыв" 
          />
        </v-card-text>
        <v-card-actions style="justify-content: flex-end; gap: 8px;">
          <v-btn 
            color="primary" 
            @click="submitReview" 
            v-tooltip="'Отправить отзыв'" 
            aria-label="Отправить отзыв"
          >
            Отправить
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-snackbar v-model="snackbar" :color="snackbarColor" timeout="3000" aria-label="Уведомление о результате">
      {{ snackbarText }}
    </v-snackbar>
  </v-container>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRuntimeConfig } from 'nuxt/app'
import { useRoute, useRouter } from 'vue-router'
import { useNuxtApp } from 'nuxt/app'
import { useCourseStore } from '~/stores/course'
import { useAuthStore } from '~/stores/auth'

const { $emitter } = useNuxtApp()
const emitter = $emitter
const config = useRuntimeConfig()
const route = useRoute()
const router = useRouter()
const courseStore = useCourseStore()
const authStore = useAuthStore()
const token = ref(authStore.token)
const userId = ref(authStore.userId)
const isLoggedIn = ref(authStore.isLoggedIn)
const role = ref(authStore.role)
const dialogPayment = ref(false)
const dialogReview = ref(false)
const cardNum = ref('')
const expiry = ref('')
const cvv = ref('')
const reviewText = ref('')
const snackbar = ref(false)
const snackbarText = ref('')
const snackbarColor = ref('success')

onMounted(async () => {
  if (process.client) {
    token.value = authStore.token
    userId.value = authStore.userId
    isLoggedIn.value = authStore.isLoggedIn
    role.value = authStore.role
  }
  try {
    await courseStore.loadCourse(route.params.id)
    if (isLoggedIn.value) {
      await courseStore.checkCanReview(userId.value)
    }
  } catch (error) {
    courseStore.error = 'Ошибка загрузки курса: ' + (error.message || 'Неизвестная ошибка')
    snackbarText.value = courseStore.error
    snackbarColor.value = 'error'
    snackbar.value = true
  }
  emitter.on('message', courseStore.handleNewMessage)
})

onUnmounted(() => {
  emitter.off('message', courseStore.handleNewMessage)
})

watch(() => route.params.id, async (newId) => {
  try {
    await courseStore.loadCourse(newId)
    if (isLoggedIn.value) {
      await courseStore.checkCanReview(userId.value)
    }
  } catch (error) {
    courseStore.error = 'Ошибка загрузки курса: ' + (error.message || 'Неизвестная ошибка')
    snackbarText.value = courseStore.error
    snackbarColor.value = 'error'
    snackbar.value = true
  }
})

const goBack = () => {
  router.back()
}

const openPayment = () => {
  if (!isLoggedIn.value) {
    router.push('/login')
    return
  }
  dialogPayment.value = true
}

const submitPayment = async () => {
  if (!cardNum.value || !expiry.value || !cvv.value) {
    snackbarText.value = 'Заполните все поля'
    snackbarColor.value = 'error'
    snackbar.value = true
    return
  }
  try {
    await courseStore.submitPayment(route.params.id, userId.value)
    snackbarText.value = `Оплата успешна! ID: ${courseStore.transactionId || 'unknown'}`
    snackbarColor.value = 'success'
    snackbar.value = true
    await courseStore.loadCourse(route.params.id)
  } catch (error) {
    snackbarText.value = 'Ошибка оплаты: ' + (error.message || 'Неизвестная ошибка')
    snackbarColor.value = 'error'
    snackbar.value = true
  } finally {
    dialogPayment.value = false
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
    console.log('API Response:', data) // Временное логирование для отладки
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