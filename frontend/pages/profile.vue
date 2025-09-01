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
            <v-row v-if="profile.avatar_url" class="mb-4">
              <v-col cols="12" class="text-center">
                <v-img
                  :src="profile.avatar_url || '/images/nutri-placeholder.jpg'"
                  max-width="120"
                  max-height="120"
                  class="mx-auto rounded-circle"
                  alt="Аватар пользователя"
                  @error="profile.avatar_url = '/images/nutri-placeholder.jpg'"
                />
              </v-col>
            </v-row>
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
              <v-file-input
                v-model="avatarFile"
                label="Загрузить аватар (JPG/PNG, до 5MB)"
                accept="image/jpeg,image/png"
                :rules="[v => !v || v.size < 5 * 1024 * 1024 || 'Файл должен быть меньше 5MB']"
                aria-label="Загрузка аватара"
              />
              <v-btn 
                color="primary" 
                type="submit" 
                v-tooltip="'Сохранить изменения'" 
                aria-label="Сохранить изменения профиля"
                :disabled="loading"
              >
                Сохранить
              </v-btn>
            </v-form>
            <v-row v-else>
              <v-col cols="12">
                <h3 style="font-size: 18px; color: #212529;" aria-label="Имя пользователя">{{ profile.full_name || profile.username || 'Имя не указано' }}</h3>
                <p v-if="profile.role === 'nutri'" style="font-size: 15px; color: #6C757D;" aria-label="Роль пользователя">Роль: {{ profile.role === 'nutri' ? 'Нутрициолог' : 'Клиент' }}</p>
              </v-col>
              <v-col cols="12" v-if="profile.role === 'nutri'">
                <h4 style="font-size: 16px; color: #2E7D32;" aria-label="Услуги">Услуги</h4>
                <p v-if="profile.description" style="font-size: 15px; line-height: 1.5;" aria-label="Описание">{{ profile.description }}</p>
                <p v-else style="font-size: 15px; color: #6C757D;" aria-label="Описание не указано">Описание не указано</p>
              </v-col>
              <v-col cols="12" v-if="role === 'nutri' && !otherId">
                <h4 style="font-size: 16px; color: #2E7D32;" aria-label="Баланс">Баланс: {{ profile.balance || 0 }} руб.</h4>
              </v-col>
              <v-col cols="12" v-if="role === 'nutri' && !otherId">
                <h4 style="font-size: 16px; color: #2E7D32;" aria-label="Реквизиты">Реквизиты</h4>
                <v-form @submit.prevent="updateCard">
                  <v-text-field 
                    v-model="cardNumber" 
                    label="Номер карты" 
                    type="text" 
                    :rules="[v => v && v.length === 16 && /^\d+$/.test(v) || 'Номер карты: 16 цифр']" 
                    aria-label="Номер карты" 
                  />
                  <v-btn color="primary" type="submit" aria-label="Обновить карту" :disabled="loading">Обновить карту</v-btn>
                </v-form>
              </v-col>
              <v-col cols="12">
                <h4 style="font-size: 16px; color: #2E7D32;" aria-label="Услуги">Услуги</h4>
                <v-row v-if="courses.length > 0" class="mt-4">
                  <v-col v-for="course in sortedCourses" :key="course.id" cols="12" sm="6" md="4">
                    <NuxtLink :to="`/courses/${course.id}`" style="text-decoration: none; display: block;">
                      <v-card 
                        class="pa-4" 
                        style="border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);"
                        aria-label="Карточка услуги"
                      >
                        <v-card-title class="text-h6" style="font-size: 18px; color: #212529;" aria-label="Название услуги">{{ course.title }}</v-card-title>
                        <v-card-subtitle style="font-size: 15px; color: #6C757D;" aria-label="Описание услуги">{{ course.description ? course.description.substring(0, 50) + '...' : 'Нет описания' }}</v-card-subtitle>
                        <div style="font-size: 16px; color: #28A745; font-weight: bold;" aria-label="Цена услуги">
                          Чистая: {{ course.net_price }} руб., Итоговая: {{ course.gross_price }} руб.
                        </div>
                      </v-card>
                    </NuxtLink>
                  </v-col>
                </v-row>
                <p v-else aria-label="Нет услуг">Нет услуг</p>
                <v-btn 
                  v-if="role === 'nutri' && !otherId" 
                  color="primary" 
                  to="/courses/create" 
                  block 
                  class="mt-4" 
                  v-tooltip="'Создать новую услугу'" 
                  aria-label="Создать услугу"
                >
                  Создать услугу
                </v-btn>
              </v-col>
              <v-col cols="12" v-if="role === 'client' && !otherId">
                <h4 style="font-size: 16px; color: #2E7D32;" aria-label="Записанные услуги">Записанные услуги</h4>
                <v-row v-if="enrolled.length > 0">
                  <v-col v-for="course in enrolled" :key="course.id" cols="12" sm="6" md="4">
                    <NuxtLink :to="`/courses/${course.id}`" style="text-decoration: none; display: block;">
                      <v-card 
                        class="pa-4" 
                        style="border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);"
                        aria-label="Карточка записанной услуги"
                      >
                        <v-card-title class="text-h6" style="font-size: 18px;" aria-label="Название услуги">{{ course.course?.title || 'Без названия' }}</v-card-title>
                        <v-card-subtitle style="font-size: 15px;" aria-label="Описание услуги">{{ course.course?.description ? course.course.description.substring(0, 50) + '...' : 'Нет описания' }}</v-card-subtitle>
                      </v-card>
                    </NuxtLink>
                  </v-col>
                </v-row>
                <p v-else aria-label="Нет записей">Нет записей</p>
              </v-col>
              <v-col cols="12">
                <h4 style="font-size: 16px; color: #2E7D32;" aria-label="Отзывы">Отзывы</h4>
                <v-list v-if="reviews.length > 0" aria-label="Список отзывов">
                  <v-list-item v-for="(review, i) in reviews" :key="i">
                    <v-list-item-content>
                      <v-list-item-title aria-label="Содержание отзыва">{{ review.content }}</v-list-item-title>
                      <v-list-item-subtitle aria-label="Автор отзыва">{{ review.author?.full_name || review.author?.username || 'Аноним' }}</v-list-item-subtitle>
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
import { useChatStore } from '~/stores/chat'

const config = useRuntimeConfig()
const route = useRoute()
const router = useRouter()
const chatStore = useChatStore()
const profile = ref({})
const courses = ref([])
const enrolled = ref([])
const reviews = ref([])
const editMode = ref(false)
const cardNumber = ref('')
const avatarFile = ref(null)
const errorMessage = ref('')
const token = ref(null)
const role = ref('')
const userId = ref(null)
const otherId = ref(route.params.id || null)
const loading = ref(false)
const isLoggedIn = computed(() => !!token.value)

const sortedCourses = computed(() => [...courses.value].sort((a, b) => a.title.localeCompare(b.title) || a.id - b.id))

onMounted(async () => {
  if (process.client) {
    token.value = localStorage.getItem('token')
    role.value = localStorage.getItem('role')
    userId.value = localStorage.getItem('userId') ? parseInt(localStorage.getItem('userId')) : null
    chatStore.connectWebSocket()
  }
  await loadProfile()
})

watch(() => route.params.id, async (newId) => {
  otherId.value = newId || null
  await loadProfile()
})

async function loadProfile() {
  loading.value = true
  const headers = token.value ? { Authorization: `Bearer ${token.value}` } : {}
  const path = otherId.value ? `/api/profile/${otherId.value}` : '/api/profile'
  try {
    const data = await $fetch(`${config.public.apiBase}${path}`, { headers })
    profile.value = data.profile || {}
    courses.value = data.courses || []
    reviews.value = data.reviews || []
    if (!profile.value.id) {
      errorMessage.value = 'Данные профиля не загружены, проверьте базу данных'
      console.log('loadProfile: Пустой профиль', data)
    }
    if (!otherId.value && role.value === 'client') {
      const enrolledData = await $fetch(`${config.public.apiBase}/api/enrolled`, { headers })
      enrolled.value = enrolledData || []
    }
    if (profile.value.id) {
      localStorage.setItem('profile_user_id', profile.value.id)
      localStorage.setItem('profile_avatar_url', profile.value.avatar_url || '')
    }
  } catch (error) {
    console.log('loadProfile: Ошибка', error)
    errorMessage.value = 'Ошибка загрузки профиля: ' + (error.message || 'Неизвестная ошибка')
    if (error.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('role')
      localStorage.removeItem('userId')
      router.push('/login')
    }
  } finally {
    loading.value = false
  }
}

async function updateProfile() {
  loading.value = true
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
    if (avatarFile.value) {
      await uploadAvatar()
    }
    editMode.value = false
    await loadProfile()
  } catch (error) {
    console.log('updateProfile: Ошибка', error)
    errorMessage.value = 'Ошибка обновления профиля: ' + (error.message || 'Неизвестная ошибка')
  } finally {
    loading.value = false
  }
}

async function uploadAvatar() {
  if (!avatarFile.value) return
  const headers = { Authorization: `Bearer ${token.value}` }
  const formData = new FormData()
  formData.append('avatar', avatarFile.value)
  try {
    const response = await $fetch(`${config.public.apiBase}/api/profile/upload-avatar`, {
      method: 'POST',
      headers,
      body: formData
    })
    profile.value.avatar_url = response.profile?.avatar_url || profile.value.avatar_url
    errorMessage.value = 'Аватар загружен'
    // Отправка WebSocket-уведомления
    if (chatStore.websocket && chatStore.websocket.readyState === WebSocket.OPEN) {
      chatStore.websocket.send(JSON.stringify({
        type: 'avatar_updated',
        data: {
          user_id: userId.value,
          avatar_url: profile.value.avatar_url
        }
      }))
    }
    avatarFile.value = null
  } catch (error) {
    console.log('uploadAvatar: Ошибка', error)
    errorMessage.value = 'Ошибка загрузки аватара: ' + (error.message || 'Неизвестная ошибка')
  }
}

async function updateCard() {
  loading.value = true
  const headers = { Authorization: `Bearer ${token.value}` }
  const body = { card_number: cardNumber.value }
  try {
    await $fetch(`${config.public.apiBase}/api/profile/update-card`, { 
      method: 'POST', 
      headers, 
      body 
    })
    errorMessage.value = 'Карта обновлена'
    cardNumber.value = ''
  } catch (error) {
    console.log('updateCard: Ошибка', error)
    errorMessage.value = 'Ошибка обновления карты: ' + (error.message || 'Неизвестная ошибка')
  } finally {
    loading.value = false
  }
}

async function openChat() {
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
    console.log('openChat: Ошибка', error)
    errorMessage.value = 'Ошибка открытия чата: ' + (error.message || 'Неизвестная ошибка')
  }
}
</script>