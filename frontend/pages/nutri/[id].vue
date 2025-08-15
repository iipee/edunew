<template>
  <v-container>
    <v-row justify="center">
      <v-col cols="12" sm="8" md="6">
        <v-card class="pa-6" elevation="2" style="border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);" aria-label="Профиль нутрициолога">
          <v-card-title class="justify-center text-h4" style="font-size: 24px; font-weight: bold; color: #212529;" aria-label="Заголовок профиля">
            Профиль нутрициолога
          </v-card-title>
          <v-card-text>
            <v-alert v-if="errorMessage" type="error" dismissible class="mb-4" aria-label="Сообщение об ошибке">
              {{ errorMessage }}
            </v-alert>
            <v-row>
              <v-col cols="12">
                <h3 style="font-size: 20px; color: #212529;" aria-label="Имя нутрициолога">{{ profile.full_name || profile.username || 'Имя не указано' }}</h3>
              </v-col>
              <v-col cols="12">
                <h4 style="font-size: 18px; color: #2E7D32;" aria-label="Услуги">Услуги</h4>
                <p v-if="profile.description" style="font-size: 16px; line-height: 1.5;" aria-label="Описание">{{ profile.description }}</p>
                <p v-else style="font-size: 16px; color: #6C757D;" aria-label="Описание не указано">Описание не указано</p>
              </v-col>
              <v-col cols="12">
                <h4 style="font-size: 18px; color: #2E7D32;" aria-label="Курсы">Курсы</h4>
                <v-row v-if="courses.length > 0" class="mt-4">
                  <v-col v-for="course in sortedCourses" :key="course.id" cols="12" sm="6" md="4">
                    <NuxtLink :to="`/courses/${course.id}`" style="text-decoration: none; display: block;">
                      <v-card 
                        class="v-card v-theme--light v-card--density-default v-card--variant-elevated pa-4" 
                        style="border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); max-width: 366px; min-height: 100px; width: 100%; display: flex; flex-direction: column;"
                        aria-label="Карточка курса"
                      >
                        <div class="v-row v-row--no-gutters align-center">
                          <div class="v-col">
                            <div class="v-card-title text-h6 pa-0" 
                              style="font-size: 20px; color: #212529; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; padding-bottom: 4px;"
                              aria-label="Название курса"
                            >
                              {{ course.title }}
                            </div>
                          </div>
                          <div class="v-col v-col-auto">
                            <span style="font-size: 16px; color: #28A745; font-weight: bold; padding-left: 8px;">
                              {{ course.price }} руб.
                            </span>
                          </div>
                        </div>
                        <div class="v-card-text" 
                          style="font-size: 16px; line-height: 1.5; color: #6C757D; overflow: hidden; text-overflow: ellipsis; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; line-clamp: 2;"
                        >
                          {{ course.description }}
                        </div>
                      </v-card>
                    </NuxtLink>
                  </v-col>
                </v-row>
                <p v-else style="font-size: 16px; color: #6C757D;" aria-label="Курсов нет">Курсов нет</p>
              </v-col>
              <v-col cols="12">
                <h4 style="font-size: 18px; color: #2E7D32;" aria-label="Отзывы">Отзывы</h4>
                <v-list v-if="reviews.length > 0" aria-label="Список отзывов">
                  <v-list-item v-for="review in reviews" :key="review.id" aria-label="Отзыв">
                    <v-list-item-title style="font-size: 16px;" aria-label="Содержание отзыва">{{ review.content }}</v-list-item-title>
                    <v-list-item-subtitle style="font-size: 14px; color: #6C757D;" aria-label="Автор отзыва">— Пользователь #{{ review.author_id }}</v-list-item-subtitle>
                  </v-list-item>
                </v-list>
                <p v-else style="font-size: 16px; color: #6C757D;" aria-label="Отзывов нет">Отзывов нет</p>
              </v-col>
            </v-row>
            <v-btn 
              v-if="profile.role === 'nutri' && isLoggedIn && role === 'client'" 
              color="#28A745" 
              block 
              @click="openChat" 
              v-tooltip="'Связаться с нутрициологом'" 
              aria-label="Связаться с нутрициологом"
            >
              Связаться
            </v-btn>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    <v-snackbar v-model="snackbar" color="error" timeout="3000" aria-label="Уведомление об ошибке">
      {{ errorMessage }}
    </v-snackbar>
  </v-container>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRuntimeConfig } from 'nuxt/app'
import { useRoute, useRouter } from 'vue-router'
import { useNuxtApp } from 'nuxt/app'

const { $emitter } = useNuxtApp()
const emitter = $emitter
const config = useRuntimeConfig()
const route = useRoute()
const router = useRouter()
const otherId = ref(route.params.id || null)
const profile = ref({})
const courses = ref([])
const enrolled = ref([])
const reviews = ref([])
const token = ref(null)
const role = ref('')
const userId = ref(null)
const editMode = ref(false)
const errorMessage = ref('')

const sortedCourses = computed(() => {
  return [...courses.value].sort((a, b) => a.title.localeCompare(b.title) || a.id - b.id)
})

const isLoggedIn = computed(() => !!token.value)

onMounted(async () => {
  if (process.client) {
    token.value = localStorage.getItem('token')
    role.value = localStorage.getItem('role') || ''
    userId.value = localStorage.getItem('userId') ? parseInt(localStorage.getItem('userId')) : null
  }
  await loadProfile()
  emitter.on('login', loadProfile)
})

watch(courses, () => {
  loadProfile()
})

const loadProfile = async () => {
  const headers = token.value ? { Authorization: `Bearer ${token.value}` } : {}
  const path = otherId.value ? `/api/profile/${otherId.value}` : '/api/profile'
  try {
    const data = await $fetch(`${config.public.apiBase}${path}`, { headers })
    profile.value = data.profile || {}
    courses.value = data.courses || []
    const reviewPath = otherId.value ? `/api/reviews/user/${otherId.value}` : `/api/reviews/user/${userId.value}`
    const reviewData = await $fetch(`${config.public.apiBase}${reviewPath}`, { headers })
    reviews.value = reviewData || []
    if (!otherId.value && role.value === 'client') {
      const enrolledData = await $fetch(`${config.public.apiBase}/api/enrolled`, { headers })
      enrolled.value = enrolledData || []
    }
  } catch (error) {
    errorMessage.value = 'Ошибка загрузки профиля: ' + (error.message || 'Неизвестная ошибка')
    if (error.statusCode === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('role')
      localStorage.removeItem('userId')
      router.push('/login')
    }
  }
}

const updateProfile = async () => {
  const headers = { Authorization: `Bearer ${token.value}` }
  const body = {
    full_name: profile.value.full_name,
    description: profile.value.description
  }
  try {
    await $fetch(`${config.public.apiBase}/api/profile`, { 
      method: 'PUT', 
      headers, 
      body 
    })
    editMode.value = false
    await loadProfile()
  } catch (error) {
    errorMessage.value = 'Ошибка обновления профиля: ' + (error.message || 'Неизвестная ошибка')
  }
}

const openChat = async () => {
  if (!isLoggedIn.value) {
    router.push('/login')
    return
  }
  try {
    const headers = { Authorization: `Bearer ${token.value}` }
    const data = await $fetch(`${config.public.apiBase}/api/start-chat`, {
      method: 'POST',
      headers,
      body: { receiver_id: profile.value.id }
    })
    router.push(`/chats?selected=${data.receiver_id}`)
  } catch (error) {
    errorMessage.value = 'Ошибка открытия чата: ' + (error.message || 'Неизвестная ошибка')
  }
}
</script>