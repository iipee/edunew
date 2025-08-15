<template>
  <v-container>
    <v-row justify="center">
      <v-col cols="12" sm="8" md="6">
        <v-card class="pa-6" aria-label="Форма оплаты">
          <v-card-title class="justify-center text-h4" aria-label="Оплата заказа">Оплата заказа</v-card-title>
          <v-card-text>
            <p style="font-size: 20px; color: #28A745;" aria-label="Сумма заказа">Сумма: {{ amount }} руб.</p>
            <v-btn 
              color="#28A745" 
              block 
              @click="openPaymentModal" 
              v-tooltip="'Открыть форму оплаты'" 
              aria-label="Открыть форму оплаты"
            >
              Оплатить
            </v-btn>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    <v-dialog v-model="dialogPayment" max-width="500" aria-label="Диалог оплаты">
      <v-card style="padding: 24px;">
        <v-card-title aria-label="Оплата">Оплата</v-card-title>
        <v-card-text>
          <v-form>
            <v-text-field 
              label="Номер карты" 
              v-model="cardNum" 
              :rules="[v => !!v || 'Поле обязательно']" 
              style="margin-bottom: 16px;" 
              aria-label="Номер карты" 
            />
            <v-text-field 
              label="Срок действия" 
              v-model="expiry" 
              :rules="[v => !!v || 'Поле обязательно']" 
              style="margin-bottom: 16px;" 
              aria-label="Срок действия" 
            />
            <v-text-field 
              label="CVV" 
              v-model="cvv" 
              type="password" 
              :rules="[v => !!v || 'Поле обязательно']" 
              style="margin-bottom: 16px;" 
              aria-label="CVV" 
            />
          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-btn color="#28A745" @click="submitPayment" v-tooltip="'Подтвердить оплату'" aria-label="Подтвердить оплату">Оплатить</v-btn>
          <v-btn text @click="dialogPayment = false" aria-label="Отмена">Отмена</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-snackbar v-model="snackbar" :color="snackbarColor" timeout="3000" aria-label="Уведомление">
      {{ snackbarText }}
    </v-snackbar>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useRuntimeConfig } from 'nuxt/app'

const config = useRuntimeConfig()
const route = useRoute()
const router = useRouter()
const orderId = ref(route.query.order_id ? parseInt(route.query.order_id) : null)
const amount = ref(route.query.amount ? parseFloat(route.query.amount) : 0)
const token = ref(null)
const userId = ref(null)
const dialogPayment = ref(false)
const cardNum = ref('')
const expiry = ref('')
const cvv = ref('')
const snackbar = ref(false)
const snackbarText = ref('')
const snackbarColor = ref('success')

onMounted(() => {
  if (process.client) {
    token.value = localStorage.getItem('token')
    userId.value = localStorage.getItem('userId') ? parseInt(localStorage.getItem('userId')) : null
    if (!token.value) {
      router.push('/login')
    }
  }
})

const openPaymentModal = () => {
  dialogPayment.value = true
}

const submitPayment = async () => {
  if (!cardNum.value || !expiry.value || !cvv.value) {
    snackbarText.value = 'Заполните все поля'
    snackbarColor.value = 'error'
    snackbar.value = true
    return
  }
  try {
    const headers = { Authorization: `Bearer ${token.value}` }
    const body = { 
      course_id: orderId.value, 
      user_id: userId.value, 
      amount: amount.value 
    }
    const { data, error } = await useFetch(`${config.public.apiBase}/api/payments/simulate`, { 
      method: 'POST', 
      headers, 
      body 
    })
    if (error.value) {
      throw new Error(error.value.data?.error || 'Неизвестная ошибка')
    }
    snackbarText.value = data.value.transaction_id ? `Оплата успешна! ID: ${data.value.transaction_id}` : 'Оплата успешна!'
    snackbarColor.value = 'success'
    snackbar.value = true
    setTimeout(() => router.push('/profile'), 2000)
  } catch (error) {
    snackbarText.value = 'Ошибка оплаты: ' + (error.message || 'Неизвестная ошибка')
    snackbarColor.value = 'error'
    snackbar.value = true
  } finally {
    dialogPayment.value = false
    cardNum.value = ''
    expiry.value = ''
    cvv.value = ''
  }
}
</script>

<style scoped>
.v-btn {
  border-radius: 8px;
}
</style>