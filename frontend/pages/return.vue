<template>
  <v-container>
    <v-row justify="center">
      <v-col cols="12" sm="8" md="6">
        <v-card class="pa-6" aria-label="Результат оплаты">
          <v-card-title class="justify-center text-h4" aria-label="Результат оплаты">
            Результат оплаты
          </v-card-title> <!-- ИЗМЕНЕНО: h4 по ТЗ -->
          <v-card-text>
            <p :style="{ color: status === 'succeeded' ? '#28A745' : '#DC3545' }" aria-label="Сообщение"> <!-- ИЗМЕНЕНО: Зеленый/красный по ТЗ -->
              {{ message }}
            </p>
            <p v-if="transactionId" aria-label="ID транзакции">ID: {{ transactionId }}</p> <!-- ИЗМЕНЕНО: Показ ID по ТЗ -->
            <v-btn color="#28A745" @click="goToProfile" aria-label="Перейти в профиль">Перейти в профиль</v-btn>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useRuntimeConfig } from 'nuxt/app'

const config = useRuntimeConfig()
const route = useRoute()
const router = useRouter()
const status = ref('')
const message = ref('')
const transactionId = ref(null) // ИЗМЕНЕНО: Для показа ID

onMounted(async () => {
  const paymentId = route.query.payment_id
  if (!paymentId) {
    status.value = 'error'
    message.value = 'ID платежа не указан'
    return
  }
  try {
    const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` }
    const response = await $fetch(`${config.public.apiBase}/api/payments/return?payment_id=${paymentId}`, { headers })
    status.value = response.status
    message.value = response.message
    transactionId.value = response.transaction_id || null // ИЗМЕНЕНО: Получаем ID если есть
    if (status.value === 'succeeded') { // ИЗМЕНЕНО: Авто-редирект после 2 сек по ТЗ
      setTimeout(() => router.push('/profile'), 2000)
    }
  } catch (error) {
    status.value = 'error'
    message.value = 'Ошибка проверки статуса платежа: ' + (error.message || 'Неизвестная ошибка')
  }
})

const goToProfile = () => {
  router.push('/profile')
}
</script>

<style scoped>
.v-btn {
  border-radius: 8px;
}
</style>