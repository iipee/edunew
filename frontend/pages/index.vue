<template>
  <v-container fluid>
    <v-parallax :src="`/images/salad.jpg`" height="60vh">
      <v-row align="center" justify="center">
        <v-col class="text-center">
          <v-card class="pa-6 static-parallax-card" style="background: rgba(0, 0, 0, 0.6); backdrop-filter: blur(5px); box-shadow: 0 2px 4px rgba(0,0,0,0.2) !important; transform: none !important;">
            <v-card-text style="color: #FFFFFF !important; text-shadow: 0 2px 4px rgba(0,0,0,0.3);">
              <h1 class="text-h3 font-weight-bold" aria-label="Слоган платформы">
                Откройте персонализированное питание
              </h1>
              <p class="text-h5 mt-4" aria-label="Подзаголовок">
                Найдите идеального нутрициолога для вашего здоровья
              </p>
            </v-card-text>
          </v-card>
          <v-btn
            color="#28A745"
            large
            to="/search"
            class="mt-6"
            style="border-radius: 8px;"
            v-tooltip="'Перейти к поиску курсов'"
            aria-label="Поиск курсов"
          >
            Поиск курсов
          </v-btn>
        </v-col>
      </v-row>
    </v-parallax>
    <v-container>
      <v-row class="mt-12">
        <v-col cols="12">
          <h2 class="text-h4 text-center mb-8" style="color: #2E7D32;" aria-label="Рекомендуемые нутрициологи">
            Рекомендуемые нутрициологи
          </h2>
          <v-row>
            <v-col v-if="loading" v-for="n in 6" :key="n" cols="12" sm="6" md="4">
              <v-skeleton-loader type="card" aria-label="Загрузка карточки" />
            </v-col>
            <v-col v-else-if="recommended.length === 0" cols="12">
              <p class="text-center" style="font-size: 16px;" aria-label="Нет рекомендаций">Нет доступных нутрициологов</p>
            </v-col>
            <v-col v-else v-for="nutri in recommended" :key="nutri.id" cols="12" sm="6" md="4">
              <v-card 
                height="400" 
                class="pa-4 d-flex flex-column nutri-card" 
                style="width: 300px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);"
                elevation="2"
                aria-label="Карточка нутрициолога"
              >
                <v-img
                  :src="nutri.avatar_url || '/images/nutri-placeholder.jpg'"
                  width="150"
                  height="150"
                  class="mx-auto rounded-circle"
                  aria-label="Фото нутрициолога"
                  @error="onImageError"
                >
                  <template v-slot:placeholder>
                    <v-icon size="150" color="#CED4DA">mdi-account-circle</v-icon>
                  </template>
                </v-img>
                <v-card-title 
                  class="justify-center text-h6" 
                  style="font-size: 24px; color: #212529;" 
                  aria-label="Имя нутрициолога"
                >
                  {{ nutri.full_name || 'Не указано' }}
                </v-card-title>
                <v-card-text 
                  class="flex-grow-1" 
                  style="min-height: 100px; font-size: 16px; line-height: 1.5;" 
                  aria-label="Описание нутрициолога"
                >
                  {{ nutri.description ? nutri.description.slice(0, 100) + '...' : 'Описание не указано' }}
                </v-card-text>
                <v-card-actions class="justify-center">
                  <v-btn
                    color="#28A745"
                    :to="`/nutri/${nutri.id}`"
                    style="border-radius: 8px;"
                    v-tooltip="'Просмотреть профиль'"
                    aria-label="Просмотреть профиль"
                  >
                    Просмотреть
                  </v-btn>
                </v-card-actions>
              </v-card>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
      <v-row class="mt-12">
        <v-col cols="12">
          <h2 class="text-h4 text-center mb-8" style="color: #2E7D32;" aria-label="Отзывы">Отзывы</h2>
          <v-carousel height="300" hide-delimiter-background show-arrows-on-hover>
            <v-carousel-item v-for="review in reviews" :key="review.id" aria-label="Отзыв">
              <v-sheet height="100%" tile>
                <v-row class="fill-height" align="center" justify="center">
                  <v-col cols="8">
                    <p class="text-center text-h6" style="font-size: 18px; line-height: 1.5;" aria-label="Содержание отзыва">
                      "{{ review.content }}"
                    </p>
                    <p class="text-center mt-4" style="font-size: 16px; color: #6C757D;" aria-label="Автор отзыва">
                      - Пользователь #{{ review.author_id }}
                    </p>
                  </v-col>
                </v-row>
              </v-sheet>
            </v-carousel-item>
          </v-carousel>
        </v-col>
      </v-row>
      <v-snackbar v-model="snackbar" color="error" timeout="3000" aria-label="Уведомление об ошибке">
        {{ errorMessage }}
      </v-snackbar>
    </v-container>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRuntimeConfig } from 'nuxt/app'

const config = useRuntimeConfig()
const recommended = ref([])
const reviews = ref([])
const loading = ref(true)
const snackbar = ref(false)
const errorMessage = ref('')

onMounted(async () => {
  loading.value = true
  try {
    const nutrisData = await $fetch(`${config.public.apiBase}/api/nutris?limit=6&random=true`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
    })
    recommended.value = nutrisData || []
    if (!recommended.value.length) {
      errorMessage.value = 'Нет доступных нутрициологов'
      snackbar.value = true
    }
    const reviewData = await $fetch(`${config.public.apiBase}/api/reviews/random`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
    })
    reviews.value = reviewData || []
  } catch (error) {
    errorMessage.value = 'Ошибка загрузки данных: ' + (error.message || 'Неизвестная ошибка')
    snackbar.value = true
    if (error.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('role')
      localStorage.removeItem('userId')
    }
  } finally {
    loading.value = false
  }
})

definePageMeta({
  layout: 'default'
})

const onImageError = (event) => {
  event.target.src = `/images/nutri-placeholder.jpg`
  event.target.onerror = null
}
</script>

<style scoped>
/* Статичная рамка для параллакса с максимальной специфичностью */
.v-parallax .static-parallax-card {
  box-shadow: 0 2px 4px rgba(0,0,0,0.2) !important;
  transform: none !important;
  transition: none !important; /* Отключаем все переходы */
  pointer-events: none !important; /* Отключаем интерактивность для исключения hover */
}

/* Анимация только для интерактивных карточек нутрициологов */
.nutri-card:hover {
  box-shadow: 0 4px 8px rgba(0,0,0,0.2) !important;
  transition: box-shadow 0.3s ease !important;
}

@media (max-width: 600px) {
  .text-h3 {
    font-size: 18px !important;
  }
  .text-h5 {
    font-size: 14px !important;
  }
}
</style>