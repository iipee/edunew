<template>
  <v-container>
    <v-row justify="center">
      <v-col cols="12" sm="8" md="6">
        <v-card class="pa-6" elevation="2" style="border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);" aria-label="Профиль пользователя">
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
                <h3 style="font-size: 20px; color: #212529;" aria-label="Имя пользователя">{{ profile.full_name || profile.username || 'Имя не указано' }}</h3>
              </v-col>
              <v-col cols="12" v-if="profile.role === 'nutri'">
                <h4 style="font-size: 18px; color: #2E7D32;" aria-label="Услуги">Услуги</h4>
                <p v-if="profile.description" style="font-size: 16px; line-height: 1.5;" aria-label="Описание">{{ profile.description }}</p>
                <p v-else style="font-size: 16px; color: #6C757D;" aria-label="Описание не указано">Описание не указано</p>
              </v-col>
              <v-col cols="12" v-if="role === 'nutri' && !otherId">
                <h4 style="font-size: 18px; color: #2E7D32;" aria-label="Баланс">Баланс: {{ profile.balance || 0 }} руб.</h4>
              </v-col>
              <v-col cols="12" v-if="role === 'nutri' && !otherId">
                <h4 style="font-size: 18px; color: #2E7D32;" aria-label="Реквизиты">Реквизиты</h4>
                <v-form @submit.prevent="updateCard">
                  <v-text-field 
                    v-model="cardNumber" 
                    label="Номер карты" 
                    type="text" 
                    :rules="[v => v.length === 16 && /^\d+$/.test(v) || 'Номер карты: 16 цифр']" 
                    aria-label="Номер карты" 
                  />
                  <v-btn color="primary" type="submit" aria-label="Обновить карту">Обновить карту</v-btn>
                </v-form>
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
                              style="font-size: 20px; color: #212529; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;"
                              aria-label="Название курса"
                            >
                              {{ course.title }}
                            </div>
                            <div class="v-card-subtitle pa-0 mt-2" 
                              style="font-size: 16px; color: #6C757D; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;"
                              aria-label="Описание курса"
                            >
                              {{ course.description.slice(0, 50) }}...
                            </div>
                            <div class="mt-2" 
                              style="font-size: 18px; color: #28A745; font-weight: bold;"
                              aria-label="Цена курса"
                            >
                              Чистая: {{ course.net_price }} руб., Итоговая: {{ course.gross_price }} руб.
                            </div>
                          </div>
                        </div>
                      </v-card>
                    </NuxtLink>
                  </v-col>
                </v-row>
                <p v-else aria-label="Нет курсов">Нет курсов</p>
                <v-btn 
                  v-if="role === 'nutri' && !otherId" 
                  color="primary" 
                  to="/courses/create" 
                  block 
                  class="mt-4" 
                  v-tooltip="'Создать новый курс'" 
                  aria-label="Создать курс"
                >
                  Создать курс
                </v-btn>
              </v-col>
              <v-col cols="12" v-if="role === 'client' && !otherId">
                <h4 style="font-size: 18px; color: #2E7D32;" aria-label="Записанные курсы">Записанные курсы</h4>
                <v-row v-if="enrolled.length > 0">
                  <v-col v-for="course in enrolled" :key="course.id" cols="12" sm="6" md="4">
                    <NuxtLink :to="`/courses/${course.id}`" style="text-decoration: none; display: block;">
                      <v-card 
                        class="pa-4" 
                        style="border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);"
                        aria-label="Карточка записанного курса"
                      >
                        <v-card-title class="text-h6" aria-label="Название курса">{{ course.title }}</v-card-title>
                        <v-card-subtitle aria-label="Описание курса">{{ course.description.slice(0, 50) }}...</v-card-subtitle>
                      </v-card>
                    </NuxtLink>
                  </v-col>
                </v-row>
                <p v-else aria-label="Нет записей">Нет записей</p>
              </v-col>
              <v-col cols="12">
                <h4 style="font-size: 18px; color: #2E7D32;" aria-label="Отзывы">Отзывы</h4>
                <v-list v-if="reviews.length > 0" aria-label="Список отзывов">
                  <v-list-item v-for="(review, i) in reviews" :key="i">
                    <v-list-item-content>
                      <v-list-item-title aria-label="Содержание отзыва">{{ review.content }}</v-list-item-title>
                    </v-list-item-content>
                  </v-list-item>
                </v-list>
                <p v-else aria-label="Нет отзывов">Нет отзывов</p>
              </v-col>
              <v-col cols="12" v-if="otherId && isLoggedIn">
                <v-btn 
                  color="primary" 
                  block 
                  @click="openChat" 
                  v-tooltip="'Открыть чат'" 
                  aria-label="Открыть чат"
                >
                  Открыть чат
                </v-btn>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter, useRuntimeConfig } from 'nuxt/app'

const config = useRuntimeConfig()
const route = useRoute()
const router = useRouter()
const profile = ref({})
const courses = ref([])
const enrolled = ref([])
const reviews = ref([])
const editMode = ref(false)
const cardNumber = ref('')
const errorMessage = ref('')
const token = ref(null)
const role = ref('')
const userId = ref(null)
const otherId = ref(route.params.id || null)
const isLoggedIn = computed(() => !!token.value)

const sortedCourses = computed(() => [...courses.value].sort((a, b) => a.title.localeCompare(b.title) || a.id - b.id))

onMounted(async () => {
  if (process.client) {
    token.value = localStorage.getItem('token')
    role.value = localStorage.getItem('role')
    userId.value = localStorage.getItem('userId') ? parseInt(localStorage.getItem('userId')) : null
  }
  await loadProfile()
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

const updateCard = async () => {
  const headers = { Authorization: `Bearer ${token.value}` }
  const body = { card_number: cardNumber.value }
  try {
    await $fetch(`${config.public.apiBase}/api/profile/update-card`, { 
      method: 'POST', 
      headers, 
      body 
    })
    errorMessage.value = 'Карта обновлена'
  } catch (error) {
    errorMessage.value = 'Ошибка обновления карты: ' + (error.message || 'Неизвестная ошибка')
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