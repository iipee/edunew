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
                label="Описание" 
                aria-label="Описание" 
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
                <h4 style="font-size: 18px; color: #2E7D32;" aria-label="Описание">Описание</h4>
                <p v-if="profile.description" style="font-size: 16px; line-height: 1.5;" aria-label="Описание">{{ profile.description }}</p>
                <p v-else style="font-size: 16px; color: #6C757D;" aria-label="Описание не указано">Описание не указано</p>
              </v-col>
              <v-col cols="12" v-if="role === 'nutri' && !otherId">
                <h4 style="font-size: 18px; color: #2E7D32;" aria-label="Услуги">Услуги</h4>
                <v-row v-if="courses.length > 0" class="mt-4">
                  <v-col v-for="course in courses" :key="course.id" cols="12" sm="6" md="4">
                    <NuxtLink :to="`/courses/${course.id}`" style="text-decoration: none; display: block;">
                      <v-card 
                        class="pa-4" 
                        style="border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);"
                        aria-label="Карточка услуги"
                      >
                        <v-card-title style="font-size: 18px; color: #212529;">{{ course.title }}</v-card-title>
                        <v-card-subtitle style="color: #6C757D;">Чистая цена: {{ course.net_price }} руб.</v-card-subtitle>
                        <v-card-text>{{ course.description.substring(0, 100) }}...</v-card-text>
                      </v-card>
                    </NuxtLink>
                  </v-col>
                </v-row>
                <p v-else aria-label="Нет услуг">Нет услуг</p>
              </v-col>
              <v-col cols="12">
                <h4 style="font-size: 18px; color: #2E7D32;" aria-label="Отзывы">Отзывы</h4>
                <v-row v-if="reviews.length > 0" class="mt-4">
                  <v-col v-for="(review, i) in reviews.slice(0, 3)" :key="i" cols="12">
                    <v-card flat class="pa-3" style="border-radius: 8px; border: 1px solid #e0e0e0;">
                      <v-card-text style="font-size: 14px; color: #212529;">
                        "{{ review.content }}" — {{ review.author ? review.author.full_name : 'Аноним' }}
                      </v-card-text>
                    </v-card>
                  </v-col>
                </v-row>
                <p v-else style="font-size: 16px; color: #6C757D;" aria-label="Нет отзывов">Нет отзывов</p>
              </v-col>
              <v-col cols="12" v-if="role === 'client' && !otherId">
                <h4 style="font-size: 18px; color: #2E7D32;" aria-label="Записанные услуги">Записанные услуги</h4>
                <v-row v-if="enrolled.length > 0" class="mt-4">
                  <v-col v-for="enroll in enrolled" :key="enroll.id" cols="12">
                    <v-card flat class="pa-3">
                      <v-card-text>{{ enroll.course.title }} — Оплата: {{ enroll.payment.gross_amount }} руб.</v-card-text>
                    </v-card>
                  </v-col>
                </v-row>
                <p v-else aria-label="Нет записей">Нет записей</p>
              </v-col>
              <v-col cols="12" v-if="role === 'nutri' && !otherId">
                <h4 style="font-size: 18px; color: #2E7D32;" aria-label="Реквизиты">Реквизиты для выплат</h4>
                <v-form @submit.prevent="updateCard">
                  <v-text-field 
                    v-model="cardNumber" 
                    label="Номер карты (16 цифр)" 
                    type="text" 
                    :rules="[v => !!v || 'Обязательно', v => /^\d{16}$/.test(v) || '16 цифр']"
                    aria-label="Номер карты"
                  />
                  <v-btn type="submit" color="primary" aria-label="Обновить карту">Обновить карту</v-btn>
                </v-form>
                <p v-if="profile.balance" style="font-size: 16px; color: #28A745; margin-top: 8px;" aria-label="Баланс">Баланс: {{ profile.balance }} руб.</p>
              </v-col>
            </v-row>
          </v-card-text>
          <v-card-actions v-if="otherId && isLoggedIn && userId !== otherId" class="justify-center pa-4">
            <v-btn 
              color="#28A745" 
              variant="outlined" 
              @click="openChat" 
              v-tooltip="'Открыть чат'" 
              aria-label="Связаться"
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

const config = useRuntimeConfig()
const route = useRoute()
const router = useRouter()
const profile = ref({})
const courses = ref([])
const reviews = ref([])
const token = ref(null)
const role = ref('')
const userId = ref(null)
const otherId = ref(null)
const isLoggedIn = ref(false)
const editMode = ref(false)
const errorMessage = ref('')
const cardNumber = ref('')
const enrolled = ref([])

onMounted(async () => {
  if (process.client) {
    token.value = localStorage.getItem('token')
    role.value = localStorage.getItem('role') || ''
    userId.value = localStorage.getItem('userId') ? parseInt(localStorage.getItem('userId')) : null
    isLoggedIn.value = !!token.value
    otherId.value = route.params.id ? parseInt(route.params.id) : null
  }
  await loadProfile()
})

watch(() => route.params.id, async (newId) => {
  otherId.value = newId ? parseInt(newId) : null
  editMode.value = false
  await loadProfile()
})

const loadProfile = async () => {
  const headers = token.value ? { Authorization: `Bearer ${token.value}` } : {}
  try {
    const data = await $fetch(`${config.public.apiBase}/api/profile${otherId.value ? '/' + otherId.value : ''}`, { headers })
    profile.value = data.profile || {}
    courses.value = data.courses || []
    reviews.value = data.reviews || []
    if (isLoggedIn.value && role.value === 'client' && !otherId.value) {
      const enrollData = await $fetch(`${config.public.apiBase}/api/enrolled`, { headers })
      enrolled.value = enrollData || []
    }
  } catch (error) {
    errorMessage.value = 'Ошибка загрузки профиля: ' + (error.response?.data?.error || error.message || 'Неизвестная ошибка')
  }
}

const updateProfile = async () => {
  if (!token.value) {
    router.push('/login')
    return
  }
  const headers = { Authorization: `Bearer ${token.value}` }
  try {
    await $fetch(`${config.public.apiBase}/api/profile`, {
      method: 'PUT',
      headers,
      body: {
        full_name: profile.value.full_name,
        description: profile.value.description
      }
    })
    editMode.value = false
    errorMessage.value = 'Профиль обновлён'
  } catch (error) {
    errorMessage.value = 'Ошибка обновления профиля: ' + (error.response?.data?.error || error.message || 'Неизвестная ошибка')
  }
}

const updateCard = async () => {
  if (!cardNumber.value || !/^\d{16}$/.test(cardNumber.value)) {
    errorMessage.value = 'Введите корректный номер карты (16 цифр)'
    return
  }
  const headers = { Authorization: `Bearer ${token.value}` }
  try {
    await $fetch(`${config.public.apiBase}/api/profile/update-card`, {
      method: 'POST',
      headers,
      body: { card_number: cardNumber.value }
    })
    errorMessage.value = 'Карта обновлена'
    cardNumber.value = ''
  } catch (error) {
    errorMessage.value = 'Ошибка обновления карты: ' + (error.response?.data?.error || error.message || 'Неизвестная ошибка')
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
      body: { receiver_id: otherId.value }
    })
    router.push(`/chats?selected=${data.receiver_id}`)
  } catch (error) {
    errorMessage.value = 'Ошибка открытия чата: ' + (error.response?.data?.error || error.message || 'Неизвестная ошибка')
  }
}
</script>