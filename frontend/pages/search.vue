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
            <p style="font-size: 24px; font-weight: bold; color: #28A745; margin: 16px 0;" aria-label="Стоимость">
              Цена: {{ course.gross_price }} руб.
            </p>
          </v-card-text>
          <v-card-actions style="display: flex; justify-content: space-between; gap: 8px;">
            <v-btn
              color="#28A745"
              :to="`/courses/${course.id}`"
              style="min-width: 100px;"
              v-tooltip="'Узнать подробности о курсе'"
              aria-label="Подробней"
            >
              Подробней
            </v-btn>
            <v-btn
              v-if="isLoggedIn"
              outlined
              color="#6C757D"
              @click="openChat(course.teacher.id)"
              style="min-width: 100px;"
              v-tooltip="'Связаться с автором'"
              aria-label="Связаться"
            >
              Связаться
            </v-btn>
            <v-btn
              v-else
              to="/login"
              style="min-width: 100px;"
              v-tooltip="'Войти для действий'"
              aria-label="Войти"
            >
              Войти
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
      <v-col cols="12" v-if="!loading && courses.length === 0">
        <p class="text-center" style="font-size: 16px; color: #6C757D;" aria-label="Нет результатов поиска">Нет результатов</p>
      </v-col>
    </v-row>
    <v-snackbar v-model="snackbar" :color="snackbarColor" timeout="3000" aria-label="Уведомление">
      {{ snackbarText }}
    </v-snackbar>
  </v-container>
</template>

<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import { debounce } from 'lodash'
import { useRuntimeConfig } from 'nuxt/app'
import { useRouter } from 'vue-router'

const config = useRuntimeConfig()
const router = useRouter()
const searchQuery = ref('')
const courses = ref([])
const token = ref(null)
const isLoggedIn = ref(false)
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
    isLoggedIn.value = !!token.value
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
    snackbarText.value = 'Ошибка поиска: ' + (error.message || 'Неизвестная ошибка')
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

const openChat = async (teacherId) => {
  console.log('Opening chat with teacherId:', teacherId)
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
    snackbarText.value = 'Ошибка открытия чата: ' + (error.message || 'Неизвестная ошибка')
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