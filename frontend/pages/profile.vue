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
              v-if="!otherId && role !== 'admin'" 
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
            <v-form v-if="editMode && role !== 'admin'" @submit.prevent="updateProfile">
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
              <v-col cols="12" v-if="profile.role === 'nutri' && role !== 'admin'">
                <h4 style="font-size: 16px; color: #2E7D32;" aria-label="Услуги">Услуги</h4>
                <p v-if="profile.description" style="font-size: 15px; line-height: 1.5;" aria-label="Описание">{{ profile.description }}</p>
                <p v-else style="font-size: 15px; color: #6C757D;" aria-label="Описание не указано">Описание не указано</p>
              </v-col>
              <v-col cols="12" v-if="isNutri && !otherId && role !== 'admin'">
                <h4 style="font-size: 16px; color: #2E7D32;" aria-label="Баланс">Баланс: {{ profile.balance }} руб.</h4>
              </v-col>
              <v-col cols="12" v-if="isNutri && !otherId && role !== 'admin'">
                <h4 style="font-size: 16px; color: #2E7D32;" aria-label="Форма для карты">Форма для карты</h4>
                <v-form @submit.prevent="updateCard">
                  <v-text-field
                    v-model="cardNumber"
                    label="Номер карты"
                    type="text"
                    :rules="[v => !!v || 'Номер карты обязателен', v => v.length === 16 || 'Должно быть 16 цифр', v => /^\d{16}$/.test(v) || 'Только цифры']"
                    required
                    aria-label="Номер карты"
                  />
                  <v-btn 
                    color="primary" 
                    type="submit" 
                    v-tooltip="'Обновить карту'" 
                    aria-label="Обновить карту"
                    :disabled="loading"
                  >
                    Обновить карту
                  </v-btn>
                </v-form>
              </v-col>
              <v-col cols="12" v-if="role === 'client' && role !== 'admin'">
                <h4 style="font-size: 16px; color: #2E7D32;" aria-label="Записи клиента">Мои записи</h4>
                <v-list v-if="enrolled.length > 0" aria-label="Список записей">
                  <v-list-item v-for="enrollment in enrolled" :key="enrollment.id" @click="goToCourse(enrollment.course_id)" style="cursor: pointer;">
                    <v-list-item-content>
                      <v-list-item-title aria-label="Название курса">{{ enrollment.course.title }}</v-list-item-title>
                      <v-list-item-subtitle aria-label="Дата записи">{{ formatDate(enrollment.created_at) }}</v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-list>
                <p v-else style="font-size: 15px; color: #6C757D;" aria-label="Нет записей">Нет записей</p>
              </v-col>
              <v-col cols="12" v-if="isNutri && !otherId && role !== 'admin'">
                <h4 style="font-size: 16px; color: #2E7D32;" aria-label="Мои курсы">Мои курсы</h4>
                <v-btn color="#28A745" to="/courses/create" v-tooltip="'Создать курс'" aria-label="Создать новый курс">Создать курс</v-btn>
                <v-list v-if="courses.length > 0" aria-label="Список курсов">
                  <v-list-item v-for="course in courses" :key="course.id" @click="goToCourse(course.id)" style="cursor: pointer;">
                    <v-list-item-content>
                      <v-list-item-title aria-label="Название курса">{{ course.title }}</v-list-item-title>
                      <v-list-item-subtitle aria-label="Чистая цена">Чистая: {{ course.net_price }} руб., Итоговая: {{ course.gross_price }} руб. (вкл. комиссия)</v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-list>
                <p v-else style="font-size: 15px; color: #6C757D;" aria-label="Нет курсов">Нет курсов</p>
              </v-col>
              <v-col cols="12" v-if="otherId && role !== 'admin'">
                <h4 style="font-size: 16px; color: #2E7D32;" aria-label="Курсы нутрициолога">Курсы нутрициолога</h4>
                <v-list v-if="courses.length > 0" aria-label="Список курсов">
                  <v-list-item v-for="course in sortedCourses" :key="course.id" @click="goToCourse(course.id)" style="cursor: pointer;">
                    <v-list-item-content>
                      <v-list-item-title aria-label="Название курса">{{ course.title }}</v-list-item-title>
                      <v-list-item-subtitle aria-label="Стоимость">{{ displayedPrice(course) }} руб.</v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-list>
                <p v-else style="font-size: 15px; color: #6C757D;" aria-label="Нет курсов">Нет курсов</p>
              </v-col>
              <v-col cols="12" v-if="role !== 'admin'">
                <h4 style="font-size: 16px; color: #2E7D32;" aria-label="Отзывы">Отзывы</h4>
                <v-row v-if="reviews.length > 0" class="mt-4">
                  <v-col v-for="(review, i) in reviews" :key="i" cols="12">
                    <v-card flat class="pa-3" style="border-radius: 8px; border: 1px solid #e0e0e0;">
                      <v-card-text style="font-size: 14px; color: #212529;">
                        "{{ review.content }}" — {{ review.author ? review.author.full_name : 'Аноним' }}
                      </v-card-text>
                    </v-card>
                  </v-col>
                </v-row>
                <p v-else style="font-size: 15px; color: #6C757D;" aria-label="Нет отзывов">Нет отзывов</p>
              </v-col>
              <v-col cols="12" v-if="role === 'admin' && !otherId">
                <v-btn color="primary" @click="goToAdmin" aria-label="Управление выплатами">Управление выплатами</v-btn>
              </v-col>
            </v-row>
          </v-card-text>
          <v-card-actions v-if="isLoggedIn && otherId && userId !== otherId" class="justify-center pa-4">
            <v-btn 
              color="#28A745" 
              variant="outlined" 
              @click="openChat" 
              v-tooltip="'Открыть чат'" 
              aria-label="Открыть чат"
            >
              <v-icon left>mdi-message-text</v-icon>
              Связаться
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useRuntimeConfig } from 'nuxt/app'
import { useChatStore } from '~/stores/chat'

const config = useRuntimeConfig()
const route = useRoute()
const router = useRouter()
const chatStore = useChatStore()
const profile = ref({})
const courses = ref([])
const reviews = ref([])
const enrolled = ref([])
const token = ref(null)
const role = ref('')
const userId = ref(null)
const otherId = ref(null)
const isLoggedIn = ref(false)
const editMode = ref(false)
const avatarFile = ref(null)
const cardNumber = ref('')
const loading = ref(false)
const errorMessage = ref('')

onMounted(async () => {
  if (process.client) {
    token.value = localStorage.getItem('token')
    role.value = localStorage.getItem('role') || ''
    userId.value = localStorage.getItem('userId') ? parseInt(localStorage.getItem('userId')) : null
    isLoggedIn.value = !!token.value
    otherId.value = route.params.id ? parseInt(route.params.id) : null
    console.log('Profile.vue mounted, role:', role.value, 'userId:', userId.value, 'otherId:', otherId.value)
    if (!token.value) {
      console.log('Profile.vue: Нет токена, редирект на /login')
      router.push('/login')
      return
    }
  }
  await loadProfile()
})

watch(() => route.params.id, async (newId) => {
  otherId.value = newId ? parseInt(newId) : null
  console.log('Profile.vue: Изменен otherId:', otherId.value)
  await loadProfile()
})

const isNutri = computed(() => role.value === 'nutri')
const sortedCourses = computed(() => courses.value.sort((a, b) => a.title.localeCompare(b.title)))

const displayedPrice = (course) => {
  return isNutri.value ? `Чистая: ${course.net_price} руб., Итоговая: ${course.gross_price} руб.` : course.gross_price
}

const formatDate = (date) => new Date(date).toLocaleDateString('ru-RU')

const goToCourse = (id) => {
  console.log('Profile.vue: Переход на курс:', id)
  router.push(`/courses/${id}`)
}

const goToAdmin = () => {
  console.log('Profile.vue: Нажата кнопка Управление выплатами, переход на /admin')
  router.push('/admin')
}

const loadProfile = async () => {
  loading.value = true
  const headers = token.value ? { Authorization: `Bearer ${token.value}` } : {}
  const endpoint = otherId.value ? `/api/profile/${otherId.value}` : '/api/profile'
  try {
    console.log('Profile.vue: Загрузка профиля, endpoint:', endpoint)
    const data = await $fetch(`${config.public.apiBase}${endpoint}`, { headers })
    profile.value = data.profile || {}
    courses.value = data.courses || []
    reviews.value = data.reviews || []
    if (isLoggedIn.value && role.value === 'client') {
      console.log('Profile.vue: Загрузка записей для клиента')
      const enrollData = await $fetch(`${config.public.apiBase}/api/enrolled`, { headers })
      enrolled.value = enrollData || []
    }
    if (process.client) {
      localStorage.setItem('profile_user_id', profile.value.id)
      localStorage.setItem('profile_avatar_url', profile.value.avatar_url || '')
    }
  } catch (error) {
    console.error('Profile.vue: Ошибка загрузки профиля:', error)
    errorMessage.value = 'Ошибка загрузки профиля: ' + (error.message || 'Неизвестная ошибка')
    if (error.status === 401) {
      console.log('Profile.vue: 401 Unauthorized, удаление токена и редирект на /login')
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
    console.log('Profile.vue: Обновление профиля, body:', body)
    await $fetch(`${config.public.apiBase}/api/profile`, { 
      method: 'PUT', 
      headers, 
      body 
    })
    if (avatarFile.value) {
      console.log('Profile.vue: Загрузка аватара')
      await uploadAvatar()
    }
    editMode.value = false
    await loadProfile()
    errorMessage.value = 'Профиль обновлен'
  } catch (error) {
    console.error('Profile.vue: Ошибка обновления профиля:', error)
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
    console.log('Profile.vue: Загрузка аватара')
    const response = await $fetch(`${config.public.apiBase}/api/profile/upload-avatar`, {
      method: 'POST',
      headers,
      body: formData
    })
    profile.value.avatar_url = response.profile?.avatar_url || profile.value.avatar_url
    errorMessage.value = 'Аватар загружен'
    if (chatStore.websocket && chatStore.websocket.readyState === WebSocket.OPEN) {
      console.log('Profile.vue: Отправка WebSocket уведомления об обновлении аватара')
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
    console.error('Profile.vue: Ошибка загрузки аватара:', error)
    errorMessage.value = 'Ошибка загрузки аватара: ' + (error.message || 'Неизвестная ошибка')
  }
}

async function updateCard() {
  loading.value = true
  const headers = { Authorization: `Bearer ${token.value}` }
  const body = { card_number: cardNumber.value }
  try {
    console.log('Profile.vue: Обновление карты, card_number:', cardNumber.value)
    await $fetch(`${config.public.apiBase}/api/profile/update-card`, { 
      method: 'POST', 
      headers, 
      body 
    })
    errorMessage.value = 'Карта обновлена успешно'
    cardNumber.value = ''
    await loadProfile()
  } catch (error) {
    console.error('Profile.vue: Ошибка обновления карты:', error)
    errorMessage.value = 'Ошибка обновления карты: ' + (error.response?.data?.error || error.message || 'Неизвестная ошибка')
  } finally {
    loading.value = false
  }
}

async function openChat() {
  if (!isLoggedIn.value) {
    console.log('Profile.vue: Пользователь не авторизован, редирект на /login')
    router.push('/login')
    return
  }
  try {
    console.log('Profile.vue: Открытие чата с receiver_id:', profile.value.id)
    const headers = { Authorization: `Bearer ${token.value}` }
    const data = await $fetch(`${config.public.apiBase}/api/start-chat`, {
      method: 'POST',
      headers,
      body: { receiver_id: profile.value.id }
    })
    router.push(`/chats?selected=${data.receiver_id}`)
  } catch (error) {
    console.error('Profile.vue: Ошибка открытия чата:', error)
    errorMessage.value = 'Ошибка открытия чата: ' + (error.message || 'Неизвестная ошибка')
  }
}
</script>