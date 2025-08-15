<template>
  <v-container>
    <v-row justify="center">
      <v-col cols="12" sm="8" md="6">
        <v-card class="pa-6" aria-label="Профиль пользователя">
          <v-card-title class="justify-center">
            <h2 aria-label="Заголовок профиля">Профиль</h2>
            <v-btn 
              icon 
              @click="editMode = !editMode" 
              v-if="!otherId" 
              v-tooltip="'Редактировать профиль'" 
              aria-label="Редактировать профиль"
            >
              <v-icon>mdi-pencil</v-icon>
            </v-btn>
          </v-card-title>
          <v-card-text>
            <v-alert v-if="errorMessage" type="error" dismissible class="mb-4" aria-label="Сообщение об ошибке">
              {{ errorMessage }}
            </v-alert>
            <v-form v-if="editMode" @submit.prevent="updateProfile">
              <v-text-field 
                v-model="profile.full_name" 
                label="Полное имя" 
                :rules="[v => !!v || 'Полное имя обязательно']" 
                aria-label="Полное имя" 
              />
              <v-textarea 
                v-model="profile.description" 
                label="Описание услуг" 
                aria-label="Описание услуг" 
              />
              <v-btn 
                color="primary" 
                type="submit" 
                v-tooltip="'Сохранить изменения'" 
                aria-label="Сохранить изменения профиля"
              >
                Сохранить
              </v-btn>
            </v-form>
            <v-row v-else>
              <v-col cols="12">
                <h3 aria-label="Имя пользователя">{{ profile.full_name || profile.username || 'Имя не указано' }}</h3>
              </v-col>
              <v-col cols="12" v-if="profile.role === 'nutri'">
                <h4 aria-label="Услуги">Услуги</h4>
                <p v-if="profile.description" aria-label="Описание">{{ profile.description }}</p>
                <p v-else aria-label="Описание не указано">Описание не указано</p>
              </v-col>
              <v-col cols="12" v-if="profile.role === 'nutri'">
                <h4 aria-label="Мои курсы">Мои курсы</h4>
                <v-btn 
                  color="primary" 
                  to="/courses/create" 
                  v-if="courses.length === 0 && !otherId && role === 'nutri'" 
                  v-tooltip="'Создать новый курс'" 
                  aria-label="Создать курс"
                >
                  Создать курс
                </v-btn>
                <v-list v-else aria-label="Список курсов">
                  <v-list-item v-for="(course, index) in sortedCourses" :key="index" aria-label="Курс">
                    <v-list-item-title aria-label="Название курса">{{ course.title }} - {{ course.price }} руб.</v-list-item-title>
                    <v-list-item-subtitle aria-label="Описание курса">{{ course.description }}</v-list-item-subtitle>
                  </v-list-item>
                </v-list>
              </v-col>
              <v-col cols="12" v-if="role === 'client' && !otherId">
                <h4 aria-label="Записанные курсы">Записанные курсы</h4>
                <v-list v-if="enrolled.length > 0" aria-label="Список записанных курсов">
                  <v-list-item v-for="(course, index) in enrolled" :key="index" aria-label="Записанный курс">
                    <v-list-item-title aria-label="Название курса">{{ course.title }}</v-list-item-title>
                  </v-list-item>
                </v-list>
                <p v-else aria-label="Нет записанных курсов">Нет записанных курсов</p>
              </v-col>
              <v-col cols="12">
                <h4 aria-label="Отзывы">Отзывы</h4>
                <v-list v-if="reviews.length > 0" aria-label="Список отзывов">
                  <v-list-item v-for="(review, index) in reviews" :key="index" aria-label="Отзыв">
                    <v-list-item-title aria-label="Содержание отзыва">{{ review.content }}</v-list-item-title>
                    <v-list-item-subtitle aria-label="Автор отзыва">— Пользователь #{{ review.author_id }}</v-list-item-subtitle>
                  </v-list-item>
                </v-list>
                <p v-else aria-label="Отзывов нет">Отзывов нет</p>
              </v-col>
            </v-row>
          </v-card-text>
          <v-btn 
            v-if="profile.role === 'nutri' && isLoggedIn && role === 'client'" 
            color="primary" 
            block 
            @click="openChat" 
            v-tooltip="'Связаться с нутрициологом'" 
            aria-label="Связаться с нутрициологом"
          >
            Связаться
          </v-btn>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue'
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

watch(() => route.params.id, async (newId) => {
  otherId.value = newId || null
  await loadProfile()
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
    if (error.status === 401) {
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