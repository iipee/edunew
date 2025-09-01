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
                <h4 style="font-size: 18px; color: #2E7D32;" aria-label="Описание">Описание</h4>
                <p v-if="profile.description" style="font-size: 16px; line-height: 1.5;" aria-label="Описание">{{ profile.description }}</p>
                <p v-else style="font-size: 16px; color: #6C757D;" aria-label="Описание не указано">Описание не указано</p>
              </v-col>
              <v-col cols="12">
                <h4 style="font-size: 18px; color: #2E7D32;" aria-label="Услуги">Услуги</h4>
                <v-row v-if="courses.length > 0" class="mt-4">
                  <v-col v-for="course in sortedCourses" :key="course.id" cols="12" sm="6" md="4">
                    <NuxtLink :to="`/courses/${course.id}`" style="text-decoration: none; display: block;">
                      <v-card 
                        class="v-card v-theme--light v-card--density-default v-card--variant-elevated pa-4" 
                        style="border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); max-width: 366px; min-height: 100px; width: 100%; display: flex; flex-direction: column;"
                        aria-label="Карточка услуги"
                      >
                        <div class="v-row v-row--no-gutters align-center">
                          <div class="v-col">
                            <div class="v-card-title text-h6 pa-0" 
                              style="font-size: 20px; color: #212529; white-space: normal;"
                            >
                              {{ course.title }}
                            </div>
                            <div class="v-card-subtitle pa-0 mt-2" style="font-size: 14px; color: #6C757D;">
                              Услуги: {{ course.services ? course.services.join(', ') : 'Не указаны' }}
                            </div>
                            <div class="v-card-text pa-0 mt-2" style="font-size: 14px; color: #212529; line-height: 1.4;">
                              {{ course.description ? course.description.substring(0, 100) + '...' : 'Описание не указано' }}
                            </div>
                            <div class="v-card-actions pa-0 mt-3" style="display: flex; justify-content: space-between; align-items: center;">
                              <span style="font-size: 16px; color: #28A745; font-weight: bold;">
                                {{ role === 'nutri' ? course.net_price + ' руб. (чистая)' : course.gross_price + ' руб.' }}
                              </span>
                              <!-- Удалена кнопка "Подробнее", так как карточка кликабельна -->
                            </div>
                          </div>
                        </div>
                      </v-card>
                    </NuxtLink>
                  </v-col>
                </v-row>
                <p v-else style="font-size: 16px; color: #6C757D;" aria-label="Нет услуг">Нет доступных услуг</p>
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
            </v-row>
          </v-card-text>
          <v-card-actions v-if="isLoggedIn && userId !== otherId" class="justify-center pa-4">
            <v-btn 
              color="#28A745" 
              variant="outlined" 
              @click="openChat" 
              v-tooltip="'Открыть чат с нутрициологом'" 
              aria-label="Связаться с нутрициологом"
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
const errorMessage = ref('')
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
  await loadProfile()
})

const sortedCourses = computed(() => courses.value.sort((a, b) => a.title.localeCompare(b.title)))

const loadProfile = async () => {
  const headers = token.value ? { Authorization: `Bearer ${token.value}` } : {}
  try {
    const data = await $fetch(`${config.public.apiBase}/api/profile/${route.params.id}`, { headers })
    profile.value = data.profile || {}
    courses.value = data.courses || []
    reviews.value = data.reviews || []
    if (isLoggedIn.value && role.value === 'client') {
      const enrollData = await $fetch(`${config.public.apiBase}/api/enrolled`, { headers })
      enrolled.value = enrollData || []
    }
  } catch (error) {
    errorMessage.value = 'Ошибка загрузки профиля: ' + (error.response?.data?.error || error.message || 'Неизвестная ошибка')
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