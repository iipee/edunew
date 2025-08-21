<template>
  <v-container>
    <v-row justify="center">
      <v-col cols="12">
        <v-text-field
          v-model="searchQuery"
          label="Поиск курсов или услуг"
          prepend-icon="mdi-magnify"
          clearable
          :style="searchFieldStyle"
          @input="debouncedSearch"
          aria-label="Поле для поиска курсов или услуг"
        />
      </v-col>
    </v-row>
    <v-row>
      <v-col v-if="loading" v-for="n in 6" :key="n" cols="12" sm="6" md="4">
        <v-skeleton-loader type="card" aria-label="Загрузка карточки" />
      </v-col>
      <v-col v-else v-for="course in courses" :key="course.id" cols="12" sm="6" md="4">
        <v-card 
          class="pa-4" 
          height="450" 
          style="width: 320px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); background-color: #FFFFFF;"
          elevation="2"
          aria-label="Карточка курса"
        >
          <v-card-title class="text-h6" style="font-size: 20px; color: #212529;" aria-label="Название курса">
            {{ course.title }}
          </v-card-title>
          <v-card-subtitle style="font-size: 16px; color: #212529;" aria-label="Автор курса">
            Автор: 
            <NuxtLink 
              :to="`/nutri/${course.teacher.id}`" 
              style="color: #28A745; text-decoration: none;" 
              class="hover:underline" 
              aria-label="Ссылка на профиль автора"
            >
              {{ course.teacher.full_name || 'Не указан' }}
            </NuxtLink>
          </v-card-subtitle>
          <v-card-text style="min-height: 200px; padding: 16px 0;">
            <div style="margin: 16px 0;" aria-label="Услуги курса">
              <p v-for="(service, i) in course.services.slice(0, 3)" :key="i" style="font-size: 16px; color: #212529; margin: 4px 0;">
                • {{ service }}
              </p>
              <p v-if="course.services.length > 3" style="font-size: 16px; color: #212529; margin: 4px 0;">...</p>
            </div>
            <p style="font-size: 16px; line-height: 1.5; color: #6C757D; margin: 16px 0;">
              {{ course.description.slice(0, 100) }}...
            </p>
            <p 
              style="font-size: 20px; color: #28A745; font-weight: bold;" 
              aria-label="Цена курса"
            >
              {{ role === 'nutri' && course.teacher.id === userId ? course.net_price : course.gross_price }} руб.
            </p>
          </v-card-text>
          <v-card-actions>
            <v-btn 
              color="#28A745" 
              @click="enroll(course.id)" 
              v-tooltip="'Записаться на курс'" 
              aria-label="Записаться на курс"
            >
              Записаться
            </v-btn>
            <v-btn 
              color="primary" 
              @click="openChat(course.teacher.id)" 
              v-tooltip="'Чат с автором'" 
              aria-label="Чат с автором"
            >
              Чат
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
    <v-snackbar v-model="snackbar" :color="snackbarColor" timeout="3000" aria-label="Уведомление">
      {{ snackbarText }}
    </v-snackbar>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useRuntimeConfig } from 'nuxt/app'
import debounce from 'lodash/debounce'

const config = useRuntimeConfig()
const router = useRouter()
const searchQuery = ref('')
const courses = ref([])
const token = ref(null)
const role = ref('')
const userId = ref(null)
const isLoggedIn = computed(() => !!token.value)
const loading = ref(true)
const snackbar = ref(false)
const snackbarText = ref('')
const snackbarColor = ref('error')

const searchFieldStyle = computed(() => ({
  backgroundColor: '#FFFFFF',
  border: '1px solid #CED4DA',
  borderRadius: '8px',
  '--v-theme-placeholder': '#6C757D',
  '--v-theme-prepend-inner-icon': '#28A745'
}))

onMounted(() => {
  if (process.client) {
    token.value = localStorage.getItem('token')
    role.value = localStorage.getItem('role')
    userId.value = localStorage.getItem('userId') ? parseInt(localStorage.getItem('userId')) : null
  }
  search()
})

const search = async () => {
  loading.value = true
  const headers = token.value ? { Authorization: `Bearer ${token.value}` } : {}
  const query = searchQuery.value.trim()
  try {
    const data = await $fetch(`${config.public.apiBase}/api/search?q=${encodeURIComponent(query)}`, { headers })
    courses.value = data || []
  } catch (error) {
    snackbarText.value = 'Ошибка поиска: ' + (error.response?.data?.error || error.message || 'Неизвестная ошибка')
    snackbarColor.value = 'error'
    snackbar.value = true
    courses.value = []
  } finally {
    loading.value = false
  }
}

const debouncedSearch = debounce(search, 300)

watch(searchQuery, (newVal) => {
  if (newVal === '') {
    search()
  } else {
    debouncedSearch()
  }
})

const enroll = (id) => {
  router.push(`/courses/${id}`)
}

const openChat = async (teacherId) => {
  if (!isLoggedIn.value) {
    router.push('/login')
    return
  }
  try {
    const headers = { Authorization: `Bearer ${token.value}` }
    const data = await $fetch(`${config.public.apiBase}/api/start-chat`, {
      method: 'POST',
      headers,
      body: { receiver_id: teacherId }
    })
    router.push(`/chats?selected=${data.receiver_id}`)
  } catch (error) {
    snackbarText.value = 'Ошибка открытия чата: ' + (error.response?.data?.error || error.message || 'Неизвестная ошибка')
    snackbarColor.value = 'error'
    snackbar.value = true
  }
}
</script>

<style scoped>
.v-card:hover {
  box-shadow: 0 4px 8px rgba(0,0,0,0.2);
}
:deep(.v-text-field input::placeholder) {
  color: #6C757D !important;
}
:deep(.v-text-field .v-icon) {
  color: #28A745 !important;
}
.hover\:underline:hover {
  text-decoration: underline;
}
</style>